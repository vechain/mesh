package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"

	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/ethereum/go-ethereum/crypto"
	meshcommon "github.com/vechain/mesh/common"
	meshcrypto "github.com/vechain/mesh/common/crypto"
	meshoperations "github.com/vechain/mesh/common/operations"
	meshtx "github.com/vechain/mesh/common/tx"
	"github.com/vechain/mesh/common/vip180"
	"github.com/vechain/mesh/config"
	meshthor "github.com/vechain/mesh/thor"
	"github.com/vechain/thor/v2/api"
	"github.com/vechain/thor/v2/thor"
	"github.com/vechain/thor/v2/tx"
)

// ConstructionService handles construction API endpoints
type ConstructionService struct {
	vechainClient       meshthor.VeChainClientInterface
	encoder             *meshtx.MeshTransactionEncoder
	builder             *meshtx.TransactionBuilder
	config              *config.Config
	bytesHandler        *meshcrypto.BytesHandler
	operationsExtractor *meshoperations.OperationsExtractor
	vip180Encoder       *vip180.VIP180Encoder
	clauseParser        *meshoperations.ClauseParser
}

// NewConstructionService creates a new construction service
func NewConstructionService(vechainClient meshthor.VeChainClientInterface, config *config.Config) *ConstructionService {
	operationsExtractor := meshoperations.NewOperationsExtractor()
	return &ConstructionService{
		vechainClient:       vechainClient,
		encoder:             meshtx.NewMeshTransactionEncoder(vechainClient),
		builder:             meshtx.NewTransactionBuilder(),
		config:              config,
		bytesHandler:        meshcrypto.NewBytesHandler(),
		operationsExtractor: operationsExtractor,
		vip180Encoder:       vip180.NewVIP180Encoder(),
		clauseParser:        meshoperations.NewClauseParser(vechainClient, operationsExtractor),
	}
}

// ConstructionDerive derives an address from a public key
func (c *ConstructionService) ConstructionDerive(
	ctx context.Context,
	req *types.ConstructionDeriveRequest,
) (*types.ConstructionDeriveResponse, *types.Error) {
	// Extract public key from request
	if len(req.PublicKey.Bytes) == 0 {
		return nil, meshcommon.GetError(meshcommon.ErrPublicKeyRequired)
	}

	// Derive address from public key
	address, err := c.bytesHandler.ComputeAddress(req.PublicKey)
	if err != nil {
		return nil, meshcommon.GetErrorWithMetadata(meshcommon.ErrInvalidPublicKeyFormat, map[string]any{
			"error": err.Error(),
		})
	}

	return &types.ConstructionDeriveResponse{
		AccountIdentifier: &types.AccountIdentifier{
			Address: address,
		},
		Metadata: map[string]any{
			"derivation_path": "m/44'/818'/0'/0/0", // VeChain BIP44 path
		},
	}, nil
}

// ConstructionPreprocess preprocesses a transaction
func (c *ConstructionService) ConstructionPreprocess(
	ctx context.Context,
	req *types.ConstructionPreprocessRequest,
) (*types.ConstructionPreprocessResponse, *types.Error) {
	// Get transaction origins
	origins := c.operationsExtractor.GetTxOrigins(req.Operations)
	if len(origins) > 1 {
		return nil, meshcommon.GetError(meshcommon.ErrTransactionMultipleOrigins)
	} else if len(origins) == 0 {
		return nil, meshcommon.GetError(meshcommon.ErrTransactionOriginNotExist)
	}

	// Get fee delegator from metadata
	delegator := c.operationsExtractor.GetFeeDelegatorAccount(req.Metadata)

	// Get VET and token operations
	vetOpers := c.operationsExtractor.GetVETOperations(req.Operations)
	tokensOpers := c.operationsExtractor.GetTokensOperations(req.Operations)

	// Validate operations
	if len(vetOpers) == 0 && len(tokensOpers) == 0 {
		return nil, meshcommon.GetError(meshcommon.ErrNoTransferOperation)
	}

	// Build clauses
	var clauses []map[string]any

	// Add VET transfer clauses
	for _, op := range vetOpers {
		clauses = append(clauses, map[string]any{
			"to":    op["to"],
			"value": op["value"],
			"data":  "0x",
		})
	}

	// Add VIP180 token transfer clauses
	for _, op := range tokensOpers {
		// Encode VIP180 transfer call data
		transferData, err := c.vip180Encoder.EncodeVIP180TransferCallData(op["to"], op["value"])
		if err != nil {
			return nil, meshcommon.GetErrorWithMetadata(meshcommon.ErrInternalServerError, map[string]any{
				"error": err.Error(),
			})
		}

		clauses = append(clauses, map[string]any{
			"to":    op["token"],
			"value": "0",
			"data":  transferData,
		})
	}

	// Build response
	response := &types.ConstructionPreprocessResponse{
		Options: map[string]any{
			"clauses": clauses,
		},
		RequiredPublicKeys: []*types.AccountIdentifier{
			{Address: origins[0]},
		},
	}

	// Add delegator to required public keys if present
	if delegator != "" && delegator != origins[0] {
		response.RequiredPublicKeys = append(response.RequiredPublicKeys, &types.AccountIdentifier{
			Address: delegator,
		})
	}

	return response, nil
}

