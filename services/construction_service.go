package services

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"strings"

	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/vechain/mesh/config"
	meshmodels "github.com/vechain/mesh/models"
	meshthor "github.com/vechain/mesh/thor"
	meshutils "github.com/vechain/mesh/utils"
	"github.com/vechain/thor/v2/thor"
	"github.com/vechain/thor/v2/tx"
)

// ConstructionService handles construction API endpoints
type ConstructionService struct {
	vechainClient *meshthor.VeChainClient
	encoder       *meshutils.MeshTransactionEncoder
	config        *config.Config
}

// NewConstructionService creates a new construction service
func NewConstructionService(vechainClient *meshthor.VeChainClient, config *config.Config) *ConstructionService {
	return &ConstructionService{
		vechainClient: vechainClient,
		encoder:       meshutils.NewMeshTransactionEncoder(),
		config:        config,
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
	pubKey, err := crypto.DecompressPubkey(request.PublicKey.Bytes)
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

	meshutils.WriteJSONResponse(w, response)
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

	meshutils.WriteJSONResponse(w, response)
}

// ConstructionMetadata gets metadata for construction
func (c *ConstructionService) ConstructionMetadata(w http.ResponseWriter, r *http.Request) {
	var request types.ConstructionMetadataRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Get basic transaction info
	bestBlock, chainTag, err := c.getBasicTransactionInfo()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Determine transaction type
	transactionType := meshutils.GetStringFromOptions(request.Options, "transactionType", "dynamic")

	// Calculate gas and create blockRef
	gas := c.calculateGas(request.Options)
	blockRef := bestBlock.ID[:18]
	nonce, err := meshutils.GenerateNonce()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Build metadata based on transaction type
	metadata, gasPrice, err := c.buildMetadata(transactionType, blockRef, int64(chainTag), gas, nonce)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Calculate fee and build response
	fee := new(big.Int).Mul(big.NewInt(gas), gasPrice)
	response := &types.ConstructionMetadataResponse{
		Metadata: metadata,
		SuggestedFee: []*types.Amount{
			{
				Value:    "-" + fee.String(),
				Currency: meshmodels.VTHOCurrency,
			},
		},
	}

	meshutils.WriteJSONResponse(w, response)
}

// ConstructionPayloads creates payloads for construction
func (c *ConstructionService) ConstructionPayloads(w http.ResponseWriter, r *http.Request) {
	var request types.ConstructionPayloadsRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Get transaction origin from operations
	origins := meshutils.GetTxOrigins(request.Operations)
	txOrigin := origins[0]

	// Check fee delegation
	txDelegator := c.getFeeDelegatorAccount(request.Metadata)
	hasFeeDelegation := txDelegator != ""

	// Validate origin address matches first public key
	originAddress, err := meshutils.ComputeAddress(request.PublicKeys[0])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if originAddress != txOrigin {
		http.Error(w, fmt.Sprintf("Origin address mismatch: expected %s, got %s", txOrigin, originAddress), http.StatusBadRequest)
		return
	}

	// Validate delegator address if present
	if hasFeeDelegation {
		delegatorAddress, err := meshutils.ComputeAddress(request.PublicKeys[1])
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if delegatorAddress != txDelegator {
			http.Error(w, fmt.Sprintf("Delegator address mismatch: expected %s, got %s", txDelegator, delegatorAddress), http.StatusBadRequest)
			return
		}
	}

	// Build transaction
	vechainTx, err := c.buildTransaction(request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Create signing payloads
	payloads, err := c.createSigningPayloads(vechainTx, request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Get origin and delegator addresses
	originAddr, err := meshutils.ComputeAddress(request.PublicKeys[0])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	originBytes, _ := hex.DecodeString(originAddr[2:]) // Remove 0x prefix

	var delegatorBytes []byte
	if hasFeeDelegation {
		delegatorAddr, err := meshutils.ComputeAddress(request.PublicKeys[1])
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		delegatorBytes, _ = hex.DecodeString(delegatorAddr[2:]) // Remove 0x prefix
	}

	// Encode transaction using Rosetta schema
	unsignedTx, err := c.encoder.EncodeUnsignedTransaction(vechainTx, originBytes, delegatorBytes)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := &types.ConstructionPayloadsResponse{
		UnsignedTransaction: fmt.Sprintf("0x%s", hex.EncodeToString(unsignedTx)),
		Payloads:            payloads,
	}

	meshutils.WriteJSONResponse(w, response)
}

// ConstructionParse parses a transaction
func (c *ConstructionService) ConstructionParse(w http.ResponseWriter, r *http.Request) {
	var request types.ConstructionParseRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Decode transaction
	txBytes, err := meshutils.DecodeHexStringWithPrefix(request.Transaction)
	if err != nil {
		http.Error(w, "Invalid transaction hex", http.StatusBadRequest)
		return
	}

	var vechainTx *tx.Transaction
	var rosettaTx *meshutils.MeshTransaction

	if request.Signed {
		// For signed transactions, try to decode as Rosetta transaction first
		rosettaTx, err = c.encoder.DecodeSignedTransaction(txBytes)
		if err != nil {
			// Fallback to native Thor decoding
			var nativeTx tx.Transaction
			stream := rlp.NewStream(bytes.NewReader(txBytes), 0)
			if err := nativeTx.DecodeRLP(stream); err != nil {
				http.Error(w, "Failed to decode transaction", http.StatusBadRequest)
				return
			}
			vechainTx = &nativeTx
		} else {
			vechainTx = rosettaTx.Transaction
		}
	} else {
		// For unsigned transactions, decode as Rosetta transaction
		rosettaTx, err = c.encoder.DecodeUnsignedTransaction(txBytes)
		if err != nil {
			http.Error(w, "Failed to decode unsigned transaction", http.StatusBadRequest)
			return
		}
		vechainTx = rosettaTx.Transaction
	}

	// Parse operations
	var operations []*types.Operation
	var signers []*types.AccountIdentifier

	// Add origin signer
	var originAddr thor.Address
	var delegatorAddr *thor.Address

	if rosettaTx != nil {
		// Use Rosetta transaction fields
		originAddr = thor.BytesToAddress(rosettaTx.Origin)
		if len(rosettaTx.Delegator) > 0 {
			delegator := thor.BytesToAddress(rosettaTx.Delegator)
			delegatorAddr = &delegator
		}
	} else {
		// Use native Thor methods
		originAddr, _ = vechainTx.Origin()
		delegator, err := vechainTx.Delegator()
		if err == nil && delegator != nil {
			delegatorAddr = delegator
		}
	}

	signers = append(signers, &types.AccountIdentifier{
		Address: originAddr.String(),
	})

	// Add delegator signer if present
	if delegatorAddr != nil {
		signers = append(signers, &types.AccountIdentifier{
			Address: delegatorAddr.String(),
		})
	}

	// Parse clauses as operations
	clauses := vechainTx.Clauses()
	operationIndex := 0
	for _, clause := range clauses {
		to := clause.To()
		if to != nil {
			// Sender operation (negative amount)
			senderOp := &types.Operation{
				OperationIdentifier: &types.OperationIdentifier{
					Index: int64(operationIndex),
				},
				Type: "Transfer",
				Account: &types.AccountIdentifier{
					Address: originAddr.String(),
				},
				Amount: &types.Amount{
					Value:    "-" + clause.Value().String(),
					Currency: meshmodels.VETCurrency,
				},
			}
			operations = append(operations, senderOp)
			operationIndex++

			// Receiver operation (positive amount)
			receiverOp := &types.Operation{
				OperationIdentifier: &types.OperationIdentifier{
					Index: int64(operationIndex),
				},
				Type: "Transfer",
				Account: &types.AccountIdentifier{
					Address: to.String(),
				},
				Amount: &types.Amount{
					Value:    clause.Value().String(),
					Currency: meshmodels.VETCurrency,
				},
			}
			operations = append(operations, receiverOp)
			operationIndex++
		}
	}

	// Calculate gas price based on transaction type
	var gasPrice *big.Int
	if vechainTx.Type() == tx.TypeLegacy {
		gasPrice = big.NewInt(1000000000) // 1 Gwei for legacy
	} else {
		// For dynamic fee, use maxFeePerGas
		gasPrice = vechainTx.MaxFeePerGas()
	}

	// Calculate fee amount
	feeAmount := new(big.Int).Mul(big.NewInt(int64(vechainTx.Gas())), gasPrice)

	// Add fee operation
	if delegatorAddr != nil {
		// Fee delegation operation
		feeDelegationOp := &types.Operation{
			OperationIdentifier: &types.OperationIdentifier{
				Index: int64(operationIndex),
			},
			Type: "FeeDelegation",
			Account: &types.AccountIdentifier{
				Address: delegatorAddr.String(),
			},
			Amount: &types.Amount{
				Value:    "-" + feeAmount.String(),
				Currency: meshmodels.VTHOCurrency,
			},
		}
		operations = append(operations, feeDelegationOp)
	} else {
		// Regular fee operation
		feeOp := &types.Operation{
			OperationIdentifier: &types.OperationIdentifier{
				Index: int64(operationIndex),
			},
			Type: "Fee",
			Account: &types.AccountIdentifier{
				Address: originAddr.String(),
			},
			Amount: &types.Amount{
				Value:    "-" + feeAmount.String(),
				Currency: meshmodels.VTHOCurrency,
			},
		}
		operations = append(operations, feeOp)
	}

	// Build metadata based on transaction type
	metadata := map[string]any{
		"chainTag":   vechainTx.ChainTag(),
		"blockRef":   fmt.Sprintf("0x%x", vechainTx.BlockRef()),
		"expiration": vechainTx.Expiration(),
		"gas":        vechainTx.Gas(),
		"nonce":      fmt.Sprintf("0x%x", vechainTx.Nonce()),
	}

	if vechainTx.Type() == tx.TypeLegacy {
		metadata["gasPriceCoef"] = vechainTx.GasPriceCoef()
	} else {
		metadata["maxFeePerGas"] = vechainTx.MaxFeePerGas().String()
		metadata["maxPriorityFeePerGas"] = vechainTx.MaxPriorityFeePerGas().String()
	}

	response := &types.ConstructionParseResponse{
		Operations:               operations,
		AccountIdentifierSigners: signers,
		Metadata:                 metadata,
	}

	meshutils.WriteJSONResponse(w, response)
}

// ConstructionCombine combines signed transactions
func (c *ConstructionService) ConstructionCombine(w http.ResponseWriter, r *http.Request) {
	var request types.ConstructionCombineRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	txBytes, err := meshutils.DecodeHexStringWithPrefix(request.UnsignedTransaction)
	if err != nil {
		http.Error(w, "Invalid unsigned transaction", http.StatusBadRequest)
		return
	}

	// Decode unsigned transaction using unified method
	rosettaTx, err := c.encoder.DecodeUnsignedTransaction(txBytes)
	if err != nil {
		http.Error(w, "Failed to decode unsigned transaction", http.StatusBadRequest)
		return
	}

	// Apply signatures to Rosetta transaction
	if len(request.Signatures) == 2 {
		// VIP191 Fee Delegation with two signatures
		originSig := request.Signatures[0]
		delegatorSig := request.Signatures[1]

		// Combine signatures for VIP191
		combinedSig := append(originSig.Bytes, delegatorSig.Bytes...)
		rosettaTx.Signature = combinedSig

	} else if len(request.Signatures) == 1 {
		// Regular transaction: only origin signature
		sig := request.Signatures[0]
		rosettaTx.Signature = sig.Bytes

	} else {
		http.Error(w, "Invalid number of signatures", http.StatusBadRequest)
		return
	}

	// Encode signed Rosetta transaction
	signedTxBytes, err := c.encoder.EncodeSignedTransaction(rosettaTx)
	if err != nil {
		http.Error(w, "Failed to encode signed transaction", http.StatusInternalServerError)
		return
	}

	response := &types.ConstructionCombineResponse{
		SignedTransaction: fmt.Sprintf("0x%s", hex.EncodeToString(signedTxBytes)),
	}

	meshutils.WriteJSONResponse(w, response)
}

// ConstructionHash gets the hash of a transaction
func (c *ConstructionService) ConstructionHash(w http.ResponseWriter, r *http.Request) {
	var request types.ConstructionHashRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Decode transaction
	txBytes, err := meshutils.DecodeHexStringWithPrefix(request.SignedTransaction)
	if err != nil {
		http.Error(w, "Invalid transaction hex", http.StatusBadRequest)
		return
	}

	meshTx, err := c.encoder.DecodeSignedTransaction(txBytes)
	if err != nil {
		http.Error(w, "Failed to decode transaction", http.StatusBadRequest)
		return
	}

	response := map[string]any{
		"transaction_identifier": &types.TransactionIdentifier{
			Hash: meshTx.Transaction.ID().String(),
		},
	}

	meshutils.WriteJSONResponse(w, response)
}

// ConstructionSubmit submits a transaction to the network
func (c *ConstructionService) ConstructionSubmit(w http.ResponseWriter, r *http.Request) {
	var request types.ConstructionSubmitRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Decode transaction using our utility method
	txBytes, err := meshutils.DecodeHexStringWithPrefix(request.SignedTransaction)
	if err != nil {
		http.Error(w, "Invalid transaction hex", http.StatusBadRequest)
		return
	}

	// Decode Mesh transaction to get the native Thor transaction
	meshTx, err := c.encoder.DecodeSignedTransaction(txBytes)
	if err != nil {
		http.Error(w, "Failed to decode Mesh transaction", http.StatusBadRequest)
		return
	}

	// Build a new Thor transaction with proper signature and reserved fields
	thorTx, err := c.buildThorTransactionFromMesh(meshTx)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to build Thor transaction: %v", err), http.StatusInternalServerError)
		return
	}

	// Encode the Thor transaction to bytes
	var txBuffer bytes.Buffer
	if err := thorTx.EncodeRLP(&txBuffer); err != nil {
		http.Error(w, fmt.Sprintf("Failed to encode transaction: %v", err), http.StatusInternalServerError)
		return
	}

	// Submit the native Thor transaction to VeChain network
	txID, err := c.vechainClient.SubmitTransaction(txBuffer.Bytes())
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to submit transaction: %v", err), http.StatusInternalServerError)
		return
	}

	response := map[string]any{
		"transaction_identifier": &types.TransactionIdentifier{
			Hash: txID,
		},
	}

	meshutils.WriteJSONResponse(w, response)
}

// buildThorTransactionFromMesh builds a native Thor transaction from a Mesh transaction
func (c *ConstructionService) buildThorTransactionFromMesh(meshTx *meshutils.MeshTransaction) (*tx.Transaction, error) {
	var builder *tx.Builder
	if meshTx.Type() == tx.TypeLegacy {
		builder = tx.NewBuilder(tx.TypeLegacy)
		builder.GasPriceCoef(meshTx.GasPriceCoef())
	} else {
		builder = tx.NewBuilder(tx.TypeDynamicFee)
		builder.MaxFeePerGas(meshTx.MaxFeePerGas())
		builder.MaxPriorityFeePerGas(meshTx.MaxPriorityFeePerGas())
	}

	builder.ChainTag(meshTx.ChainTag())
	builder.BlockRef(meshTx.BlockRef())
	builder.Expiration(meshTx.Expiration())
	builder.Gas(meshTx.Gas())
	builder.Nonce(meshTx.Nonce())

	for _, clause := range meshTx.Clauses() {
		builder.Clause(clause)
	}

	if len(meshTx.Delegator) > 0 && meshTx.Delegator[0] != 0 {
		// Set reserved field for fee delegation (features: 1)
		builder.Features(1)
	}

	thorTx := builder.Build()

	if len(meshTx.Signature) > 0 {
		thorTx = thorTx.WithSignature(meshTx.Signature)
	}

	return thorTx, nil
}

// getBasicTransactionInfo gets basic transaction information from the network
func (c *ConstructionService) getBasicTransactionInfo() (*meshthor.Block, int, error) {
	bestBlock, err := c.vechainClient.GetBestBlock()
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get best block: %w", err)
	}

	chainTag, err := c.vechainClient.GetChainID()
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get chain tag: %w", err)
	}

	return bestBlock, chainTag, nil
}

