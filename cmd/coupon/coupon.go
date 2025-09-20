package main

import (
	srv "emshop/internal/app/coupon/srv"
	"os"
	"runtime"
)

func main() {
	if len(os.Getenv("GOMAXPROCS")) == 0 {
		runtime.GOMAXPROCS(runtime.NumCPU())
	}
	srv.NewApp("coupon-server").Run()
}
