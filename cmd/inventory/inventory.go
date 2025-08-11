package main

import (
	"emshop/internal/app/inventory/srv"
	"os"
	"runtime"
)

func main() {
	if len(os.Getenv("GOMAXPROCS")) == 0 {
		runtime.GOMAXPROCS(runtime.NumCPU())
	}
	srv.NewApp("inventory-server").Run()
}
