package app


type App struct {



}

func New() *App {
	return &App{}
}


// 启动整个服务
func (a *App) Run() error {
	// 注册的信息

	return nil
}


// 停止服务
func (a *App) Stop() error {
	return nil
}

// 创建服务注册的结构体
func (a *App) buildInstance() {
	// 初始化一些组件
	// 1. 初始化日志
	// 2. 初始化配置
	// 3. 初始化数据库连接
	// 4. 初始化缓存连接
	// 5. 初始化服务注册中心连接
	// 6. 初始化其他组件
}