package operations

import (
	"fmt"
	"log"
	"math/big"
	"strconv"

	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/ethereum/go-ethereum/common/math"
	meshcommon "github.com/vechain/mesh/common"
	meshcrypto "github.com/vechain/mesh/common/crypto"
	"github.com/vechain/mesh/common/vip180"
	meshthor "github.com/vechain/mesh/thor"
	"github.com/vechain/thor/v2/api"
	"github.com/vechain/thor/v2/thor"
	"github.com/vechain/thor/v2/tx"
)

// The rationale for this is to support both api.JSONClause and api.Clause
// in the future, after amending the thorclient, we should not need this adapter

// ClauseValue interface for handling both JSONClause.Value and Clause.Value
type ClauseValue interface {
	MarshalText() ([]byte, error)
}

// ClauseData interface for handling both JSONClause and Clause
type ClauseData interface {
	GetValue() ClauseValue
	GetTo() *thor.Address
	GetData() string
}

// JSONClauseAdapter adapts api.JSONClause to ClauseData interface
type JSONClauseAdapter struct {
	Clause *api.JSONClause
}

func (j JSONClauseAdapter) GetValue() ClauseValue { return &j.Clause.Value }
func (j JSONClauseAdapter) GetTo() *thor.Address  { return j.Clause.To }
func (j JSONClauseAdapter) GetData() string       { return j.Clause.Data }

// ClauseAdapter adapts api.Clause to ClauseData interface
type ClauseAdapter struct {
	Clause *api.Clause
}

func (c ClauseAdapter) GetValue() ClauseValue { return c.Clause.Value }
func (c ClauseAdapter) GetTo() *thor.Address  { return c.Clause.To }
func (c ClauseAdapter) GetData() string       { return c.Clause.Data }

type ClauseParser struct {
	vechainClient       meshthor.VeChainClientInterface
	operationsExtractor *OperationsExtractor
	vip180Encoder       *vip180.VIP180Encoder
	bytesHandler        *meshcrypto.BytesHandler
}

func NewClauseParser(vechainClient meshthor.VeChainClientInterface, operationsExtractor *OperationsExtractor) *ClauseParser {
	return &ClauseParser{vechainClient: vechainClient, operationsExtractor: operationsExtractor, vip180Encoder: vip180.NewVIP180Encoder(), bytesHandler: meshcrypto.NewBytesHandler()}
}

// ParseTransactionOperationsFromClauseData parses operations from clause data with client for contract calls
func (e *ClauseParser) ParseTransactionOperationsFromClauseData(clauseData []ClauseData, originAddr string, delegatorAddr string, gas uint64, status *string) ([]*types.Operation, error) {
	var operations []*types.Operation
	hasValueTransfer, hasContractInteraction, hasEnergyTransfer, err := e.analyzeClauses(clauseData, gas)
	if err != nil {
		return nil, err
	}

	if !hasValueTransfer && !hasContractInteraction && !hasEnergyTransfer {
		return operations, nil
	}

	operationIndex := 0
	for clauseIndex, clause := range clauseData {
		value, err := e.getClauseValue(clause)
		if err != nil {
			return nil, err
		}

		// Try VIP180 token transfer first
		if e.isVIP180Transfer(clause, value) {
			ops, nextIndex := e.parseVIP180Transfer(clause, clauseIndex, operationIndex, originAddr, status)
			operations = append(operations, ops...)
			operationIndex = nextIndex
			continue
		}

		// Regular VET transfer
		if value.Cmp(big.NewInt(0)) > 0 {
			ops, nextIndex := e.parseVETTransfer(clause, clauseIndex, operationIndex, originAddr, value, status)
			operations = append(operations, ops...)
			operationIndex = nextIndex
		}

		// Contract interaction
		if e.hasContractInteraction(clause) {
			op := e.createContractInteractionOperation(clause, clauseIndex, operationIndex, originAddr, status)
			operations = append(operations, op)
			operationIndex++
		}
	}

	// Add energy transfer operation if needed
	if hasEnergyTransfer {
		op := e.createEnergyTransferOperation(operationIndex, originAddr, delegatorAddr, gas, status)
		operations = append(operations, op)
	}

	return operations, nil
}

// analyzeClauses analyzes clauses to determine what types of operations are present
func (e *ClauseParser) analyzeClauses(clauseData []ClauseData, gas uint64) (hasValueTransfer, hasContractInteraction, hasEnergyTransfer bool, err error) {
	hasEnergyTransfer = gas > 0

	for _, clause := range clauseData {
		value, err := e.getClauseValue(clause)
		if err != nil {
			return false, false, false, err
		}
		if value.Cmp(big.NewInt(0)) > 0 {
			hasValueTransfer = true
		}
		if e.hasContractInteraction(clause) {
			hasContractInteraction = true
		}
	}

	return hasValueTransfer, hasContractInteraction, hasEnergyTransfer, nil
}