// ConstructionMetadata gets metadata for construction
func (c *ConstructionService) ConstructionMetadata(
	ctx context.Context,
	req *types.ConstructionMetadataRequest,
) (*types.ConstructionMetadataResponse, *types.Error) {
	// Get basic transaction info
	bestBlock, chainTag, err := c.getBasicTransactionInfo()
	if err != nil {
		return nil, meshcommon.GetErrorWithMetadata(meshcommon.ErrGettingBlockchainMetadata, map[string]any{
			"error": err.Error(),
		})
	}

	// Determine transaction type
	transactionType := c.operationsExtractor.GetStringFromOptions(req.Options, "transactionType")

	// Calculate gas and create blockRef
	gas, err := c.calculateGas(req.Options)
	if err != nil {
		return nil, meshcommon.GetErrorWithMetadata(meshcommon.ErrGettingBlockchainMetadata, map[string]any{
			"error": err.Error(),
		})
	}
	blockRef := bestBlock.ID[:8]
	nonce, err := c.bytesHandler.GenerateNonce()
	if err != nil {
		return nil, meshcommon.GetErrorWithMetadata(meshcommon.ErrGettingBlockchainMetadata, map[string]any{
			"error": err.Error(),
		})
	}

	// Build metadata based on transaction type
	metadata, gasPrice, err := c.buildMetadata(transactionType, fmt.Sprintf("0x%x", blockRef), uint64(chainTag), gas, nonce)
	if err != nil {
		return nil, meshcommon.GetErrorWithMetadata(meshcommon.ErrGettingBlockchainMetadata, map[string]any{
			"error": err.Error(),
		})
	}

	// Calculate fee and build response
	fee := new(big.Int).Mul(big.NewInt(int64(gas)), gasPrice)
	return &types.ConstructionMetadataResponse{
		Metadata: metadata,
		SuggestedFee: []*types.Amount{
			{
				Value:    "-" + fee.String(),
				Currency: meshcommon.VTHOCurrency,
			},
		},
	}, nil
}