// calculateGas calculates gas based on operations
func (c *ConstructionService) calculateGas(options map[string]any) int64 {
	gas := int64(20000) // Base gas

	if clauses, ok := options["clauses"].([]any); ok {
		for _, clause := range clauses {
			if clauseMap, ok := clause.(map[string]any); ok {
				if to, ok := clauseMap["to"].(string); ok {
					// VTHO contract requires more gas
					if strings.EqualFold(to, meshmodels.VTHOCurrency.Metadata["contractAddress"].(string)) {
						gas += 50000
					} else {
						gas += 10000
					}
				}
			}
		}
	}

	// Add 20% buffer
	return int64(float64(gas) * 1.2)
}

// buildMetadata builds metadata based on transaction type
func (c *ConstructionService) buildMetadata(transactionType, blockRef string, chainTag, gas int64, nonce string) (map[string]any, *big.Int, error) {
	if transactionType == "legacy" {
		return c.buildLegacyMetadata(blockRef, chainTag, gas, nonce)
	}
	return c.buildDynamicMetadata(blockRef, chainTag, gas, nonce)
}

// buildLegacyMetadata builds metadata for legacy transactions
func (c *ConstructionService) buildLegacyMetadata(blockRef string, chainTag, gas int64, nonce string) (map[string]any, *big.Int, error) {
	// Generate random gasPriceCoef (0-255)
	randomBytes := make([]byte, 1)
	if _, err := rand.Read(randomBytes); err != nil {
		return nil, nil, fmt.Errorf("failed to generate random gas price coefficient: %w", err)
	}
	gasPriceCoef := randomBytes[0]

	metadata := map[string]any{
		"transactionType": "legacy",
		"blockRef":        blockRef,
		"chainTag":        chainTag,
		"gas":             gas,
		"nonce":           nonce,
		"gasPriceCoef":    gasPriceCoef,
	}

	return metadata, c.config.GetBaseGasPrice(), nil
}

