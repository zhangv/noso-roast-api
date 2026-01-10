# 服务器配置服务

创建文件：/etc/systemd/system/nosoapi.service
文件内容：
```ini
[Unit]
Description=Nosoapi Service
After=network.target

[Service]
ExecStart=/home2/noso-api/nosoapi
WorkingDirectory=/home2/noso-api
Restart=always
User=root
Environment="ENV=prd"
# 优雅关闭配置 - 未测试过
KillMode=mixed
KillSignal=SIGTERM
TimeoutStopSec=30
FinalKillSignal=SIGKILL
RestartSec=5

[Install]
WantedBy=multi-user.target
```
执行生效：
```shell
sudo systemctl daemon-reload
sudo systemctl enable nosoapi

sudo systemctl restart nosoapi
```
查看日志：
```shell
journalctl -u nosoapi.service -b
```