// ConstructionPayloads creates payloads for construction
func (c *ConstructionService) ConstructionPayloads(
	ctx context.Context,
	req *types.ConstructionPayloadsRequest,
) (*types.ConstructionPayloadsResponse, *types.Error) {
	// Get transaction origin from operations
	origins := c.operationsExtractor.GetTxOrigins(req.Operations)
	txOrigin := origins[0]

	// Check fee delegation
	txDelegator := c.operationsExtractor.GetFeeDelegatorAccount(req.Metadata)
	hasFeeDelegation := txDelegator != ""

	// Validate origin address matches first public key
	originAddress, err := c.bytesHandler.ComputeAddress(req.PublicKeys[0])
	if err != nil {
		return nil, meshcommon.GetErrorWithMetadata(meshcommon.ErrInvalidPublicKeyFormat, map[string]any{
			"error": err.Error(),
		})
	}
	if originAddress != txOrigin {
		return nil, meshcommon.GetErrorWithMetadata(meshcommon.ErrOriginAddressMismatch, map[string]any{
			"expected": txOrigin,
			"got":      originAddress,
		})
	}

	// Validate delegator address if present
	if hasFeeDelegation {
		delegatorAddress, err := c.bytesHandler.ComputeAddress(req.PublicKeys[1])
		if err != nil {
			return nil, meshcommon.GetErrorWithMetadata(meshcommon.ErrInvalidPublicKeyFormat, map[string]any{
				"error": err.Error(),
			})
		}
		if delegatorAddress != txDelegator {
			return nil, meshcommon.GetErrorWithMetadata(meshcommon.ErrDelegatorAddressMismatch, map[string]any{
				"expected": txDelegator,
				"got":      delegatorAddress,
			})
		}
	}

	// Build transaction
	vechainTx, err := c.builder.BuildTransactionFromRequest(*req, c.config)
	if err != nil {
		return nil, meshcommon.GetErrorWithMetadata(meshcommon.ErrInvalidRequestParameters, map[string]any{
			"error": err.Error(),
		})
	}

	// Create signing payloads
	payloads, err := c.createSigningPayloads(vechainTx, *req)
	if err != nil {
		return nil, meshcommon.GetErrorWithMetadata(meshcommon.ErrInvalidRequestParameters, map[string]any{
			"error": err.Error(),
		})
	}

	// Get origin and delegator addresses
	originAddr, err := c.bytesHandler.ComputeAddress(req.PublicKeys[0])
	if err != nil {
		return nil, meshcommon.GetErrorWithMetadata(meshcommon.ErrInvalidPublicKeyFormat, map[string]any{
			"error": err.Error(),
		})
	}
	originBytes, err := c.bytesHandler.DecodeHexStringWithPrefix(originAddr)
	if err != nil {
		return nil, meshcommon.GetErrorWithMetadata(meshcommon.ErrInvalidPublicKeyFormat, map[string]any{
			"error":      err.Error(),
			"originAddr": originAddr,
		})
	}

	var delegatorBytes []byte
	if hasFeeDelegation {
		delegatorAddr, err := c.bytesHandler.ComputeAddress(req.PublicKeys[1])
		if err != nil {
			return nil, meshcommon.GetErrorWithMetadata(meshcommon.ErrInvalidPublicKeyFormat, map[string]any{
				"error":         err.Error(),
				"delegatorAddr": delegatorAddr,
			})
		}
		delegatorBytes, err = c.bytesHandler.DecodeHexStringWithPrefix(delegatorAddr)
		if err != nil {
			return nil, meshcommon.GetErrorWithMetadata(meshcommon.ErrInvalidPublicKeyFormat, map[string]any{
				"error": err.Error(),
			})
		}
	}

	// Encode transaction using Mesh schema
	unsignedTx, err := c.encoder.EncodeTransaction(&meshtx.MeshTransaction{
		Transaction: vechainTx,
		Origin:      originBytes,
		Delegator:   delegatorBytes,
	})
	if err != nil {
		return nil, meshcommon.GetErrorWithMetadata(meshcommon.ErrFailedToEncodeTransaction, map[string]any{
			"error": err.Error(),
		})
	}

	return &types.ConstructionPayloadsResponse{
		UnsignedTransaction: fmt.Sprintf("0x%s", hex.EncodeToString(unsignedTx)),
		Payloads:            payloads,
	}, nil
}

