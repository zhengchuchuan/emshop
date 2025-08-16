package rpc

import (
	"context"
	ipb "emshop/api/inventory/v1"
	"emshop/gin-micro/registry"
	rpcserver "emshop/gin-micro/server/rpc-server"
	clientinterceptors "emshop/gin-micro/server/rpc-server/client-interceptors"
	"emshop/internal/app/emshop/api/data"
	"emshop/pkg/log"
	"time"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type inventory struct {
	client ipb.InventoryClient
}


const inventoryServiceName = "discovery:///emshop-inventory-srv"
const fallbackInventoryAddress = "127.0.0.1:28055"


// NewInventory 返回一个 InventoryData 的实例.
func NewInventory(client ipb.InventoryClient) data.InventoryData {
	return &inventory{client: client}
}

// NewInventoryServiceClient 基于服务发现生成grpc连接，支持健壮的重试和fallback机制
func NewInventoryServiceClient(r registry.Discovery) ipb.InventoryClient {
	log.Infof("Initializing gRPC connection to service: %s", inventoryServiceName)
	
	// 首先尝试服务发现连接，使用更健壮的配置
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	log.Infof("Attempting service discovery connection to: %s", inventoryServiceName)
	conn, err := rpcserver.DialInsecure(
		ctx,
		rpcserver.WithEndpoint(inventoryServiceName),
		rpcserver.WithDiscovery(r),
		rpcserver.WithClientTimeout(15*time.Second),
		rpcserver.WithClientUnaryInterceptor(clientinterceptors.UnaryTracingInterceptor),
	)
	
	if err != nil {
		log.Warnf("Service discovery connection failed: %v, falling back to direct connection", err)
		// fallback到直连
		ctx2, cancel2 := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel2()
		
		log.Infof("Attempting direct connection to: %s", fallbackInventoryAddress)
		conn, err = rpcserver.DialInsecure(
			ctx2,
			rpcserver.WithEndpoint(fallbackInventoryAddress),
			rpcserver.WithClientTimeout(15*time.Second),
			rpcserver.WithClientUnaryInterceptor(clientinterceptors.UnaryTracingInterceptor),
		)
		
		if err != nil {
			log.Fatalf("Both service discovery and direct connection failed: %v", err)
		}
		log.Infof("Successfully connected to inventory service via direct connection fallback")
	} else {
		log.Infof("Successfully connected to inventory service via service discovery")
		// 即使服务发现成功，也要测试连接是否真的可用
		// 如果连接有问题，立即切换到localhost fallback
		log.Infof("Testing service discovery connection...")
		testCtx, testCancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer testCancel()
		
		testClient := ipb.NewInventoryClient(conn)
		_, testErr := testClient.InvDetail(testCtx, &ipb.GoodsInvInfo{GoodsId: 1})
		
		if testErr != nil {
			log.Warnf("Service discovery connection test failed: %v, switching to localhost fallback", testErr)
			conn.Close()
			
			// 立即尝试localhost连接
			fallbackCtx, fallbackCancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer fallbackCancel()
			
			log.Infof("Attempting localhost fallback connection to: %s", fallbackInventoryAddress)
			conn, err = rpcserver.DialInsecure(
				fallbackCtx,
				rpcserver.WithEndpoint(fallbackInventoryAddress),
				rpcserver.WithClientTimeout(15*time.Second),
				rpcserver.WithClientUnaryInterceptor(clientinterceptors.UnaryTracingInterceptor),
			)
			
			if err != nil {
				log.Fatalf("Localhost fallback connection also failed: %v", err)
			}
			log.Infof("Successfully connected to inventory service via localhost fallback")
		} else {
			log.Infof("Service discovery connection test successful")
		}
	}
	
	return ipb.NewInventoryClient(conn)
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