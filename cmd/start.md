启动时需要 命令行输入 -c 参数指定config路径

go run cmd/user/user.go -c configs/user/srv.yaml
go run cmd/goods/goods.go -c configs/goods/srv.yaml

go run cmd/inventory/inventory.go -c configs/inventory/srv.yaml
go run cmd/order/order.go -c configs/order/srv.yaml


go run cmd/admin/admin.go -c configs/admin/admin.yaml
go run cmd/shop/api.go -c configs/shop/api.yaml