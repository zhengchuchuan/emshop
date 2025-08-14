package main

import (
	"emshop/internal/app/userop/srv"
	"os"
	"runtime"
)

func main() {
	// 设置最大CPU数
	if len(os.Getenv("GOMAXPROCS")) == 0 {
		runtime.GOMAXPROCS(runtime.NumCPU())
	}
	srv.NewApp("userop-server").Run()
}