package main

import (
	"emshop/internal/app/user/srv"
	"os"
	"runtime"
)

func main() {
	// ​​Go 1.20 开始，math/rand包的​​全局随机数生成器​​（rand.Intn, rand.Float64等）在首次使用时​​会自动设置随机种子​​，不再需要手动调用 rand.Seed。
	// rand.Seed(time.Now().UnixNano())
	
	// 设置最大CPU数
	if len(os.Getenv("GOMAXPROCS")) == 0 {
		runtime.GOMAXPROCS(runtime.NumCPU())
	}
	srv.NewApp("user-server").Run()
}