// getClauseValue extracts the value from a clause as big.Int
func (e *ClauseParser) getClauseValue(clause ClauseData) (*big.Int, error) {
	valueBytes, err := clause.GetValue().MarshalText()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal clause value: %w", err)
	}
	value := new(big.Int)
	value.SetString(string(valueBytes), 0) // 0 means auto-detect base (hex with 0x prefix or decimal)
	return value, nil
}

// isVIP180Transfer checks if a clause represents a VIP180 token transfer
func (e *ClauseParser) isVIP180Transfer(clause ClauseData, value *big.Int) bool {
	return value.Cmp(big.NewInt(0)) == 0 &&
		len(clause.GetData()) > 0 &&
		clause.GetTo() != nil &&
		e.vip180Encoder.IsVIP180TransferCallData(clause.GetData())
}

// hasContractInteraction checks if a clause has contract interaction
func (e *ClauseParser) hasContractInteraction(clause ClauseData) bool {
	return len(clause.GetData()) > 0
}

// parseVIP180Transfer parses VIP180 token transfer operations
func (e *ClauseParser) parseVIP180Transfer(clause ClauseData, clauseIndex, operationIndex int, originAddr string, status *string) ([]*types.Operation, int) {
	transferData, err := e.vip180Encoder.DecodeVIP180TransferCallData(clause.GetData())
	if err != nil {
		return []*types.Operation{}, operationIndex
	}

	// Get token currency
	tokenCurrency, err := e.operationsExtractor.GetTokenCurrencyFromContractAddress(clause.GetTo().String(), e.vechainClient)
	if err != nil {
		log.Println("error getting token currency, assigning UNKNOWN symbol", err)
		tokenCurrency = &types.Currency{
			Symbol:   "UNKNOWN",
			Decimals: 18,
		}
	}

	networkIndex := int64(clauseIndex)
	operations := []*types.Operation{
		e.createTransferOperation(operationIndex, &networkIndex, originAddr, "-"+transferData.Value.String(), tokenCurrency, clauseIndex, status),
		e.createTransferOperation(operationIndex+1, &networkIndex, transferData.To.String(), transferData.Value.String(), tokenCurrency, clauseIndex, status),
	}

	return operations, operationIndex + 2
}

// parseVETTransfer parses VET transfer operations
func (e *ClauseParser) parseVETTransfer(clause ClauseData, clauseIndex, operationIndex int, originAddr string, value *big.Int, status *string) ([]*types.Operation, int) {
	valueStr := value.String()
	operations := []*types.Operation{
		e.createTransferOperation(operationIndex, nil, originAddr, "-"+valueStr, meshcommon.VETCurrency, clauseIndex, status),
	}

	// Add receiver operation if there's a valid recipient
	if clause.GetTo() != nil && !clause.GetTo().IsZero() {
		operations = append(operations,
			e.createTransferOperation(operationIndex+1, nil, clause.GetTo().String(), valueStr, meshcommon.VETCurrency, clauseIndex, status))
		return operations, operationIndex + 2
	}

	return operations, operationIndex + 1
}

// createTransferOperation creates a transfer operation with common fields
func (e *ClauseParser) createTransferOperation(index int, networkIndex *int64, address, amount string, currency *types.Currency, clauseIndex int, status *string) *types.Operation {
	return &types.Operation{
		OperationIdentifier: &types.OperationIdentifier{
			Index:        int64(index),
			NetworkIndex: networkIndex,
		},
		Type:     meshcommon.OperationTypeTransfer,
		Status:   status,
		Account:  &types.AccountIdentifier{Address: address},
		Amount:   &types.Amount{Value: amount, Currency: currency},
		Metadata: map[string]any{"clauseIndex": clauseIndex},
	}
}

// createContractInteractionOperation creates a contract interaction operation
func (e *ClauseParser) createContractInteractionOperation(clause ClauseData, clauseIndex, operationIndex int, originAddr string, status *string) *types.Operation {
	toAddr := ""
	if clause.GetTo() != nil {
		toAddr = clause.GetTo().String()
	}

	return &types.Operation{
		OperationIdentifier: &types.OperationIdentifier{Index: int64(operationIndex)},
		Type:                meshcommon.OperationTypeContractCall,
		Status:              status,
		Account:             &types.AccountIdentifier{Address: originAddr},
		Amount:              &types.Amount{Value: "0", Currency: meshcommon.VETCurrency},
		Metadata: map[string]any{
			"clauseIndex": clauseIndex,
			"to":          toAddr,
			"data":        "0x" + fmt.Sprintf("%x", clause.GetData()),
		},
	}
}

// createEnergyTransferOperation creates an energy transfer operation
func (e *ClauseParser) createEnergyTransferOperation(operationIndex int, originAddr string, delegatorAddr string, gas uint64, status *string) *types.Operation {
	feeType := meshcommon.OperationTypeFee
	metadata := map[string]any{"gas": strconv.FormatUint(gas, 10)}

	if delegatorAddr != "" {
		feeType = meshcommon.OperationTypeFeeDelegation
		metadata[meshcommon.DelegatorAccountMetadataKey] = delegatorAddr
	}

	return &types.Operation{
		OperationIdentifier: &types.OperationIdentifier{Index: int64(operationIndex)},
		Type:                feeType,
		Status:              status,
		Account:             &types.AccountIdentifier{Address: originAddr},
		Amount:              &types.Amount{Value: "-" + strconv.FormatUint(gas, 10), Currency: meshcommon.VTHOCurrency},
		Metadata:            metadata,
	}
}

