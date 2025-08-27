package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"emshop/internal/app/coupon/srv"
	"emshop/pkg/log"
)

var (
	configFile = flag.String("c", "configs/coupon/srv.yaml", "配置文件路径")
	version    = flag.Bool("v", false, "显示版本信息")
)

func main() {
	flag.Parse()

	if *version {
		fmt.Println("EMShop Coupon Service v1.0.0")
		fmt.Println("基于Ristretto+Redis+MySQL三层缓存架构的高并发优惠券秒杀系统")
		os.Exit(0)
	}

	// 创建上下文
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 创建并启动应用
	couponApp, err := app.NewCouponApp(*configFile)
	if err != nil {
		log.Fatalf("创建优惠券应用失败: %v", err)
	}

	// 启动服务
	if err := couponApp.Run(ctx); err != nil {
		log.Fatalf("启动优惠券服务失败: %v", err)
	}

	// 等待信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("正在优雅关闭优惠券服务...")

	// 停止服务
	if err := couponApp.Stop(); err != nil {
		log.Errorf("停止优惠券服务失败: %v", err)
	}

	log.Info("优惠券服务已停止")
}