// ConstructionParse parses a transaction
func (c *ConstructionService) ConstructionParse(
	ctx context.Context,
	req *types.ConstructionParseRequest,
) (*types.ConstructionParseResponse, *types.Error) {
	// Decode transaction
	txBytes, err := c.bytesHandler.DecodeHexStringWithPrefix(req.Transaction)
	if err != nil {
		return nil, meshcommon.GetError(meshcommon.ErrInvalidTransactionHex)
	}

	// Use common parsing function
	meshTx, operations, signers, err := c.encoder.ParseTransactionFromBytes(txBytes, req.Signed)
	if err != nil {
		if req.Signed {
			return nil, meshcommon.GetError(meshcommon.ErrFailedToDecodeTransaction)
		}
		return nil, meshcommon.GetError(meshcommon.ErrFailedToDecodeUnsignedTransaction)
	}

	// Calculate gas price based on transaction type
	var gasPrice *big.Int
	if meshTx.Type() == tx.TypeLegacy {
		gasPrice = c.config.GetBaseGasPrice()
	} else {
		dynamicGasPrice, err := c.vechainClient.GetDynamicGasPrice()
		if err != nil {
			gasPrice = c.config.GetBaseGasPrice()
		} else {
			gasPrice = new(big.Int).Add(dynamicGasPrice.BaseFee, dynamicGasPrice.Reward)
		}
	}

	// Calculate fee amount
	feeAmount := new(big.Int).Mul(big.NewInt(int64(meshTx.Gas())), gasPrice)

	// Add fee operation
	delegatorAddr := thor.BytesToAddress(meshTx.Delegator)
	if len(meshTx.Delegator) > 0 && !delegatorAddr.IsZero() {
		// Fee delegation operation
		feeDelegationOp := &types.Operation{
			OperationIdentifier: &types.OperationIdentifier{
				Index: int64(len(operations)),
			},
			Type: meshcommon.OperationTypeFeeDelegation,
			Account: &types.AccountIdentifier{
				Address: delegatorAddr.String(),
			},
			Amount: &types.Amount{
				Value:    "-" + feeAmount.String(),
				Currency: meshcommon.VTHOCurrency,
			},
		}
		operations = append(operations, feeDelegationOp)
	} else {
		// Regular fee operation
		feeOp := &types.Operation{
			OperationIdentifier: &types.OperationIdentifier{
				Index: int64(len(operations)),
			},
			Type: meshcommon.OperationTypeFee,
			Account: &types.AccountIdentifier{
				Address: thor.BytesToAddress(meshTx.Origin).String(),
			},
			Amount: &types.Amount{
				Value:    "-" + feeAmount.String(),
				Currency: meshcommon.VTHOCurrency,
			},
		}
		operations = append(operations, feeOp)
	}

	// Build metadata based on transaction type
	metadata := map[string]any{
		"chainTag":   meshTx.ChainTag(),
		"blockRef":   fmt.Sprintf("0x%x", meshTx.BlockRef()),
		"expiration": meshTx.Expiration(),
		"gas":        meshTx.Gas(),
		"nonce":      fmt.Sprintf("0x%x", meshTx.Nonce()),
	}

	if meshTx.Type() == tx.TypeLegacy {
		metadata["gasPriceCoef"] = meshTx.GasPriceCoef()
	} else {
		metadata["maxFeePerGas"] = meshTx.MaxFeePerGas().String()
		metadata["maxPriorityFeePerGas"] = meshTx.MaxPriorityFeePerGas().String()
	}

	response := &types.ConstructionParseResponse{
		Operations: operations,
		Metadata:   metadata,
	}

	// Only include signers for signed transactions
	if req.Signed {
		response.AccountIdentifierSigners = signers
	}

	return response, nil
}