// buildDynamicMetadata builds metadata for dynamic fee transactions
func (c *ConstructionService) buildDynamicMetadata(blockRef string, chainTag, gas int64, nonce string) (map[string]any, *big.Int, error) {
	// Get dynamic gas price from network
	dynamicGasPrice, err := c.vechainClient.GetDynamicGasPrice()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get dynamic gas price: %w", err)
	}

	if dynamicGasPrice.BaseFee.Cmp(big.NewInt(0)) == 0 {
		// Case where we are building a dynamic fee transaction but the base fee is 0
		// This happens when the node is catching up with the chain block-wise
		metadata := map[string]any{
			"transactionType":      "dynamic",
			"blockRef":             blockRef,
			"chainTag":             chainTag,
			"gas":                  gas,
			"nonce":                nonce,
			"maxFeePerGas":         c.config.GetBaseGasPrice().String(),
			"maxPriorityFeePerGas": 0,
		}
		return metadata, c.config.GetBaseGasPrice(), nil
	}

	// Normal case: use actual base fee and reward
	gasPrice := new(big.Int).Add(dynamicGasPrice.BaseFee, dynamicGasPrice.Reward)
	metadata := map[string]any{
		"transactionType":      "dynamic",
		"blockRef":             blockRef,
		"chainTag":             chainTag,
		"gas":                  gas,
		"nonce":                nonce,
		"maxFeePerGas":         gasPrice.String(),
		"maxPriorityFeePerGas": dynamicGasPrice.Reward.String(),
	}

	return metadata, gasPrice, nil
}

