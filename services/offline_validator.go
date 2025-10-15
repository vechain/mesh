package services

import (
	"context"

	"github.com/coinbase/rosetta-sdk-go/types"
	meshcommon "github.com/vechain/mesh/common"
	meshconfig "github.com/vechain/mesh/config"
)

// OfflineNetworkService wraps NetworkService to handle offline mode
type OfflineNetworkService struct {
	*NetworkService
	config *meshconfig.Config
}

// NewOfflineNetworkService creates a wrapped network service with offline checks
func NewOfflineNetworkService(service *NetworkService, config *meshconfig.Config) *OfflineNetworkService {
	return &OfflineNetworkService{
		NetworkService: service,
		config:         config,
	}
}

// NetworkStatus checks offline mode before calling the real implementation
func (o *OfflineNetworkService) NetworkStatus(ctx context.Context, req *types.NetworkRequest) (*types.NetworkStatusResponse, *types.Error) {
	if o.config.Mode == meshcommon.OfflineMode {
		return nil, meshcommon.GetError(meshcommon.ErrAPIDoesNotSupportOfflineMode)
	}
	return o.NetworkService.NetworkStatus(ctx, req)
}

// OfflineAccountService wraps AccountService to handle offline mode
type OfflineAccountService struct {
	*AccountService
	config *meshconfig.Config
}

// NewOfflineAccountService creates a wrapped account service with offline checks
func NewOfflineAccountService(service *AccountService, config *meshconfig.Config) *OfflineAccountService {
	return &OfflineAccountService{
		AccountService: service,
		config:         config,
	}
}

// AccountBalance checks offline mode before calling the real implementation
func (o *OfflineAccountService) AccountBalance(ctx context.Context, req *types.AccountBalanceRequest) (*types.AccountBalanceResponse, *types.Error) {
	if o.config.Mode == meshcommon.OfflineMode {
		return nil, meshcommon.GetError(meshcommon.ErrAPIDoesNotSupportOfflineMode)
	}
	return o.AccountService.AccountBalance(ctx, req)
}

// OfflineBlockService wraps BlockService to handle offline mode
type OfflineBlockService struct {
	*BlockService
	config *meshconfig.Config
}

// NewOfflineBlockService creates a wrapped block service with offline checks
func NewOfflineBlockService(service *BlockService, config *meshconfig.Config) *OfflineBlockService {
	return &OfflineBlockService{
		BlockService: service,
		config:       config,
	}
}

// Block checks offline mode before calling the real implementation
func (o *OfflineBlockService) Block(ctx context.Context, req *types.BlockRequest) (*types.BlockResponse, *types.Error) {
	if o.config.Mode == meshcommon.OfflineMode {
		return nil, meshcommon.GetError(meshcommon.ErrAPIDoesNotSupportOfflineMode)
	}
	return o.BlockService.Block(ctx, req)
}

// BlockTransaction checks offline mode before calling the real implementation
func (o *OfflineBlockService) BlockTransaction(ctx context.Context, req *types.BlockTransactionRequest) (*types.BlockTransactionResponse, *types.Error) {
	if o.config.Mode == meshcommon.OfflineMode {
		return nil, meshcommon.GetError(meshcommon.ErrAPIDoesNotSupportOfflineMode)
	}
	return o.BlockService.BlockTransaction(ctx, req)
}

// OfflineConstructionService wraps ConstructionService to handle offline mode
type OfflineConstructionService struct {
	*ConstructionService
	config *meshconfig.Config
}

// NewOfflineConstructionService creates a wrapped construction service with offline checks
func NewOfflineConstructionService(service *ConstructionService, config *meshconfig.Config) *OfflineConstructionService {
	return &OfflineConstructionService{
		ConstructionService: service,
		config:              config,
	}
}

// ConstructionMetadata checks offline mode before calling the real implementation
func (o *OfflineConstructionService) ConstructionMetadata(ctx context.Context, req *types.ConstructionMetadataRequest) (*types.ConstructionMetadataResponse, *types.Error) {
	if o.config.Mode == meshcommon.OfflineMode {
		return nil, meshcommon.GetError(meshcommon.ErrAPIDoesNotSupportOfflineMode)
	}
	return o.ConstructionService.ConstructionMetadata(ctx, req)
}

