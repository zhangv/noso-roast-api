package main

import (
	"hahacking/config"
	"hahacking/database"
	"log"
	"os"
	"testing"
)

/**
# 运行所有测试
   go test ./...

   # 运行特定包的测试
   go test ./api/models/tag

   # 运行特定测试函数
   go test -run TestAddTag ./api/models/tag

   # 显示详细的测试输出
   go test -v ./...
*/

func TestMain(m *testing.M) {
	log.Println("Setting up test environment...")

	// 初始化配置
	config.Setup(".", "test")

	// 设置数据库
	database.Setup()
	dbmigrate()

	// 运行所有测试
	code := m.Run()

	// 清理工作可以放在这里
	log.Println("Cleaning up test environment...")

	// 退出并返回测试结果
	os.Exit(code)
}
