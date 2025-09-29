package validation

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"slices"
	"strings"

	"github.com/coinbase/rosetta-sdk-go/types"
	meshoperations "github.com/vechain/mesh/common/operations"
)

// ValidationMiddleware handles request validation
type ValidationMiddleware struct {
	networkIdentifier   *types.NetworkIdentifier
	runMode             string
	operationsExtractor *meshoperations.OperationsExtractor
}

// ValidationType represents the type of validation to perform
type ValidationType int

const (
	ValidationNetwork ValidationType = iota
	ValidationRunMode
	ValidationModeNetwork
	ValidationAccount
	ValidationConstructionPayloads
)

// Common validation sets for different endpoints
var (
	// AccountBalanceValidations includes all validations needed for account/balance
	AccountBalanceValidations = []ValidationType{
		ValidationNetwork,
		ValidationRunMode,
		ValidationModeNetwork,
		ValidationAccount,
	}

	// NetworkValidations includes validations needed for network endpoints
	NetworkValidations = []ValidationType{
		ValidationNetwork,
		ValidationRunMode,
		ValidationModeNetwork,
	}

	// NetworkListValidations includes validations needed for network/list (no network identifier required)
	NetworkListValidations = []ValidationType{
		ValidationRunMode,
	}

	// ConstructionValidations includes validations needed for construction endpoints
	ConstructionValidations = []ValidationType{
		ValidationNetwork,
		ValidationRunMode,
		ValidationModeNetwork,
	}

	// ConstructionPayloadsValidations includes specific validations for construction/payloads
	ConstructionPayloadsValidations = []ValidationType{
		ValidationNetwork,
		ValidationRunMode,
		ValidationModeNetwork,
		ValidationConstructionPayloads,
	}
)

// NewValidationMiddleware creates a new validation middleware
func NewValidationMiddleware(networkIdentifier *types.NetworkIdentifier, runMode string) *ValidationMiddleware {
	return &ValidationMiddleware{
		networkIdentifier:   networkIdentifier,
		runMode:             runMode,
		operationsExtractor: meshoperations.NewOperationsExtractor(),
	}
}

// CheckNetwork validates the network identifier in the request
func (v *ValidationMiddleware) CheckNetwork(w http.ResponseWriter, r *http.Request, requestData []byte) bool {
	var request struct {
		NetworkIdentifier *types.NetworkIdentifier `json:"network_identifier"`
	}

	// Parse the request
	if err := json.Unmarshal(requestData, &request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return false
	}

	// Validate network identifier
	if request.NetworkIdentifier == nil {
		http.Error(w, "Network identifier is required", http.StatusBadRequest)
		return false
	}

	// Check blockchain
	if request.NetworkIdentifier.Blockchain != v.networkIdentifier.Blockchain {
		http.Error(w, fmt.Sprintf("Invalid blockchain: expected %s, got %s",
			v.networkIdentifier.Blockchain, request.NetworkIdentifier.Blockchain), http.StatusBadRequest)
		return false
	}

	// Check network
	if request.NetworkIdentifier.Network != v.networkIdentifier.Network {
		http.Error(w, fmt.Sprintf("Invalid network: expected %s, got %s",
			v.networkIdentifier.Network, request.NetworkIdentifier.Network), http.StatusBadRequest)
		return false
	}

	// Check sub network identifier if present
	if request.NetworkIdentifier.SubNetworkIdentifier != nil {
		if v.networkIdentifier.SubNetworkIdentifier == nil ||
			request.NetworkIdentifier.SubNetworkIdentifier.Network != v.networkIdentifier.SubNetworkIdentifier.Network {
			http.Error(w, "Invalid sub network identifier", http.StatusBadRequest)
			return false
		}
	}

	return true
}

// CheckRunMode validates the run mode
func (v *ValidationMiddleware) CheckRunMode(w http.ResponseWriter, r *http.Request) bool {
	// In Mesh, run mode is typically "online" or "offline"
	if v.runMode != "online" {
		http.Error(w, fmt.Sprintf("Invalid run mode: this endpoint requires online mode, got %s", v.runMode), http.StatusBadRequest)
		return false
	}
	return true
}

// CheckModeNetwork validates that the mode and network are compatible
func (v *ValidationMiddleware) CheckModeNetwork(w http.ResponseWriter, r *http.Request) bool {
	validNetworks := []string{"main", "test", "solo"}

	isValidNetwork := slices.Contains(validNetworks, v.networkIdentifier.Network)

	if !isValidNetwork {
		http.Error(w, fmt.Sprintf("Unsupported network: %s", v.networkIdentifier.Network), http.StatusBadRequest)
		return false
	}

	// For online mode, we need to ensure the network is accessible
	if v.runMode == "online" && !isValidNetwork {
		http.Error(w, "Network not accessible in online mode", http.StatusBadRequest)
		return false
	}

	return true
}

