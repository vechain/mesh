package services

import (
	"context"

	"github.com/coinbase/rosetta-sdk-go/types"
	meshcommon "github.com/vechain/mesh/common"
	meshconfig "github.com/vechain/mesh/config"
)

// requireOnlineMode returns an error if in offline mode, nil otherwise
func requireOnlineMode(config *meshconfig.Config) *types.Error {
	if config.Mode == meshcommon.OfflineMode {
		return meshcommon.GetError(meshcommon.ErrAPIDoesNotSupportOfflineMode)
	}
	return nil
}

// OfflineNetworkService wraps NetworkService with offline validation
type OfflineNetworkService struct {
	*NetworkService
	config *meshconfig.Config
}

func NewOfflineNetworkService(service *NetworkService, config *meshconfig.Config) *OfflineNetworkService {
	return &OfflineNetworkService{NetworkService: service, config: config}
}

func (o *OfflineNetworkService) NetworkStatus(ctx context.Context, req *types.NetworkRequest) (*types.NetworkStatusResponse, *types.Error) {
	if err := requireOnlineMode(o.config); err != nil {
		return nil, err
	}
	return o.NetworkService.NetworkStatus(ctx, req)
}

// OfflineAccountService wraps AccountService with offline validation
type OfflineAccountService struct {
	*AccountService
	config *meshconfig.Config
}

func NewOfflineAccountService(service *AccountService, config *meshconfig.Config) *OfflineAccountService {
	return &OfflineAccountService{AccountService: service, config: config}
}

func (o *OfflineAccountService) AccountBalance(ctx context.Context, req *types.AccountBalanceRequest) (*types.AccountBalanceResponse, *types.Error) {
	if err := requireOnlineMode(o.config); err != nil {
		return nil, err
	}
	return o.AccountService.AccountBalance(ctx, req)
}

// OfflineBlockService wraps BlockService with offline validation
type OfflineBlockService struct {
	*BlockService
	config *meshconfig.Config
}

func NewOfflineBlockService(service *BlockService, config *meshconfig.Config) *OfflineBlockService {
	return &OfflineBlockService{BlockService: service, config: config}
}

func (o *OfflineBlockService) Block(ctx context.Context, req *types.BlockRequest) (*types.BlockResponse, *types.Error) {
	if err := requireOnlineMode(o.config); err != nil {
		return nil, err
	}
	return o.BlockService.Block(ctx, req)
}

func (o *OfflineBlockService) BlockTransaction(ctx context.Context, req *types.BlockTransactionRequest) (*types.BlockTransactionResponse, *types.Error) {
	if err := requireOnlineMode(o.config); err != nil {
		return nil, err
	}
	return o.BlockService.BlockTransaction(ctx, req)
}

// OfflineConstructionService wraps ConstructionService with offline validation
type OfflineConstructionService struct {
	*ConstructionService
	config *meshconfig.Config
}

func NewOfflineConstructionService(service *ConstructionService, config *meshconfig.Config) *OfflineConstructionService {
	return &OfflineConstructionService{ConstructionService: service, config: config}
}

func (o *OfflineConstructionService) ConstructionMetadata(ctx context.Context, req *types.ConstructionMetadataRequest) (*types.ConstructionMetadataResponse, *types.Error) {
	if err := requireOnlineMode(o.config); err != nil {
		return nil, err
	}
	return o.ConstructionService.ConstructionMetadata(ctx, req)
}

func (o *OfflineConstructionService) ConstructionSubmit(ctx context.Context, req *types.ConstructionSubmitRequest) (*types.TransactionIdentifierResponse, *types.Error) {
	if err := requireOnlineMode(o.config); err != nil {
		return nil, err
	}
	return o.ConstructionService.ConstructionSubmit(ctx, req)
}

// OfflineMempoolService wraps MempoolService with offline validation
type OfflineMempoolService struct {
	*MempoolService
	config *meshconfig.Config
}

func NewOfflineMempoolService(service *MempoolService, config *meshconfig.Config) *OfflineMempoolService {
	return &OfflineMempoolService{MempoolService: service, config: config}
}

func (o *OfflineMempoolService) Mempool(ctx context.Context, req *types.NetworkRequest) (*types.MempoolResponse, *types.Error) {
	if err := requireOnlineMode(o.config); err != nil {
		return nil, err
	}
	return o.MempoolService.Mempool(ctx, req)
}

func (o *OfflineMempoolService) MempoolTransaction(ctx context.Context, req *types.MempoolTransactionRequest) (*types.MempoolTransactionResponse, *types.Error) {
	if err := requireOnlineMode(o.config); err != nil {
		return nil, err
	}
	return o.MempoolService.MempoolTransaction(ctx, req)
}

// OfflineEventsService wraps EventsService with offline validation
type OfflineEventsService struct {
	*EventsService
	config *meshconfig.Config
}

func NewOfflineEventsService(service *EventsService, config *meshconfig.Config) *OfflineEventsService {
	return &OfflineEventsService{EventsService: service, config: config}
}

func (o *OfflineEventsService) EventsBlocks(ctx context.Context, req *types.EventsBlocksRequest) (*types.EventsBlocksResponse, *types.Error) {
	if err := requireOnlineMode(o.config); err != nil {
		return nil, err
	}
	return o.EventsService.EventsBlocks(ctx, req)
}

// OfflineSearchService wraps SearchService with offline validation
type OfflineSearchService struct {
	*SearchService
	config *meshconfig.Config
}

func NewOfflineSearchService(service *SearchService, config *meshconfig.Config) *OfflineSearchService {
	return &OfflineSearchService{SearchService: service, config: config}
}

func (o *OfflineSearchService) SearchTransactions(ctx context.Context, req *types.SearchTransactionsRequest) (*types.SearchTransactionsResponse, *types.Error) {
	if err := requireOnlineMode(o.config); err != nil {
		return nil, err
	}
	return o.SearchService.SearchTransactions(ctx, req)
}

// OfflineCallService wraps CallService with offline validation
type OfflineCallService struct {
	*CallService
	config *meshconfig.Config
}

func NewOfflineCallService(service *CallService, config *meshconfig.Config) *OfflineCallService {
	return &OfflineCallService{CallService: service, config: config}
}

func (o *OfflineCallService) Call(ctx context.Context, req *types.CallRequest) (*types.CallResponse, *types.Error) {
	if err := requireOnlineMode(o.config); err != nil {
		return nil, err
	}
	return o.CallService.Call(ctx, req)
}