// buildTransaction builds a VeChain transaction from the request
func (c *ConstructionService) buildTransaction(request types.ConstructionPayloadsRequest) (*tx.Transaction, error) {
	// Extract metadata
	metadata := request.Metadata
	blockRef := metadata["blockRef"].(string)
	chainTag := int(metadata["chainTag"].(float64))
	gas := int64(metadata["gas"].(float64))
	transactionType := metadata["transactionType"].(string)
	nonce := metadata["nonce"].(string)

	// Parse nonce to uint64
	nonceStr := strings.TrimPrefix(nonce, "0x")
	nonceValue := new(big.Int)
	nonceValue, ok := nonceValue.SetString(nonceStr, 16)
	if !ok {
		return nil, fmt.Errorf("invalid nonce: %s", nonceStr)
	}

	// Create transaction builder
	builder, err := c.createTransactionBuilder(transactionType, metadata)
	if err != nil {
		return nil, err
	}

	// Set common fields
	builder.ChainTag(byte(chainTag))

	blockRefBytes, err := hex.DecodeString(blockRef[2:])
	if err != nil {
		return nil, fmt.Errorf("invalid blockRef: %w", err)
	}

	builder.BlockRef(tx.BlockRef(blockRefBytes))

	// Set expiration from configuration
	expiration := c.config.GetExpiration()
	builder.Expiration(expiration)

	builder.Gas(uint64(gas))
	builder.Nonce(nonceValue.Uint64())

	// Add clauses from operations
	if err := c.addClausesToBuilder(builder, request.Operations); err != nil {
		return nil, err
	}

	return builder.Build(), nil
}

