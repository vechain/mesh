package tx

import (
	"fmt"

	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/pkg/errors"
	meshcommon "github.com/vechain/mesh/common"
	meshoperations "github.com/vechain/mesh/common/operations"
	meshthor "github.com/vechain/mesh/thor"
	"github.com/vechain/thor/v2/api"
	"github.com/vechain/thor/v2/thor"
	thorTx "github.com/vechain/thor/v2/tx"
)

// MeshTransaction represents a transaction with Mesh-specific fields
type MeshTransaction struct {
	*thorTx.Transaction
	// Origin and delegator are required as separate fields because they get encoded when there is no signature
	Origin    []byte
	Delegator []byte
}

// MeshTransactionEncoder handles encoding and decoding of Mesh transactions
type MeshTransactionEncoder struct {
	vechainClient meshthor.VeChainClientInterface
	clauseParser  *meshoperations.ClauseParser
}

// NewMeshTransactionEncoder creates a new Mesh transaction encoder
func NewMeshTransactionEncoder(vechainClient meshthor.VeChainClientInterface) *MeshTransactionEncoder {
	return &MeshTransactionEncoder{
		vechainClient: vechainClient,
		clauseParser:  meshoperations.NewClauseParser(vechainClient, meshoperations.NewOperationsExtractor()),
	}
}

// EncodeTransaction encodes a transaction using Mesh RLP schema
func (e *MeshTransactionEncoder) EncodeTransaction(meshTx *MeshTransaction) ([]byte, error) {
	// Use native Thor encoding and add Mesh-specific fields
	thorBytes, err := meshTx.MarshalBinary()
	if err != nil {
		return nil, fmt.Errorf("failed to encode Thor transaction: %w", err)
	}

	if len(meshTx.Signature()) > 0 {
		return thorBytes, nil
	}

	// Create Mesh structure if not signed: [thorTransaction, origin, delegator]
	meshTxRLP := []any{
		thorBytes,
		meshTx.Origin,
		meshTx.Delegator,
	}

	return rlp.EncodeToBytes(meshTxRLP)
}

// DecodeUnsignedTransaction decodes an unsigned transaction from Mesh RLP format
func (e *MeshTransactionEncoder) DecodeUnsignedTransaction(data []byte) (*MeshTransaction, error) {
	// For unsigned transactions, decode as RLP list: [thorTransaction, origin, delegator]
	var fields []any
	if err := rlp.DecodeBytes(data, &fields); err != nil {
		return nil, err
	}

	if len(fields) != 3 {
		return nil, fmt.Errorf("invalid Mesh transaction: expected 3 fields, got %d", len(fields))
	}

	// Decode Thor transaction from bytes
	thorBytes := fields[0].([]byte)
	var thorTx thorTx.Transaction
	if err := thorTx.UnmarshalBinary(thorBytes); err != nil {
		return nil, fmt.Errorf("failed to decode Thor transaction: %w", err)
	}

	origin := fields[1].([]byte)
	delegator := fields[2].([]byte)

	return &MeshTransaction{
		Transaction: &thorTx,
		Origin:      origin,
		Delegator:   delegator,
	}, nil
}

// DecodeSignedTransaction decodes a signed transaction from Mesh RLP format
func (e *MeshTransactionEncoder) DecodeSignedTransaction(data []byte) (*MeshTransaction, error) {
	// For signed transactions, decode directly as Thor transaction
	var thorTx thorTx.Transaction
	if err := thorTx.UnmarshalBinary(data); err != nil {
		return nil, fmt.Errorf("failed to decode Thor transaction: %w", err)
	}

	originAddr, err := thorTx.Origin()
	if err != nil {
		return nil, fmt.Errorf("failed to get origin: %w", err)
	}
	delegatorAddr, err := thorTx.Delegator()
	if err != nil {
		return nil, fmt.Errorf("failed to get delegator: %w", err)
	}

	var delegatorBytes []byte
	if delegatorAddr != nil {
		delegatorBytes = delegatorAddr.Bytes()
	}

	return &MeshTransaction{
		Transaction: &thorTx,
		Origin:      originAddr.Bytes(),
		Delegator:   delegatorBytes,
	}, nil
}

// ParseTransactionOperationsFromAPI parses operations directly from api.JSONEmbeddedTx
func (e *MeshTransactionEncoder) ParseTransactionOperationsFromAPI(tx *api.JSONEmbeddedTx) ([]*types.Operation, error) {
	status := meshcommon.OperationStatusSucceeded
	if tx.Reverted {
		status = meshcommon.OperationStatusReverted
	}

	var delegatorAddr string
	if tx.Delegator != nil && !tx.Delegator.IsZero() {
		delegatorAddr = tx.Delegator.String()
	}

	return e.clauseParser.ParseTransactionOperationsFromJSONClauses(tx.Clauses, tx.Origin.String(), delegatorAddr, tx.Gas, &status)
}

// ParseTransactionFromBytes parses a transaction from bytes and returns operations and signers
func (e *MeshTransactionEncoder) ParseTransactionFromBytes(txBytes []byte, signed bool) (*MeshTransaction, []*types.Operation, []*types.AccountIdentifier, error) {
	var meshTx *MeshTransaction
	var err error

	if signed {
		meshTx, err = e.DecodeSignedTransaction(txBytes)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("failed to decode signed transaction: %w", err)
		}
	} else {
		meshTx, err = e.DecodeUnsignedTransaction(txBytes)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("failed to decode unsigned transaction: %w", err)
		}
	}

	// Parse operations and signers
	operations, signers, err := e.parseTransactionSignersAndOperations(meshTx, signed)
	if err != nil {
		return nil, nil, nil, errors.Wrap(err, "failed to parse transaction signers and operations")
	}

	return meshTx, operations, signers, nil
}

// parseTransactionSignersAndOperations parses signers and operations from a transaction
func (e *MeshTransactionEncoder) parseTransactionSignersAndOperations(meshTx *MeshTransaction, signed bool) ([]*types.Operation, []*types.AccountIdentifier, error) {
	originAddr := thor.BytesToAddress(meshTx.Origin).String()
	var delegatorAddr string
	if len(meshTx.Delegator) > 0 {
		delegatorAddr = thor.BytesToAddress(meshTx.Delegator).String()
	}

	var signers []*types.AccountIdentifier
	if signed {
		signers = []*types.AccountIdentifier{
			{Address: originAddr},
		}

		// Add delegator signer if present
		if delegatorAddr != "" {
			signers = append(signers, &types.AccountIdentifier{
				Address: delegatorAddr,
			})
		}
	}

	clauses := meshTx.Clauses()
	clauseData := make([]meshoperations.ClauseData, len(clauses))
	for i, clause := range clauses {
		// Convert tx.Clause to api.Clause for the adapter
		dataStr := ""
		if len(clause.Data()) > 0 {
			dataStr = fmt.Sprintf("0x%x", clause.Data())
		}
		apiClause := &api.Clause{
			To:    clause.To(),
			Value: (*math.HexOrDecimal256)(clause.Value()),
			Data:  dataStr,
		}
		clauseData[i] = meshoperations.ClauseAdapter{Clause: apiClause}
	}

	operations, err := e.clauseParser.ParseTransactionOperationsFromClauseData(clauseData, originAddr, delegatorAddr, uint64(meshTx.Gas()), nil)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to parse operations")
	}

	return operations, signers, nil
}
