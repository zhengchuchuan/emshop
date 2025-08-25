package main

import (
	"emshop/internal/app/payment/srv"
	"os"
	"runtime"
)

func main() {
	if len(os.Getenv("GOMAXPROCS")) == 0 {
		runtime.GOMAXPROCS(runtime.NumCPU())
	}
	srv.NewApp("payment-server").Run()
}