// ConstructionCombine combines signed transactions
func (c *ConstructionService) ConstructionCombine(
	ctx context.Context,
	req *types.ConstructionCombineRequest,
) (*types.ConstructionCombineResponse, *types.Error) {
	txBytes, err := c.bytesHandler.DecodeHexStringWithPrefix(req.UnsignedTransaction)
	if err != nil {
		return nil, meshcommon.GetError(meshcommon.ErrInvalidUnsignedTransactionParameter)
	}

	// Decode unsigned transaction using unified method
	meshTx, err := c.encoder.DecodeUnsignedTransaction(txBytes)
	if err != nil {
		return nil, meshcommon.GetError(meshcommon.ErrFailedToDecodeUnsignedTransaction)
	}

	// Apply signatures to Mesh transaction
	if len(req.Signatures) == 2 {
		// VIP191 Fee Delegation with two signatures
		originSig := req.Signatures[0]
		delegatorSig := req.Signatures[1]

		// Combine signatures for VIP191
		combinedSig := append(originSig.Bytes, delegatorSig.Bytes...)
		meshTx.Transaction = meshTx.WithSignature(combinedSig)

	} else if len(req.Signatures) == 1 {
		// Regular transaction: only origin signature
		sig := req.Signatures[0]
		meshTx.Transaction = meshTx.WithSignature(sig.Bytes)
	} else {
		return nil, meshcommon.GetError(meshcommon.ErrInvalidNumberOfSignatures)
	}

	// Encode signed Mesh transaction
	signedTxBytes, err := c.encoder.EncodeTransaction(meshTx)
	if err != nil {
		return nil, meshcommon.GetError(meshcommon.ErrFailedToEncodeSignedTransaction)
	}

	return &types.ConstructionCombineResponse{
		SignedTransaction: fmt.Sprintf("0x%s", hex.EncodeToString(signedTxBytes)),
	}, nil
}

// ConstructionHash gets the hash of a transaction
func (c *ConstructionService) ConstructionHash(
	ctx context.Context,
	req *types.ConstructionHashRequest,
) (*types.TransactionIdentifierResponse, *types.Error) {
	txBytes, err := c.bytesHandler.DecodeHexStringWithPrefix(req.SignedTransaction)
	if err != nil {
		return nil, meshcommon.GetError(meshcommon.ErrInvalidTransactionHex)
	}

	meshTx, err := c.encoder.DecodeSignedTransaction(txBytes)
	if err != nil {
		return nil, meshcommon.GetErrorWithMetadata(meshcommon.ErrFailedToDecodeMeshTransaction, map[string]any{"error": err.Error()})
	}

	return &types.TransactionIdentifierResponse{
		TransactionIdentifier: &types.TransactionIdentifier{
			Hash: meshTx.Transaction.ID().String(),
		},
	}, nil
}

// ConstructionSubmit submits a transaction to the network
func (c *ConstructionService) ConstructionSubmit(
	ctx context.Context,
	req *types.ConstructionSubmitRequest,
) (*types.TransactionIdentifierResponse, *types.Error) {
	// Decode transaction using our utility method
	txBytes, err := c.bytesHandler.DecodeHexStringWithPrefix(req.SignedTransaction)
	if err != nil {
		return nil, meshcommon.GetError(meshcommon.ErrInvalidTransactionHex)
	}

	// Decode Mesh transaction to get the native Thor transaction
	meshTx, err := c.encoder.DecodeSignedTransaction(txBytes)
	if err != nil {
		return nil, meshcommon.GetError(meshcommon.ErrFailedToDecodeMeshTransaction)
	}

	// Submit the native Thor transaction to VeChain network
	txID, err := c.vechainClient.SubmitTransaction(meshTx.Transaction)
	if err != nil {
		return nil, meshcommon.GetErrorWithMetadata(meshcommon.ErrFailedToSubmitTransaction, map[string]any{
			"error": err.Error(),
		})
	}

	return &types.TransactionIdentifierResponse{
		TransactionIdentifier: &types.TransactionIdentifier{
			Hash: txID,
		},
	}, nil
}

// getBasicTransactionInfo gets basic transaction information from the network
func (c *ConstructionService) getBasicTransactionInfo() (*api.JSONExpandedBlock, int, error) {
	bestBlock, err := c.vechainClient.GetBlock("best")
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get best block: %w", err)
	}

	chainTag, err := c.vechainClient.GetChainID()
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get chain tag: %w", err)
	}

	return bestBlock, chainTag, nil
}

// calculateGas calculates gas based on clauses and applies a 20% buffer
func (c *ConstructionService) calculateGas(options map[string]any) (uint64, error) {
	clausesRaw, ok := options["clauses"]
	if !ok {
		return 0, fmt.Errorf("clauses are required for gas calculation")
	}

	txClauses, err := c.clauseParser.ParseClausesFromOptions(clausesRaw)
	if err != nil {
		return 0, fmt.Errorf("failed to parse clauses: %w", err)
	}

	if len(txClauses) == 0 {
		return 0, fmt.Errorf("at least one clause is required for a valid transaction")
	}

	intrinsicGas, err := tx.IntrinsicGas(txClauses...)
	if err != nil {
		return 0, fmt.Errorf("failed to calculate intrinsic gas: %w", err)
	}

	return uint64(float64(intrinsicGas) * 1.2), nil
}

