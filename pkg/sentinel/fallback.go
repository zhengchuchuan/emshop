package sentinel

import (
	"context"
	"time"

	cpb "emshop/api/coupon/v1"
	gpb "emshop/api/goods/v1" 
	ipb "emshop/api/inventory/v1"
	opb "emshop/api/order/v1"
	upb "emshop/api/user/v1"
	"emshop/pkg/log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// FallbackHandler 降级处理器接口
type FallbackHandler interface {
	Handle(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, err error) (interface{}, error)
}

// BusinessFallbackHandler 业务降级处理器
type BusinessFallbackHandler struct {
	serviceName string
	// 缓存配置
	enableCache bool
	// 默认响应配置
	defaultResponses map[string]interface{}
}

// NewBusinessFallbackHandler 创建业务降级处理器
func NewBusinessFallbackHandler(serviceName string, enableCache bool) *BusinessFallbackHandler {
	handler := &BusinessFallbackHandler{
		serviceName:      serviceName,
		enableCache:      enableCache,
		defaultResponses: make(map[string]interface{}),
	}

	// 初始化默认响应
	handler.initDefaultResponses()

	return handler
}

// Handle 处理降级逻辑
func (h *BusinessFallbackHandler) Handle(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, err error) (interface{}, error) {
	log.Warnf("执行业务降级处理: service=%s, method=%s, error=%v", h.serviceName, info.FullMethod, err)

	// 根据服务类型和方法名选择降级策略
	switch h.serviceName {
	case "emshop-coupon-srv":
		return h.handleCouponFallback(ctx, req, info, err)
	case "emshop-user-srv":
		return h.handleUserFallback(ctx, req, info, err)
	case "emshop-goods-srv":
		return h.handleGoodsFallback(ctx, req, info, err)
	case "emshop-inventory-srv":
		return h.handleInventoryFallback(ctx, req, info, err)
	case "emshop-order-srv":
		return h.handleOrderFallback(ctx, req, info, err)
	case "emshop-payment-srv":
		return h.handlePaymentFallback(ctx, req, info, err)
	default:
		return h.handleDefaultFallback(ctx, req, info, err)
	}
}

// handleCouponFallback 处理优惠券服务降级
func (h *BusinessFallbackHandler) handleCouponFallback(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, err error) (interface{}, error) {
	switch info.FullMethod {
	case "/coupon.v1.Coupon/GetUserCoupons":
		// 优惠券查询降级：返回空列表
		return &cpb.ListUserCouponsResponse{
			TotalCount: 0,
			Items:      []*cpb.UserCouponResponse{},
		}, nil

	case "/coupon.v1.Coupon/IssueCoupon":
		// 优惠券发放降级：返回失败
		return nil, status.Error(codes.Unavailable, "优惠券发放服务暂时不可用，请稍后重试")

	case "/coupon.v1.Coupon/UseCoupon":
		// 优惠券使用降级：返回失败，允许不使用优惠券继续下单
		return nil, status.Error(codes.Unavailable, "优惠券使用服务暂时不可用")

	case "/coupon.v1.Coupon/CalculateCouponDiscount":
		// 优惠券计算降级：返回无折扣
		return &cpb.CalculateCouponDiscountResponse{
			OriginalAmount:  0,
			DiscountAmount:  0,
			FinalAmount:     0,
			AppliedCoupons:  []int64{},
			RejectedCoupons: []*cpb.CouponRejection{},
		}, nil

	default:
		return nil, status.Error(codes.Unavailable, "优惠券服务暂时不可用")
	}
}

// handleUserFallback 处理用户服务降级
func (h *BusinessFallbackHandler) handleUserFallback(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, err error) (interface{}, error) {
	switch info.FullMethod {
	case "/user.v1.User/GetUserById":
		// 用户查询降级：从缓存返回基本信息或匿名用户
		if h.enableCache {
			// 尝试从缓存获取
			// 这里可以集成Redis缓存逻辑
		}

		// 返回匿名用户信息
		return &upb.UserInfoResponse{
			Id:       0,
			Mobile:   "***",
			NickName: "匿名用户",
		}, nil

	case "/user.v1.User/GetUserByMobile":
		// 手机号查询降级
		return nil, status.Error(codes.Unavailable, "用户服务暂时不可用")

	case "/user.v1.User/CreateUser":
		// 用户注册降级：暂停注册
		return nil, status.Error(codes.Unavailable, "用户注册服务暂时不可用，请稍后重试")

	default:
		return nil, status.Error(codes.Unavailable, "用户服务暂时不可用")
	}
}

// handleGoodsFallback 处理商品服务降级
func (h *BusinessFallbackHandler) handleGoodsFallback(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, err error) (interface{}, error) {
	switch info.FullMethod {
	case "/goods.v1.Goods/GoodsList":
		// 商品列表降级：返回缓存或热门商品
		return &gpb.GoodsListResponse{
			Total: 0,
			Data:  []*gpb.GoodsInfoResponse{},
		}, nil

	case "/goods.v1.Goods/GetGoodsDetail":
		// 商品详情降级：返回基本信息或缓存数据
		if h.enableCache {
			// 从缓存获取商品基本信息
		}

		return &gpb.GoodsInfoResponse{
			Id:          0,
			Name:        "商品暂时不可用",
			ShopPrice:   0,
			MarketPrice: 0,
			GoodsBrief:  "商品服务暂时不可用",
		}, nil

	case "/goods.v1.Goods/BatchGetGoods":
		// 批量获取商品降级
		return &gpb.GoodsListResponse{
			Total: 0,
			Data:  []*gpb.GoodsInfoResponse{},
		}, nil

	default:
		return nil, status.Error(codes.Unavailable, "商品服务暂时不可用")
	}
}

// handleInventoryFallback 处理库存服务降级
func (h *BusinessFallbackHandler) handleInventoryFallback(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, err error) (interface{}, error) {
	switch info.FullMethod {
	case "/inventory.v1.Inventory/InvDetail":
		// 库存查询降级：返回默认库存
		return &ipb.GoodsInvInfo{
			GoodsId: 0,
			Num:     0, // 库存为0，避免超卖
		}, nil

	case "/inventory.v1.Inventory/Sell":
		// 库存扣减降级：拒绝订单
		return nil, status.Error(codes.ResourceExhausted, "库存服务繁忙，请稍后重试")

	case "/inventory.v1.Inventory/Reback":
		// 库存回滚降级：记录异步处理
		log.Errorf("库存回滚服务降级，需要异步处理: req=%+v", req)
		return nil, status.Error(codes.Unavailable, "库存回滚服务暂时不可用")

	default:
		return nil, status.Error(codes.Unavailable, "库存服务暂时不可用")
	}
}

// handleOrderFallback 处理订单服务降级
func (h *BusinessFallbackHandler) handleOrderFallback(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, err error) (interface{}, error) {
	switch info.FullMethod {
	case "/order.v1.Order/CreateOrder":
		// 订单创建降级：拒绝创建新订单
		return nil, status.Error(codes.Unavailable, "下单服务暂时不可用，请稍后重试")

	case "/order.v1.Order/OrderList":
		// 订单列表降级：返回空列表
		return &opb.OrderListResponse{
			Total: 0,
			Data:  []*opb.OrderInfoResponse{},
		}, nil

	case "/order.v1.Order/OrderDetail":
		// 订单详情降级：从缓存获取或返回基本信息
		return &opb.OrderInfoResponse{
			Id:      0,
			UserId:  0,
			OrderSn: "",
			Status:  "查询失败",
			Total:   0,
		}, nil

	default:
		return nil, status.Error(codes.Unavailable, "订单服务暂时不可用")
	}
}

// handlePaymentFallback 处理支付服务降级
func (h *BusinessFallbackHandler) handlePaymentFallback(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, err error) (interface{}, error) {
	switch info.FullMethod {
	case "/payment.v1.Payment/CreatePayment":
		// 支付创建降级：暂停支付
		return nil, status.Error(codes.Unavailable, "支付服务暂时不可用，请稍后重试")

	case "/payment.v1.Payment/PaymentNotify":
		// 支付通知降级：记录到队列延后处理
		log.Errorf("支付通知服务降级，需要延后处理: req=%+v", req)
		return nil, status.Error(codes.Unavailable, "支付通知处理服务暂时不可用")

	case "/payment.v1.Payment/QueryPayment":
		// 支付查询降级：返回未知状态
		return nil, status.Error(codes.Unavailable, "支付状态查询服务暂时不可用")

	default:
		return nil, status.Error(codes.Unavailable, "支付服务暂时不可用")
	}
}

// handleDefaultFallback 处理默认降级
func (h *BusinessFallbackHandler) handleDefaultFallback(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, err error) (interface{}, error) {
	// 检查是否有预定义的默认响应
	if response, exists := h.defaultResponses[info.FullMethod]; exists {
		return response, nil
	}

	// 返回通用错误
	return nil, status.Error(codes.Unavailable, "服务暂时不可用，请稍后重试")
}

// initDefaultResponses 初始化默认响应
func (h *BusinessFallbackHandler) initDefaultResponses() {
	// 可以从配置文件或者数据库加载默认响应
	// 这里先硬编码一些通用响应
}

// SetDefaultResponse 设置默认响应
func (h *BusinessFallbackHandler) SetDefaultResponse(method string, response interface{}) {
	h.defaultResponses[method] = response
}

// CacheBasedFallbackHandler 基于缓存的降级处理器
type CacheBasedFallbackHandler struct {
	*BusinessFallbackHandler
	cacheClient interface{} // Redis client or other cache client
}

// NewCacheBasedFallbackHandler 创建基于缓存的降级处理器
func NewCacheBasedFallbackHandler(serviceName string, cacheClient interface{}) *CacheBasedFallbackHandler {
	return &CacheBasedFallbackHandler{
		BusinessFallbackHandler: NewBusinessFallbackHandler(serviceName, true),
		cacheClient:             cacheClient,
	}
}

// getFromCache 从缓存获取数据
func (h *CacheBasedFallbackHandler) getFromCache(key string) (interface{}, error) {
	// 实现缓存查询逻辑
	// 这里需要根据具体的缓存客户端实现
	return nil, nil
}

// AsyncFallbackProcessor 异步降级处理器
type AsyncFallbackProcessor struct {
	queue chan FallbackTask
}

// FallbackTask 降级任务
type FallbackTask struct {
	ServiceName string      `json:"serviceName"`
	Method      string      `json:"method"`
	Request     interface{} `json:"request"`
	Timestamp   int64       `json:"timestamp"`
	Retry       int         `json:"retry"`
}

// NewAsyncFallbackProcessor 创建异步降级处理器
func NewAsyncFallbackProcessor(bufferSize int) *AsyncFallbackProcessor {
	processor := &AsyncFallbackProcessor{
		queue: make(chan FallbackTask, bufferSize),
	}

	// 启动处理协程
	go processor.process()

	return processor
}

// AddTask 添加异步任务
func (p *AsyncFallbackProcessor) AddTask(task FallbackTask) {
	select {
	case p.queue <- task:
		log.Infof("添加异步降级任务: service=%s, method=%s", task.ServiceName, task.Method)
	default:
		log.Errorf("异步降级队列已满，丢弃任务: service=%s, method=%s", task.ServiceName, task.Method)
	}
}

// process 处理异步任务
func (p *AsyncFallbackProcessor) process() {
	for task := range p.queue {
		// 处理降级任务
		if err := p.processTask(task); err != nil {
			log.Errorf("处理异步降级任务失败: %v", err)

			// 重试逻辑
			if task.Retry < 3 {
				task.Retry++
				time.Sleep(time.Second * 2)
				select {
				case p.queue <- task:
				default:
					log.Errorf("重试任务入队失败: service=%s, method=%s", task.ServiceName, task.Method)
				}
			}
		}
	}
}

// processTask 处理单个任务
func (p *AsyncFallbackProcessor) processTask(task FallbackTask) error {
	// 根据任务类型进行处理
	log.Infof("异步处理降级任务: service=%s, method=%s, retry=%d",
		task.ServiceName, task.Method, task.Retry)

	// 这里可以实现具体的异步处理逻辑
	// 比如发送消息队列、写入数据库等

	return nil
}

// Close 关闭处理器
func (p *AsyncFallbackProcessor) Close() {
	close(p.queue)
}
