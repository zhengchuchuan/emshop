package rpc

import (
	"context"
	lpbv1 "emshop/api/logistics/v1"
	"emshop/internal/app/api/emshop/data"
	"emshop/pkg/log"
	"google.golang.org/protobuf/types/known/emptypb"
)

type logistics struct {
	lc lpbv1.LogisticsClient
}

func NewLogistics(lc lpbv1.LogisticsClient) *logistics {
	return &logistics{lc}
}

// GetLogisticsInfo 获取物流信息
func (l *logistics) GetLogisticsInfo(ctx context.Context, request *lpbv1.GetLogisticsInfoRequest) (*lpbv1.GetLogisticsInfoResponse, error) {
	log.Infof("Calling GetLogisticsInfo gRPC with request: %+v", request)
	response, err := l.lc.GetLogisticsInfo(ctx, request)
	if err != nil {
		log.Errorf("GetLogisticsInfo gRPC call failed: %v", err)
		return nil, err
	}
	log.Infof("GetLogisticsInfo gRPC call successful, logisticsSn: %s, status: %d",
		response.LogisticsSn, response.LogisticsStatus)
	return response, nil
}

// GetLogisticsTracks 获取物流轨迹
func (l *logistics) GetLogisticsTracks(ctx context.Context, request *lpbv1.GetLogisticsTracksRequest) (*lpbv1.GetLogisticsTracksResponse, error) {
	log.Infof("Calling GetLogisticsTracks gRPC with request: %+v", request)
	response, err := l.lc.GetLogisticsTracks(ctx, request)
	if err != nil {
		log.Errorf("GetLogisticsTracks gRPC call failed: %v", err)
		return nil, err
	}
	log.Infof("GetLogisticsTracks gRPC call successful, tracks count: %d", len(response.Tracks))
	return response, nil
}

// CalculateShippingFee 计算运费
func (l *logistics) CalculateShippingFee(ctx context.Context, request *lpbv1.CalculateShippingFeeRequest) (*lpbv1.CalculateShippingFeeResponse, error) {
	log.Infof("Calling CalculateShippingFee gRPC with weight: %.2f, volume: %.2f",
		request.TotalWeight, request.TotalVolume)
	response, err := l.lc.CalculateShippingFee(ctx, request)
	if err != nil {
		log.Errorf("CalculateShippingFee gRPC call failed: %v", err)
		return nil, err
	}
	log.Infof("CalculateShippingFee gRPC call successful, shippingFee: %.2f", response.ShippingFee)
	return response, nil
}

// GetLogisticsCompanies 获取物流公司列表
func (l *logistics) GetLogisticsCompanies(ctx context.Context) (*lpbv1.LogisticsCompaniesResponse, error) {
	log.Infof("Calling GetLogisticsCompanies gRPC")
	response, err := l.lc.GetLogisticsCompanies(ctx, &emptypb.Empty{})
	if err != nil {
		log.Errorf("GetLogisticsCompanies gRPC call failed: %v", err)
		return nil, err
	}
	log.Infof("GetLogisticsCompanies gRPC call successful, companies count: %d", len(response.Companies))
	return response, nil
}

var _ data.LogisticsData = &logistics{}