// buildMetadata builds metadata based on transaction type
func (c *ConstructionService) buildMetadata(transactionType, blockRef string, chainTag, gas uint64, nonce string) (map[string]any, *big.Int, error) {
	if transactionType == meshcommon.TransactionTypeLegacy {
		return c.buildLegacyMetadata(blockRef, chainTag, gas, nonce)
	}
	return c.buildDynamicMetadata(blockRef, chainTag, gas, nonce)
}

// buildLegacyMetadata builds metadata for legacy transactions
func (c *ConstructionService) buildLegacyMetadata(blockRef string, chainTag, gas uint64, nonce string) (map[string]any, *big.Int, error) {
	// Generate random gasPriceCoef (0-255)
	randomBytes := make([]byte, 1)
	if _, err := rand.Read(randomBytes); err != nil {
		return nil, nil, fmt.Errorf("failed to generate random gas price coefficient: %w", err)
	}
	gasPriceCoef := randomBytes[0]

	metadata := map[string]any{
		"transactionType": meshcommon.TransactionTypeLegacy,
		"blockRef":        blockRef,
		"chainTag":        chainTag,
		"gas":             gas,
		"nonce":           nonce,
		"gasPriceCoef":    gasPriceCoef,
	}

	return metadata, c.config.GetBaseGasPrice(), nil
}

// buildDynamicMetadata builds metadata for dynamic fee transactions
func (c *ConstructionService) buildDynamicMetadata(blockRef string, chainTag, gas uint64, nonce string) (map[string]any, *big.Int, error) {
	// Get dynamic gas price from network
	dynamicGasPrice, err := c.vechainClient.GetDynamicGasPrice()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get dynamic gas price: %w", err)
	}

	if dynamicGasPrice.BaseFee.Cmp(big.NewInt(0)) == 0 {
		// Case where we are building a dynamic fee transaction but the base fee is 0
		// This happens when the node is catching up with the chain block-wise
		metadata := map[string]any{
			"transactionType":      meshcommon.TransactionTypeDynamic,
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
		"transactionType":      meshcommon.TransactionTypeDynamic,
		"blockRef":             blockRef,
		"chainTag":             chainTag,
		"gas":                  gas,
		"nonce":                nonce,
		"maxFeePerGas":         gasPrice.String(),
		"maxPriorityFeePerGas": dynamicGasPrice.Reward.String(),
	}

	return metadata, gasPrice, nil
}

// createSigningPayloads creates signing payloads for the transaction
func (c *ConstructionService) createSigningPayloads(vechainTx *tx.Transaction, request types.ConstructionPayloadsRequest) ([]*types.SigningPayload, error) {
	var payloads []*types.SigningPayload

	// Check for fee delegation from metadata
	txDelegator := c.operationsExtractor.GetFeeDelegatorAccount(request.Metadata)
	hasFeeDelegation := txDelegator != ""

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
			Address: strings.ToLower(originAddress.Hex()),
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
	originAddr, err := crypto.DecompressPubkey(publicKeys[0].Bytes)
	if err != nil {
		return nil, fmt.Errorf("invalid origin public key: %w", err)
	}
	originAddress := crypto.PubkeyToAddress(*originAddr)
	thorOriginAddr, err := thor.ParseAddress(originAddress.Hex())
	if err != nil {
		return nil, fmt.Errorf("invalid origin address: %w", err)
	}
	hash := vechainTx.DelegatorSigningHash(thorOriginAddr)

	return &types.SigningPayload{
		AccountIdentifier: &types.AccountIdentifier{
			Address: strings.ToLower(delegatorAddress.Hex()),
		},
		Bytes:         hash[:],
		SignatureType: types.EcdsaRecovery,
	}, nil
}