// createTransactionBuilder creates a transaction builder based on type
func (c *ConstructionService) createTransactionBuilder(transactionType string, metadata map[string]any) (*tx.Builder, error) {
	if transactionType == "legacy" {
		builder := tx.NewBuilder(tx.TypeLegacy)
		gasPriceCoef := int(metadata["gasPriceCoef"].(float64))
		builder.GasPriceCoef(uint8(gasPriceCoef))
		return builder, nil
	}

	// Dynamic fee transaction
	builder := tx.NewBuilder(tx.TypeDynamicFee)
	maxFeePerGas := metadata["maxFeePerGas"].(string)
	maxPriorityFeePerGas := metadata["maxPriorityFeePerGas"].(string)

	maxFee := new(big.Int)
	maxFee, ok := maxFee.SetString(maxFeePerGas, 10)
	if !ok {
		return nil, fmt.Errorf("invalid maxFeePerGas: %s", maxFeePerGas)
	}
	maxPriority := new(big.Int)
	maxPriority, ok = maxPriority.SetString(maxPriorityFeePerGas, 10)
	if !ok {
		return nil, fmt.Errorf("invalid maxPriorityFeePerGas: %s", maxPriorityFeePerGas)
	}

	builder.MaxFeePerGas(maxFee)
	builder.MaxPriorityFeePerGas(maxPriority)
	return builder, nil
}

