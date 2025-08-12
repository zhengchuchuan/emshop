package v1

import (
	"context"
	invpb "emshop/api/inventory/v1"
	"emshop/internal/app/inventory/srv/domain/do"
	"emshop/internal/app/inventory/srv/domain/dto"
	v1 "emshop/internal/app/inventory/srv/service/v1"
	"emshop/internal/app/pkg/code"
	"emshop/pkg/errors"
	"emshop/pkg/log"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type inventoryServer struct {
	invpb.UnimplementedInventoryServer
	srv v1.ServiceFactory
}

// 设置库存
func (is *inventoryServer) SetInv(ctx context.Context, info *invpb.GoodsInvInfo) (*emptypb.Empty, error) {
	invDTO := &dto.InventoryDTO{}
	invDTO.Goods = info.GoodsId
	invDTO.Stocks = info.Num
	err := is.srv.Inventorys().Create(ctx, invDTO)
	if err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (is *inventoryServer) InvDetail(ctx context.Context, info *invpb.GoodsInvInfo) (*invpb.GoodsInvInfo, error) {
	inv, err := is.srv.Inventorys().Get(ctx, uint64(info.GoodsId))
	if err != nil {
		return nil, err
	}
	return &invpb.GoodsInvInfo{
		GoodsId: inv.Goods,
		Num:     inv.Stocks,
	}, nil
}

func (is *inventoryServer) Sell(ctx context.Context, info *invpb.SellInfo) (*emptypb.Empty, error) {
	var detail []do.GoodsDetail
	for _, value := range info.GoodsInfo {
		detail = append(detail, do.GoodsDetail{Goods: value.GoodsId, Num: value.Num})
	}
	err := is.srv.Inventorys().Sell(ctx, info.OrderSn, detail)
	if err != nil {
		if errors.IsCode(err, code.ErrInvNotEnough) {
			return nil, status.Errorf(codes.Aborted, "%s", err.Error())
		}
		return nil, err
	}
	//time.Sleep(5 * time.Second)
	//return nil, status.Errorf(codes.Aborted, " err.Error()")
	return &emptypb.Empty{}, nil
}

// Reback 库存归还
//	@param ctx 
//	@param info 
//	@return *emptypb.Empty 
//	@return error 
func (is *inventoryServer) Reback(ctx context.Context, info *invpb.SellInfo) (*emptypb.Empty, error) {
	log.Infof("订单%s归还库存", info.OrderSn)
	var detail []do.GoodsDetail
	for _, v := range info.GoodsInfo {
		detail = append(detail, do.GoodsDetail{Goods: v.GoodsId, Num: v.Num})
	}
	err := is.srv.Inventorys().Reback(ctx, info.OrderSn, detail)
	if err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func NewInventoryServer(srv v1.ServiceFactory) *inventoryServer {
	return &inventoryServer{srv: srv}
}

var (
	_ invpb.InventoryServer = &inventoryServer{}
)