// ConstructionSubmit checks offline mode before calling the real implementation
func (o *OfflineConstructionService) ConstructionSubmit(ctx context.Context, req *types.ConstructionSubmitRequest) (*types.TransactionIdentifierResponse, *types.Error) {
	if o.config.Mode == meshcommon.OfflineMode {
		return nil, meshcommon.GetError(meshcommon.ErrAPIDoesNotSupportOfflineMode)
	}
	return o.ConstructionService.ConstructionSubmit(ctx, req)
}

// OfflineMempoolService wraps MempoolService to handle offline mode
type OfflineMempoolService struct {
	*MempoolService
	config *meshconfig.Config
}

// NewOfflineMempoolService creates a wrapped mempool service with offline checks
func NewOfflineMempoolService(service *MempoolService, config *meshconfig.Config) *OfflineMempoolService {
	return &OfflineMempoolService{
		MempoolService: service,
		config:         config,
	}
}

// Mempool checks offline mode before calling the real implementation
func (o *OfflineMempoolService) Mempool(ctx context.Context, req *types.NetworkRequest) (*types.MempoolResponse, *types.Error) {
	if o.config.Mode == meshcommon.OfflineMode {
		return nil, meshcommon.GetError(meshcommon.ErrAPIDoesNotSupportOfflineMode)
	}
	return o.MempoolService.Mempool(ctx, req)
}

// MempoolTransaction checks offline mode before calling the real implementation
func (o *OfflineMempoolService) MempoolTransaction(ctx context.Context, req *types.MempoolTransactionRequest) (*types.MempoolTransactionResponse, *types.Error) {
	if o.config.Mode == meshcommon.OfflineMode {
		return nil, meshcommon.GetError(meshcommon.ErrAPIDoesNotSupportOfflineMode)
	}
	return o.MempoolService.MempoolTransaction(ctx, req)
}

// OfflineEventsService wraps EventsService to handle offline mode
type OfflineEventsService struct {
	*EventsService
	config *meshconfig.Config
}

// NewOfflineEventsService creates a wrapped events service with offline checks
func NewOfflineEventsService(service *EventsService, config *meshconfig.Config) *OfflineEventsService {
	return &OfflineEventsService{
		EventsService: service,
		config:        config,
	}
}

// EventsBlocks checks offline mode before calling the real implementation
func (o *OfflineEventsService) EventsBlocks(ctx context.Context, req *types.EventsBlocksRequest) (*types.EventsBlocksResponse, *types.Error) {
	if o.config.Mode == meshcommon.OfflineMode {
		return nil, meshcommon.GetError(meshcommon.ErrAPIDoesNotSupportOfflineMode)
	}
	return o.EventsService.EventsBlocks(ctx, req)
}

// OfflineSearchService wraps SearchService to handle offline mode
type OfflineSearchService struct {
	*SearchService
	config *meshconfig.Config
}

// NewOfflineSearchService creates a wrapped search service with offline checks
func NewOfflineSearchService(service *SearchService, config *meshconfig.Config) *OfflineSearchService {
	return &OfflineSearchService{
		SearchService: service,
		config:        config,
	}
}

// SearchTransactions checks offline mode before calling the real implementation
func (o *OfflineSearchService) SearchTransactions(ctx context.Context, req *types.SearchTransactionsRequest) (*types.SearchTransactionsResponse, *types.Error) {
	if o.config.Mode == meshcommon.OfflineMode {
		return nil, meshcommon.GetError(meshcommon.ErrAPIDoesNotSupportOfflineMode)
	}
	return o.SearchService.SearchTransactions(ctx, req)
}

// OfflineCallService wraps CallService to handle offline mode
type OfflineCallService struct {
	*CallService
	config *meshconfig.Config
}

// NewOfflineCallService creates a wrapped call service with offline checks
func NewOfflineCallService(service *CallService, config *meshconfig.Config) *OfflineCallService {
	return &OfflineCallService{
		CallService: service,
		config:      config,
	}
}

// Call checks offline mode before calling the real implementation
func (o *OfflineCallService) Call(ctx context.Context, req *types.CallRequest) (*types.CallResponse, *types.Error) {
	if o.config.Mode == meshcommon.OfflineMode {
		return nil, meshcommon.GetError(meshcommon.ErrAPIDoesNotSupportOfflineMode)
	}
	return o.CallService.Call(ctx, req)
}
