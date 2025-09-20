package main

import (
	emshop "emshop/internal/app/api/emshop"
	"os"
	"runtime"
)

func main() {
	// rand.Seed(time.Now().UnixNano())
	if len(os.Getenv("GOMAXPROCS")) == 0 {
		runtime.GOMAXPROCS(runtime.NumCPU())
	}
	emshop.NewApp("api-server").Run()
}
