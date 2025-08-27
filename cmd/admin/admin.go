package main

import (
	"emshop/internal/app/api/admin"
	"os"
	"runtime"
)

func main() {
	if len(os.Getenv("GOMAXPROCS")) == 0 {
		runtime.GOMAXPROCS(runtime.NumCPU())
	}
	admin.NewApp("admin-server").Run()
}