// ParseTransactionOperationsFromJSONClauses is a helper function that parses operations from clauses
func (e *ClauseParser) ParseTransactionOperationsFromJSONClauses(clauses []*api.JSONClause, originAddr string, delegatorAddr string, gas uint64, status *string) ([]*types.Operation, error) {
	clauseData := make([]ClauseData, len(clauses))
	for i, clause := range clauses {
		clauseData[i] = JSONClauseAdapter{Clause: clause}
	}
	return e.ParseTransactionOperationsFromClauseData(clauseData, originAddr, delegatorAddr, gas, status)
}

// ParseOperationsFromAPIClauses is a helper function that parses operations from transactions.Clauses
func (e *ClauseParser) ParseOperationsFromAPIClauses(clauses api.Clauses, originAddr string, delegatorAddr string, gas uint64, status *string) ([]*types.Operation, error) {
	clauseData := make([]ClauseData, len(clauses))
	for i, clause := range clauses {
		clauseData[i] = ClauseAdapter{Clause: clause}
	}
	return e.ParseTransactionOperationsFromClauseData(clauseData, originAddr, delegatorAddr, gas, status)
}

// ParseClausesFromOptions parses clauses from map format to Thor tx.Clause objects
func (e *ClauseParser) ParseClausesFromOptions(clausesRaw any) ([]*tx.Clause, error) {
	// Check if clauses is a list
	clausesList, ok := clausesRaw.([]any)
	if !ok {
		return nil, fmt.Errorf("clauses must be an array")
	}

	// Convert map clauses to tx.Clause objects
	var txClauses []*tx.Clause
	for i, clauseRaw := range clausesList {
		clauseMap, ok := clauseRaw.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("clause at index %d must be an object", i)
		}

		// Parse "to" address
		var toAddr *thor.Address
		if toStr, ok := clauseMap["to"].(string); ok && toStr != "" {
			addr, err := thor.ParseAddress(toStr)
			if err != nil {
				return nil, fmt.Errorf("invalid 'to' address at index %d: %w", i, err)
			}
			toAddr = &addr
		}

		// Parse "value"
		value := new(big.Int)
		if valueStr, ok := clauseMap["value"].(string); ok {
			if _, ok := value.SetString(valueStr, 0); !ok {
				return nil, fmt.Errorf("invalid 'value' at index %d: %s", i, valueStr)
			}
		}

		// Parse "data"
		var data []byte
		if dataStr, ok := clauseMap["data"].(string); ok && dataStr != "" && dataStr != "0x" {
			var err error
			data, err = e.bytesHandler.DecodeHexStringWithPrefix(dataStr)
			if err != nil {
				return nil, fmt.Errorf("invalid 'data' hex string at index %d: %w", i, err)
			}
		}

		// Create clause
		clause := tx.NewClause(toAddr).WithValue(value).WithData(data)
		txClauses = append(txClauses, clause)
	}

	return txClauses, nil
}

// ParseHexOrDecimal256 parses a hex or decimal string into math.HexOrDecimal256
func (e *ClauseParser) ParseHexOrDecimal256(valueStr string) (*math.HexOrDecimal256, error) {
	var value math.HexOrDecimal256
	if err := value.UnmarshalText([]byte(valueStr)); err != nil {
		return nil, err
	}
	return &value, nil
}

// ParseAPIClause parses a single clause from a map to api.Clause
func (e *ClauseParser) ParseAPIClause(clauseMap map[string]any) (*api.Clause, error) {
	clause := &api.Clause{}

	// Parse 'to' address
	if toRaw, ok := clauseMap["to"]; ok && toRaw != nil {
		toStr, ok := toRaw.(string)
		if !ok {
			return nil, fmt.Errorf("'to' must be a string or null")
		}
		toAddr, err := thor.ParseAddress(toStr)
		if err != nil {
			return nil, fmt.Errorf("invalid 'to' address: %v", err)
		}
		clause.To = &toAddr
	}

	// Parse 'value'
	valueStr, ok := clauseMap["value"].(string)
	if !ok {
		return nil, fmt.Errorf("'value' is required and must be a string")
	}
	value, err := e.ParseHexOrDecimal256(valueStr)
	if err != nil {
		return nil, fmt.Errorf("invalid 'value': %v", err)
	}
	clause.Value = value

	// Parse 'data'
	dataStr, ok := clauseMap["data"].(string)
	if !ok {
		return nil, fmt.Errorf("'data' is required and must be a string")
	}
	clause.Data = dataStr

	return clause, nil
}
