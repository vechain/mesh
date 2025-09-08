package services

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"strings"

	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/vechain/thor/v2/thor"
	"github.com/vechain/thor/v2/tx"
)

// ConstructionService handles construction API endpoints
type ConstructionService struct {
	vechainClient *VeChainClient
}

// NewConstructionService creates a new construction service
func NewConstructionService(vechainClient *VeChainClient) *ConstructionService {
	return &ConstructionService{
		vechainClient: vechainClient,
	}
}

// ConstructionDerive derives an address from a public key
func (c *ConstructionService) ConstructionDerive(w http.ResponseWriter, r *http.Request) {
	var request types.ConstructionDeriveRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Extract public key from request
	if len(request.PublicKey.Bytes) == 0 {
		http.Error(w, "Public key is required", http.StatusBadRequest)
		return
	}

	// Convert public key bytes to ECDSA public key
	pubKey, err := crypto.UnmarshalPubkey(request.PublicKey.Bytes)
	if err != nil {
		http.Error(w, "Invalid public key format", http.StatusBadRequest)
		return
	}

	// Derive address from public key
	address := crypto.PubkeyToAddress(*pubKey)

	response := &types.ConstructionDeriveResponse{
		AccountIdentifier: &types.AccountIdentifier{
			Address: address.Hex(),
		},
		Metadata: map[string]any{
			"derivation_path": "m/44'/818'/0'/0/0", // VeChain BIP44 path
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ConstructionPreprocess preprocesses a transaction
func (c *ConstructionService) ConstructionPreprocess(w http.ResponseWriter, r *http.Request) {
	var request types.ConstructionPreprocessRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Extract operations and determine required public keys
	var requiredPublicKeys []*types.AccountIdentifier
	var clauses []map[string]any

	for _, op := range request.Operations {
		switch op.Type {
		case "Transfer":
			requiredPublicKeys = append(requiredPublicKeys, op.Account)

			// Extract clause information
			clause := map[string]any{
				"to":    op.Account.Address,
				"value": op.Amount.Value,
				"data":  "0x00",
			}
			clauses = append(clauses, clause)
		case "FeeDelegation":
			// For VIP191 fee delegation, we need the delegator's public key
			requiredPublicKeys = append(requiredPublicKeys, op.Account)
		}
	}

	response := &types.ConstructionPreprocessResponse{
		Options: map[string]any{
			"clauses": clauses,
		},
		RequiredPublicKeys: requiredPublicKeys,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ConstructionMetadata gets metadata for construction
func (c *ConstructionService) ConstructionMetadata(w http.ResponseWriter, r *http.Request) {
	var request types.ConstructionMetadataRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Get current block for blockRef
	bestBlock, err := c.vechainClient.GetBestBlock()
	if err != nil {
		http.Error(w, "Failed to get best block", http.StatusInternalServerError)
		return
	}

	// Get chain tag
	chainTag, err := c.vechainClient.GetChainID()
	if err != nil {
		http.Error(w, "Failed to get chain tag", http.StatusInternalServerError)
		return
	}

	// Calculate gas based on operations
	gas := int64(21000) // Base gas for simple transfer
	if clauses, ok := request.Options["clauses"].([]any); ok {
		gas = int64(len(clauses)) * 21000 // 21000 gas per clause
	}

	// Create blockRef (last 8 bytes of block ID)
	blockRef := bestBlock.ID[len(bestBlock.ID)-16:] // Remove 0x and take last 8 bytes

	response := &types.ConstructionMetadataResponse{
		Metadata: map[string]any{
			"blockRef": blockRef,
			"chainTag": chainTag,
			"gas":      gas,
		},
		SuggestedFee: []*types.Amount{
			{
				Value: fmt.Sprintf("%d", gas*1000000000), // gas * 1 Gwei
				Currency: &types.Currency{
					Symbol:   "VTHO",
					Decimals: 18,
					Metadata: map[string]any{
						"contractAddress": "0x0000000000000000000000000000456E65726779",
					},
				},
			},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ConstructionPayloads creates payloads for construction
func (c *ConstructionService) ConstructionPayloads(w http.ResponseWriter, r *http.Request) {
	var request types.ConstructionPayloadsRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Extract metadata
	metadata := request.Metadata
	blockRef := metadata["blockRef"].(string)
	chainTag := int(metadata["chainTag"].(float64))
	gas := int64(metadata["gas"].(float64))

	// Create VeChain transaction builder
	builder := tx.NewBuilder(tx.TypeLegacy)

	// Set chain tag
	builder.ChainTag(byte(chainTag))

	// Set block ref
	blockRefBytes := hexToBytes32(blockRef)
	builder.BlockRef(tx.BlockRef(blockRefBytes[:8]))

	// Set expiration (3 hours)
	builder.Expiration(720)

	// Set gas
	builder.Gas(uint64(gas))

	// Set gas price coefficient
	builder.GasPriceCoef(0)

	// Add clauses from operations
	for _, op := range request.Operations {
		if op.Type == "Transfer" {
			toAddr, err := thor.ParseAddress(op.Account.Address)
			if err != nil {
				http.Error(w, "Invalid address", http.StatusBadRequest)
				return
			}

			value, ok := new(big.Int).SetString(op.Amount.Value, 10)
			if !ok {
				http.Error(w, "Invalid amount", http.StatusBadRequest)
				return
			}

			clause := tx.NewClause(&toAddr)
			clause = clause.WithValue(value)
			builder.Clause(clause)
		} else if op.Type == "FeeDelegation" {
			// VIP191 Fee Delegation - this is handled in the signing process
			// The delegator will pay for the transaction fees
			continue
		}
	}

	// Build the transaction
	vechainTx := builder.Build()

	// Encode transaction
	var buf bytes.Buffer
	if err := vechainTx.EncodeRLP(&buf); err != nil {
		http.Error(w, "Failed to encode transaction", http.StatusInternalServerError)
		return
	}
	unsignedTx := buf.Bytes()

	// Create signing payloads
	var payloads []*types.SigningPayload

	// Get origin address for first payload
	if len(request.PublicKeys) > 0 {
		originAddr, err := crypto.UnmarshalPubkey(request.PublicKeys[0].Bytes)
		if err != nil {
			http.Error(w, "Invalid origin public key", http.StatusBadRequest)
			return
		}
		originAddress := crypto.PubkeyToAddress(*originAddr)

		// Create hash for origin signing
		hash := vechainTx.SigningHash()
		payload := &types.SigningPayload{
			AccountIdentifier: &types.AccountIdentifier{
				Address: originAddress.Hex(),
			},
			Bytes:         hash[:],
			SignatureType: types.EcdsaRecovery,
		}
		payloads = append(payloads, payload)
	}

	// Add delegator payload if VIP191
	if len(request.PublicKeys) > 1 {
		delegatorAddr, err := crypto.UnmarshalPubkey(request.PublicKeys[1].Bytes)
		if err != nil {
			http.Error(w, "Invalid delegator public key", http.StatusBadRequest)
			return
		}
		delegatorAddress := crypto.PubkeyToAddress(*delegatorAddr)

		// Create hash for delegator signing
		originAddr, _ := crypto.UnmarshalPubkey(request.PublicKeys[0].Bytes)
		originAddress := crypto.PubkeyToAddress(*originAddr)
		thorOriginAddr, _ := thor.ParseAddress(originAddress.Hex())
		hash := vechainTx.DelegatorSigningHash(thorOriginAddr)
		payload := &types.SigningPayload{
			AccountIdentifier: &types.AccountIdentifier{
				Address: delegatorAddress.Hex(),
			},
			Bytes:         hash[:],
			SignatureType: types.EcdsaRecovery,
		}
		payloads = append(payloads, payload)
	}

	response := &types.ConstructionPayloadsResponse{
		UnsignedTransaction: hex.EncodeToString(unsignedTx),
		Payloads:            payloads,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ConstructionParse parses a transaction
func (c *ConstructionService) ConstructionParse(w http.ResponseWriter, r *http.Request) {
	var request types.ConstructionParseRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Decode transaction
	txBytes, err := hex.DecodeString(request.Transaction)
	if err != nil {
		http.Error(w, "Invalid transaction hex", http.StatusBadRequest)
		return
	}

	var vechainTx tx.Transaction
	stream := rlp.NewStream(bytes.NewReader(txBytes), 0)
	if err := vechainTx.DecodeRLP(stream); err != nil {
		http.Error(w, "Failed to decode transaction", http.StatusBadRequest)
		return
	}

	// Parse operations
	var operations []*types.Operation
	var signers []*types.AccountIdentifier

	// Add origin signer
	origin, _ := vechainTx.Origin()
	signers = append(signers, &types.AccountIdentifier{
		Address: origin.String(),
	})

	// Add delegator signer if present
	delegator, err := vechainTx.Delegator()
	if err == nil && delegator != nil {
		signers = append(signers, &types.AccountIdentifier{
			Address: delegator.String(),
		})
	}

	// Parse clauses as operations
	clauses := vechainTx.Clauses()
	for i, clause := range clauses {
		to := clause.To()
		if to != nil {
			// Transfer operation
			operation := &types.Operation{
				OperationIdentifier: &types.OperationIdentifier{
					Index: int64(i),
				},
				Type: "Transfer",
				Account: &types.AccountIdentifier{
					Address: to.String(),
				},
				Amount: &types.Amount{
					Value: clause.Value().String(),
					Currency: &types.Currency{
						Symbol:   "VET",
						Decimals: 18,
					},
				},
			}
			operations = append(operations, operation)
		}
	}

	// Add FeeDelegation operation if VIP191 is used
	delegator, err = vechainTx.Delegator()
	if err == nil && delegator != nil {
		// Calculate estimated fee (simplified)
		estimatedFee := new(big.Int).SetUint64(vechainTx.Gas())
		estimatedFee.Mul(estimatedFee, big.NewInt(1000000000)) // 1 Gwei

		feeDelegationOp := &types.Operation{
			OperationIdentifier: &types.OperationIdentifier{
				Index: int64(len(operations)),
			},
			Type: "FeeDelegation",
			Account: &types.AccountIdentifier{
				Address: delegator.String(),
			},
			Amount: &types.Amount{
				Value: estimatedFee.String(),
				Currency: &types.Currency{
					Symbol:   "VTHO",
					Decimals: 18,
					Metadata: map[string]any{
						"contractAddress": "0x0000000000000000000000000000456E65726779",
					},
				},
			},
		}
		operations = append(operations, feeDelegationOp)
	}

	response := &types.ConstructionParseResponse{
		Operations:               operations,
		AccountIdentifierSigners: signers,
		Metadata: map[string]any{
			"chainTag":     vechainTx.ChainTag(),
			"blockRef":     fmt.Sprintf("0x%x", vechainTx.BlockRef()),
			"expiration":   vechainTx.Expiration(),
			"gas":          vechainTx.Gas(),
			"gasPriceCoef": vechainTx.GasPriceCoef(),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ConstructionCombine combines signed transactions
func (c *ConstructionService) ConstructionCombine(w http.ResponseWriter, r *http.Request) {
	var request types.ConstructionCombineRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Decode unsigned transaction
	txBytes, err := hex.DecodeString(request.UnsignedTransaction)
	if err != nil {
		http.Error(w, "Invalid unsigned transaction", http.StatusBadRequest)
		return
	}

	var vechainTx tx.Transaction
	stream := rlp.NewStream(bytes.NewReader(txBytes), 0)
	if err := vechainTx.DecodeRLP(stream); err != nil {
		http.Error(w, "Failed to decode unsigned transaction", http.StatusBadRequest)
		return
	}

	// Apply signatures for VIP191 Fee Delegation
	if len(request.Signatures) == 2 {
		// VIP191 Fee Delegation: origin + delegator signatures
		originSig := request.Signatures[0]
		delegatorSig := request.Signatures[1]

		// Verify origin signature
		originHash := vechainTx.SigningHash()
		originPubKey, err := crypto.SigToPub(originHash[:], originSig.Bytes)
		if err != nil {
			http.Error(w, "Invalid origin signature", http.StatusBadRequest)
			return
		}

		// Verify delegator signature
		// Get origin address from the first signature
		originAddr := crypto.PubkeyToAddress(*originPubKey)
		thorOriginAddr, _ := thor.ParseAddress(originAddr.Hex())
		delegatorHash := vechainTx.DelegatorSigningHash(thorOriginAddr)
		delegatorPubKey, err := crypto.SigToPub(delegatorHash[:], delegatorSig.Bytes)
		if err != nil {
			http.Error(w, "Invalid delegator signature", http.StatusBadRequest)
			return
		}

		// Verify addresses match
		recoveredOriginAddr := crypto.PubkeyToAddress(*originPubKey)
		recoveredDelegatorAddr := crypto.PubkeyToAddress(*delegatorPubKey)

		if !strings.EqualFold(recoveredOriginAddr.Hex(), originSig.SigningPayload.AccountIdentifier.Address) {
			http.Error(w, "Origin signature address mismatch", http.StatusBadRequest)
			return
		}

		if !strings.EqualFold(recoveredDelegatorAddr.Hex(), delegatorSig.SigningPayload.AccountIdentifier.Address) {
			http.Error(w, "Delegator signature address mismatch", http.StatusBadRequest)
			return
		}

		// Apply VIP191 signatures to transaction
		// Note: This would require proper implementation of signature application
		// For now, we'll use the transaction as-is

	} else if len(request.Signatures) == 1 {
		// Regular transaction: only origin signature
		sig := request.Signatures[0]
		hash := vechainTx.SigningHash()

		// Recover public key from signature
		pubKey, err := crypto.SigToPub(hash[:], sig.Bytes)
		if err != nil {
			http.Error(w, "Invalid signature", http.StatusBadRequest)
			return
		}

		// Verify the recovered address matches expected address
		recoveredAddr := crypto.PubkeyToAddress(*pubKey)
		expectedAddr := sig.SigningPayload.AccountIdentifier.Address
		if !strings.EqualFold(recoveredAddr.Hex(), expectedAddr) {
			http.Error(w, "Signature address mismatch", http.StatusBadRequest)
			return
		}
	} else {
		http.Error(w, "Invalid number of signatures", http.StatusBadRequest)
		return
	}

	// Encode signed transaction
	var buf bytes.Buffer
	if err := vechainTx.EncodeRLP(&buf); err != nil {
		http.Error(w, "Failed to encode signed transaction", http.StatusInternalServerError)
		return
	}
	signedTx := buf.Bytes()

	response := &types.ConstructionCombineResponse{
		SignedTransaction: hex.EncodeToString(signedTx),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ConstructionHash gets the hash of a transaction
func (c *ConstructionService) ConstructionHash(w http.ResponseWriter, r *http.Request) {
	var request types.ConstructionHashRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Decode transaction
	txBytes, err := hex.DecodeString(request.SignedTransaction)
	if err != nil {
		http.Error(w, "Invalid transaction hex", http.StatusBadRequest)
		return
	}

	var vechainTx tx.Transaction
	stream := rlp.NewStream(bytes.NewReader(txBytes), 0)
	if err := vechainTx.DecodeRLP(stream); err != nil {
		http.Error(w, "Failed to decode transaction", http.StatusBadRequest)
		return
	}

	// Get transaction ID
	txID := vechainTx.ID()

	response := map[string]any{
		"transaction_identifier": &types.TransactionIdentifier{
			Hash: txID.String(),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ConstructionSubmit submits a transaction to the network
func (c *ConstructionService) ConstructionSubmit(w http.ResponseWriter, r *http.Request) {
	var request types.ConstructionSubmitRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Decode transaction
	txBytes, err := hex.DecodeString(request.SignedTransaction)
	if err != nil {
		http.Error(w, "Invalid transaction hex", http.StatusBadRequest)
		return
	}

	// Submit transaction to VeChain network
	// Note: This would require implementing the actual submission logic
	// For now, we'll return a mock response
	response := map[string]any{
		"transaction_identifier": &types.TransactionIdentifier{
			Hash: "0x" + hex.EncodeToString(crypto.Keccak256(txBytes)),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Helper function to convert hex string to Bytes32
func hexToBytes32(hexStr string) thor.Bytes32 {
	hexStr = strings.TrimPrefix(hexStr, "0x")

	// Pad with zeros to make it 32 bytes
	for len(hexStr) < 64 {
		hexStr = "0" + hexStr
	}

	bytes, _ := hex.DecodeString(hexStr)
	var result thor.Bytes32
	copy(result[:], bytes)
	return result
}
