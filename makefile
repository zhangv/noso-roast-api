APP_NAME=nosoapi
SERVER_USER=root
SERVER_IP=api.noso.cafe
REMOTE_PATH=/home2/noso-api

dev:
	ENV=dev gin -i run main.go

build:
	@echo "🔨 编译 Go 应用..."
	GOOS=linux GOARCH=amd64 go build -o $(APP_NAME)

deploy: build
	@echo "📤 上传到服务器..."
	scp $(APP_NAME) $(SERVER_USER)@$(SERVER_IP):$(REMOTE_PATH)/

	@echo "🔄 重启服务器上的服务..."
	ssh $(SERVER_USER)@$(SERVER_IP) "sudo systemctl restart $(APP_NAME) && sudo systemctl status $(APP_NAME) --no-pager"

clean:
	@echo "🧹 清理编译文件..."
	rm -f $(APP_NAME)
