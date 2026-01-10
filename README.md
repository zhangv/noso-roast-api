# 启动命令

# gin 热重载工具：-p 指定代理端口
gin -i -p 3000 run main.go
# 或者使用默认端口(3000)
# gin -i run main.go

另一个更简单的启动：
go run gin.go - 但这种做法没有hotreload

部署方式：已经编译并上传（手动），尝试用viper（参考：<https://www.loginradius.com/blog/async/environment-variables-in-golang/>）

CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o family-api main.go

# 部署

go run deploy/deploy.go

# 服务器上启动

# 服务器配置服务

创建文件：/etc/systemd/system/nosoapi.service
文件内容：
[Unit]
Description=Nosoapi Service
After=network.target

[Service]
ExecStart=/home2/noso-api/nosoapi
WorkingDirectory=/home2/noso-api
Restart=always
User=root
Environment="ENV=prd"

[Install]
WantedBy=multi-user.target

执行生效：
sudo systemctl daemon-reload
sudo systemctl enable nosoapi

sudo systemctl restart nosoapi

查看日志：journalctl -u nosoapi.service -b

# swagger 使用

## 1.在代码中写comments 2. swag init生成 3. 访问/docs/index.html

# 启用 Go Modules 功能

go env -w GO111MODULE=on

# 配置 GOPROXY 环境变量，以下三选一

## 1. 七牛 CDN

go env -w GOPROXY=<https://goproxy.cn,direct>

## 2. 阿里云

go env -w GOPROXY=<https://mirrors.aliyun.com/goproxy/,direct>

## 3. 官方

go env -w GOPROXY=<https://goproxy.io,direct>

# 线上部署

## 编译命令

CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./webapp
or：
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o ./webapp

参考：<https://juejin.cn/post/6967156604990930974>

## 上传

## 服务器启动

ENV=prd nohup ./nosoapi >> webapp.log 2>&1 &

查进程
ps -a | grep nosoapi
杀进程
kill -9 xxx

# 本地环境

ENV=dev gin -i run main.go
或者
ENV=dev gin -i run wire_gen.go

# TODO

## 陆续改造db/gorm为wire
