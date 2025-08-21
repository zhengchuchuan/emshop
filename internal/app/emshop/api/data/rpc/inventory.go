package rpc

import (
	"context"
	"time"
	ipb "emshop/api/inventory/v1"
	"emshop/internal/app/emshop/api/data"
	"emshop/pkg/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type inventory struct {
	client ipb.InventoryClient
}




// NewInventory 返回一个 InventoryData 的实例.
func NewInventory(client ipb.InventoryClient) data.InventoryData {
	return &inventory{client: client}
}






// InvDetail 带重试机制的库存详情查询
func (i *inventory) InvDetail(ctx context.Context, request *ipb.GoodsInvInfo) (*ipb.GoodsInvInfo, error) {
	// 实现重试机制，最多重试3次
	maxRetries := 3
	for attempt := 0; attempt < maxRetries; attempt++ {
		result, err := i.client.InvDetail(ctx, request)
		
		if err == nil {
			return result, nil
		}
		
		// 检查是否是连接相关错误
		if st, ok := status.FromError(err); ok {
			switch st.Code() {
			case codes.Unavailable, codes.DeadlineExceeded, codes.ResourceExhausted:
				if attempt < maxRetries-1 {
					waitTime := time.Duration(attempt+1) * 100 * time.Millisecond
					log.Warnf("Inventory service call failed (attempt %d/%d): %v, retrying in %v", 
						attempt+1, maxRetries, err, waitTime)
					time.Sleep(waitTime)
					continue
				}
			default:
				return nil, err
			}
		} else {
			return nil, err
		}
	}
	
	return nil, status.Error(codes.Unavailable, "inventory service unavailable after retries")
}


var _ data.InventoryData = &inventory{}