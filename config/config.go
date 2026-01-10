package config

import (
	"flag"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"gopkg.in/ini.v1"
)

type Config struct {
	BasePath        string
	AppSetting      *App
	ServerSetting   *Server
	DatabaseSetting *Database
	RedisSetting    *Redis
	WxappSetting    *WxApp
	OssSetting      *OssConfig
}

var C Config

// viperConfig 全局配置变量
var viperConfig *viper.Viper
var V *viper.Viper
var ENV string

// setupOnce 确保 Setup 只被调用一次
var setupOnce sync.Once

const (
	ENV_DEV  = "dev"
	ENV_TEST = "test"
	ENV_PRD  = "prd"
)

func Setup(basePath string, environment string) (any, error) {
	var result any
	var setupErr error

	setupOnce.Do(func() {
		ENV = environment

		// 只在第一次调用时定义 flag，避免重复定义
		if !flag.Parsed() {
			flag.String("env", "dev", "可选的参数为:dev,test,prod")
			pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
			pflag.Parse()
			viper.BindPFlags(pflag.CommandLine)
		}

		//path , _ := os.Getwd()

		result = InitViper(environment, basePath)

		C.BasePath = basePath
		C.AppSetting = &App{}
		C.ServerSetting = &Server{}
		C.DatabaseSetting = &Database{}
		C.OssSetting = &OssConfig{}
		C.WxappSetting = &WxApp{}
		C.RedisSetting = &Redis{}
		initializeSetting(basePath) //这里放了logger的配置，后续移出到config里
	})

	return result, setupErr
}

func Get(name string) map[string]any {
	return viper.GetStringMap(name)
}

func InitViper(env string, basePath string) interface{} {
	viper.AddConfigPath(basePath)             //执行test会用到这里
	viper.AddConfigPath(basePath + "/config") //gin启动会读这里
	viper.SetConfigType("json")

	// 1. 先读入基础配置
	viper.SetConfigName("config/env.json")
	if err := viper.ReadInConfig(); err != nil {
		panic(err)
	}

	// 2. 读入环境覆盖配置
	viper.SetConfigName(fmt.Sprintf("config/env_%s.json", env))
	if err := viper.MergeInConfig(); err != nil {
		panic(err)
	}

	// viper.SetConfigName("env_" + env)
	// error := viper.ReadInConfig()
	// if error != nil {
	// 	panic(error)
	// }

	//监听配置文件修改
	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		// 配置文件发生变更之后会调用的回调函数
		// logger.Warnf("Config file changed: %s", e.Name)
	})
	viperConfig = viper.GetViper()
	c := viper.AllSettings() //获取所有配置
	return c
}

// GetConfig 获取全局配置
func GetConfig() *viper.Viper {
	return viperConfig
}

// 获取数据库配置信息
func GetDatabaseInfo() map[string]interface{} {
	return viper.GetStringMap("db")
}

// 获取数据库配置信息
func GetDatabaseInfo2() map[string]interface{} {
	return viper.GetStringMap("db2")
}

// 获取环境变量
func GetEnvInfo(env string) string {
	viper.AutomaticEnv()
	return viper.GetString(env)
}

var cfg *ini.File

// Setup initialize the configuration instance
func initializeSetting(base string) {
	var err error
	if base == "" {
		base = "."
	}
	cfg, err = ini.Load(base + "/config/app.ini")
	if err != nil {
		log.Fatalf("setting.Setup, fail to parse 'conf/app.ini': %v", err)
	}

	mapTo("app", C.AppSetting)
	mapTo("server", C.ServerSetting)
	mapTo("database", C.DatabaseSetting)
	mapTo("redis", C.RedisSetting)
	mapTo("wxapp", C.WxappSetting)
	mapTo("oss", C.OssSetting)

	C.ServerSetting.ReadTimeout = C.ServerSetting.ReadTimeout * time.Second
	C.ServerSetting.WriteTimeout = C.ServerSetting.WriteTimeout * time.Second
	C.RedisSetting.IdleTimeout = C.RedisSetting.IdleTimeout * time.Second
}

// mapTo map section
func mapTo(section string, v interface{}) {
	err := cfg.Section(section).MapTo(v)
	if err != nil {
		log.Fatalf("Cfg.MapTo %s err: %v", section, err)
	}
}

// IsTesting 检测是否在测试环境中运行
// 通过检查 Go 测试框架的标志来判断是否在测试环境中运行
func IsTesting() bool {
	return flag.Lookup("test.v") != nil
}