// addClausesToBuilder adds clauses to the transaction builder
func (c *ConstructionService) addClausesToBuilder(builder *tx.Builder, operations []*types.Operation) error {
	for _, op := range operations {
		if op.Type == meshutils.OperationTypeTransfer {
			// Only process Transfer operations with positive values (recipients)
			value := new(big.Int)
			value, ok := value.SetString(op.Amount.Value, 10)
			if !ok {
				return fmt.Errorf("invalid amount: %s", op.Amount.Value)
			}

			// Skip negative values (these represent the sender/origin, not a clause)
			if value.Sign() < 0 {
				continue
			}

			toAddr, err := thor.ParseAddress(op.Account.Address)
			if err != nil {
				return fmt.Errorf("invalid address: %w", err)
			}

			clause := tx.NewClause(&toAddr)
			clause = clause.WithValue(value)
			builder.Clause(clause)
		}
		// FeeDelegation operations are handled in the signing process
	}
	return nil
}

// createSigningPayloads creates signing payloads for the transaction
func (c *ConstructionService) createSigningPayloads(vechainTx *tx.Transaction, request types.ConstructionPayloadsRequest) ([]*types.SigningPayload, error) {
	var payloads []*types.SigningPayload

	// Check for fee delegation
	hasFeeDelegation := c.hasFeeDelegation(request.Operations)

	// Get origin address for first payload
	if len(request.PublicKeys) > 0 {
		originPayload, err := c.createOriginPayload(vechainTx, request.PublicKeys[0])
		if err != nil {
			return nil, err
		}
		payloads = append(payloads, originPayload)
	}

	// Add delegator payload if VIP191
	if hasFeeDelegation && len(request.PublicKeys) > 1 {
		delegatorPayload, err := c.createDelegatorPayload(vechainTx, request.PublicKeys)
		if err != nil {
			return nil, err
		}
		payloads = append(payloads, delegatorPayload)
	}

	return payloads, nil
}