// CheckAccount validates the account identifier
func (v *ValidationMiddleware) CheckAccount(w http.ResponseWriter, r *http.Request, requestData []byte) bool {
	var request struct {
		AccountIdentifier *types.AccountIdentifier `json:"account_identifier"`
	}

	// Parse the request
	if err := json.Unmarshal(requestData, &request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return false
	}

	// Validate account identifier
	if request.AccountIdentifier == nil {
		http.Error(w, "Account identifier is required", http.StatusBadRequest)
		return false
	}

	// Validate address format
	if !v.isValidVeChainAddress(request.AccountIdentifier.Address) {
		http.Error(w, fmt.Sprintf("Invalid VeChain address format: %s", request.AccountIdentifier.Address), http.StatusBadRequest)
		return false
	}

	return true
}

// isValidVeChainAddress validates VeChain address format
func (v *ValidationMiddleware) isValidVeChainAddress(address string) bool {
	if len(address) != 42 {
		return false
	}

	if !strings.HasPrefix(address, "0x") {
		return false
	}

	hexPattern := regexp.MustCompile(`^0x[0-9a-fA-F]{40}$`)
	return hexPattern.MatchString(address)
}

// CheckConstructionPayloads validates construction payloads specific requirements
func (v *ValidationMiddleware) CheckConstructionPayloads(w http.ResponseWriter, r *http.Request, requestData []byte) bool {
	var request struct {
		Operations []*types.Operation `json:"operations"`
		PublicKeys []*types.PublicKey `json:"public_keys"`
		Metadata   map[string]any     `json:"metadata"`
	}

	// Parse the request
	if err := json.Unmarshal(requestData, &request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return false
	}

	// Validate operations
	if len(request.Operations) == 0 {
		http.Error(w, "Operations are required", http.StatusBadRequest)
		return false
	}

	// Validate public keys
	if len(request.PublicKeys) == 0 {
		http.Error(w, "Public keys are required", http.StatusBadRequest)
		return false
	}
	if len(request.PublicKeys) > 2 {
		http.Error(w, "Too many public keys provided", http.StatusBadRequest)
		return false
	}

	// Validate metadata
	if request.Metadata == nil {
		http.Error(w, "Metadata is required", http.StatusBadRequest)
		return false
	}

	// Check fee delegation requirements
	txDelegator := ""
	if delegator, ok := request.Metadata["fee_delegator_account"].(string); ok {
		txDelegator = strings.ToLower(delegator)
	}

	if txDelegator != "" && len(request.PublicKeys) != 2 {
		http.Error(w, "Fee delegation requires exactly 2 public keys", http.StatusBadRequest)
		return false
	}

	// Validate origins (should be exactly 1)
	origins := v.operationsExtractor.GetTxOrigins(request.Operations)
	if len(origins) == 0 {
		http.Error(w, "No origin found in operations", http.StatusBadRequest)
		return false
	}
	if len(origins) > 1 {
		http.Error(w, "Multiple origins found in operations", http.StatusBadRequest)
		return false
	}

	return true
}

// ValidateRequest performs validations based on the provided validation types
func (v *ValidationMiddleware) ValidateRequest(w http.ResponseWriter, r *http.Request, requestData []byte, validations []ValidationType) bool {
	for _, validation := range validations {
		switch validation {
		case ValidationNetwork:
			if !v.CheckNetwork(w, r, requestData) {
				return false
			}
		case ValidationRunMode:
			if !v.CheckRunMode(w, r) {
				return false
			}
		case ValidationModeNetwork:
			if !v.CheckModeNetwork(w, r) {
				return false
			}
		case ValidationAccount:
			if !v.CheckAccount(w, r, requestData) {
				return false
			}
		case ValidationConstructionPayloads:
			if !v.CheckConstructionPayloads(w, r, requestData) {
				return false
			}
		}
	}
	return true
}

// ValidateEndpoint performs validations for a specific endpoint
func (v *ValidationMiddleware) ValidateEndpoint(w http.ResponseWriter, r *http.Request, requestData []byte, endpoint string) bool {
	validations := GetValidationsForEndpoint(endpoint)
	return v.ValidateRequest(w, r, requestData, validations)
}
