package config

import (
	"time"
)

type App struct {
	RuntimeRootPath string
	Salt            string
	LogSavePath     string
	LogSaveName     string
	LogFileExt      string
	TimeFormat      string
}

type Server struct {
	RunMode      string
	HttpPort     int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

type Database struct {
	Type        string
	User        string
	Password    string
	Host        string
	Name        string
	TablePrefix string
}

type Redis struct {
	Host        string
	Password    string
	MaxIdle     int
	MaxActive   int
	IdleTimeout time.Duration
}

type WxApp struct {
	AppId      string
	Secret     string
	MchId      string
	PrivateKey string
	NotifyUrl  string
}

type OssConfig struct {
	AccessId   string
	AccessKey  string
	Endpoint   string
	BucketName string
	ServerUrl  string
	TimeOut    int64
}