// hasFeeDelegation checks if there are fee delegation operations
func (c *ConstructionService) hasFeeDelegation(operations []*types.Operation) bool {
	for _, op := range operations {
		if op.Type == meshutils.OperationTypeFeeDelegation {
			return true
		}
	}
	return false
}

// createOriginPayload creates the origin signing payload
func (c *ConstructionService) createOriginPayload(vechainTx *tx.Transaction, publicKey *types.PublicKey) (*types.SigningPayload, error) {
	originPubKey, err := crypto.DecompressPubkey(publicKey.Bytes)
	if err != nil {
		return nil, fmt.Errorf("invalid origin public key: %w", err)
	}
	originAddress := crypto.PubkeyToAddress(*originPubKey)

	hash := vechainTx.SigningHash()
	return &types.SigningPayload{
		AccountIdentifier: &types.AccountIdentifier{
			Address: originAddress.Hex(),
		},
		Bytes:         hash[:],
		SignatureType: types.EcdsaRecovery,
	}, nil
}

// createDelegatorPayload creates the delegator signing payload
func (c *ConstructionService) createDelegatorPayload(vechainTx *tx.Transaction, publicKeys []*types.PublicKey) (*types.SigningPayload, error) {
	delegatorAddr, err := crypto.DecompressPubkey(publicKeys[1].Bytes)
	if err != nil {
		return nil, fmt.Errorf("invalid delegator public key: %w", err)
	}
	delegatorAddress := crypto.PubkeyToAddress(*delegatorAddr)

	// Create hash for delegator signing
	originAddr, _ := crypto.DecompressPubkey(publicKeys[0].Bytes)
	originAddress := crypto.PubkeyToAddress(*originAddr)
	thorOriginAddr, _ := thor.ParseAddress(originAddress.Hex())
	hash := vechainTx.DelegatorSigningHash(thorOriginAddr)

	return &types.SigningPayload{
		AccountIdentifier: &types.AccountIdentifier{
			Address: delegatorAddress.Hex(),
		},
		Bytes:         hash[:],
		SignatureType: types.EcdsaRecovery,
	}, nil
}

// getFeeDelegatorAccount extracts fee delegator account from metadata
func (c *ConstructionService) getFeeDelegatorAccount(metadata map[string]any) string {
	if delegator, ok := metadata["fee_delegator_account"].(string); ok {
		return strings.ToLower(delegator)
	}
	return ""
}
