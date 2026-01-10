package main

import (
	"context"
	"flag"
	"fmt"
	"hahacking/common/gredis"
	"hahacking/common/job"
	"hahacking/common/logger"
	"hahacking/common/redis"
	"hahacking/config"
	"hahacking/database"
	"hahacking/i18n"
	"hahacking/models/audit"
	"hahacking/models/bean"
	"hahacking/models/cupping"
	"hahacking/models/favorite"
	"hahacking/models/order"
	"hahacking/models/page"
	"hahacking/models/price"
	"hahacking/models/report"
	"hahacking/models/roast"
	"hahacking/models/shop"
	"hahacking/models/supplier"
	"hahacking/models/tag"
	"hahacking/models/user"
	"hahacking/queue"
	"hahacking/routers"
	"hahacking/scheduler"
	"hahacking/utils/acl"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fvbock/endless"
	"github.com/gin-gonic/gin"
	gormLogger "gorm.io/gorm/logger"

	//testing

	_ "github.com/go-sql-driver/mysql"
)

func Init() { //手动调用
	env := config.ENV_DEV
	if os.Getenv("ENV") != "" { //系统环境变量优先
		env = os.Getenv("ENV")
	}
	databaseLogLevel := gormLogger.Error
	//根据环境变量设置数据库日志级别
	if env == config.ENV_PRD {

	} else {
		databaseLogLevel = gormLogger.Info
	}

	configBasePath := "."
	config.Setup(configBasePath, env)
	logger.Setup()
	logger.Infof("logger setup env: %s", env)

	gredis.Setup()
	redis.Setup()
	database.SetLogLevel(databaseLogLevel)
	database.Setup()

	if env == config.ENV_PRD {
		audit.RegisterAuditHooks(database.Conn)
	}
	go queue.Setup()
	// _, cancelQueue := context.WithCancel(queueCtx)

	acl.Setup()
	job.Setup()
	i18n.Init()
	scheduler.Init()
	// time.LoadLocation("Asia/Shanghai")
	dbmigrate()
}
func dbmigrate() {
	db := database.Conn
	err := db.AutoMigrate(
		&tag.Tag{}, &tag.TagGroup{}, &tag.Tagging{},
		&order.Order{}, &order.OrderItem{}, &order.OrderPayment{}, &order.OrderPaymentItem{},
		&bean.Bean{}, &bean.BeanBatch{}, &bean.BeanBatchLog{}, &bean.BeanStockLog{},
		&roast.Roast{}, &roast.RoastItem{}, &roast.RoastProfile{}, &roast.RoastRecipe{},
		&price.Price{}, &price.PriceLevel{},
		&user.User{},
		&shop.Shop{}, &shop.ShopAccount{}, &shop.ShopAccountLog{},
		&cupping.Cupping{}, &cupping.CuppingSample{}, &cupping.CuppingScore{},
		&supplier.Supplier{},
		&page.Page{},
		&report.Report{}, &report.Column{}, &report.Parameter{}, &report.Chart{},
		&favorite.Favorite{},
	)
	if err != nil {
		log.Fatalf("AutoMigrate failed: %v", err)
	}
}
func main() { //原生版本
	Init()
	flag.Parse()
	// 初始化 Gin
	serverSetting := config.Get("server")
	mode := serverSetting["mode"].(string)
	gin.SetMode(mode)
	port := serverSetting["port"].(float64)
	log.Println("mode:", mode, " port:", port)

	r := routers.InitRouter()
	// 启动服务
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", int(port)),
		Handler: r,
	}

	go func() { //协程监听退出信号
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
		<-quit
		log.Println("Shutting down server...")

		queue.Cancel() // 取消队列

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			log.Fatalf("Server forced to shutdown: %v", err)
		}

		// if *isTest {
		// 	// 测试环境热加载
		// 	log.Println("Hot reload triggered...")
		// 	exe, _ := os.Executable()
		// 	args := append([]string{exe}, os.Args[1:]...)
		// 	if err := exec.Command(exe, args...).Start(); err != nil {
		// 		log.Fatalf("Failed to restart: %v", err)
		// 	}
		// }
		log.Println("Server exiting")
	}()

	// 生产环境简单启动
	// if *isTest {
	// 	log.Println("Running in test mode with hot reload enabled")
	// } else {
	// 	log.Println("Running in production mode")
	// }

	// 启动服务
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Listen: %v", err)
	}
}

// func main1() { //wire版本
// 	server, cleanup, err := setup(context.Background())
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	defer cleanup()

//		log.Println("start server and listen: ", server.Addr)
//		err = server.ListenAndServe()
//		if err != nil {
//			log.Fatal(err)
//		}
//	}
func main1() { //endless版
	Init()
	defer scheduler.S.Stop()
	gin.SetMode(config.C.ServerSetting.RunMode)
	// gin.Default()
	endless.DefaultReadTimeOut = config.C.ServerSetting.ReadTimeout
	endless.DefaultWriteTimeOut = config.C.ServerSetting.WriteTimeout
	endless.DefaultMaxHeaderBytes = 1 << 20
	endless.DefaultHammerTime = 5 * time.Second // 设置优雅关闭超时

	endPoint := fmt.Sprintf(":%d", config.C.ServerSetting.HttpPort)

	server := endless.NewServer(endPoint, routers.InitRouter())
	server.ReadTimeout = 10 * time.Second
	server.WriteTimeout = 10 * time.Second

	server.BeforeBegin = func(add string) {
		pid := syscall.Getpid()
		log.Printf("Actual pid is %d", pid)
		f, err := os.Create("./webapp.pid")

		if err != nil {
			fmt.Println(err.Error())
		} else {
			_, err = f.Write([]byte(fmt.Sprint(pid)))
			defer f.Close()
		}
	}
	// server.SignalHooks[endless.PRE_SIGNAL][syscall.SIGHUP] = []func(){
	// 	func() {
	// 		log.Println("重启开始...")
	// 	},
	// }
	// server.SignalHooks[endless.POST_SIGNAL][syscall.SIGHUP] = []func() {
	// 	func() {
	// 		log.Println("SIGHUP 处理完成")
	// 	},
	// }
	err := server.ListenAndServe()
	if err != nil {
		log.Printf("Server err: %v", err)
	}

}

// var (
// 	isTest = flag.Bool("test", false, "Run in test mode with hot reload")
// )

func NewServer(config *config.Config) *http.Server {
	gin.SetMode(config.ServerSetting.RunMode)
	endless.DefaultReadTimeOut = config.ServerSetting.ReadTimeout
	endless.DefaultWriteTimeOut = config.ServerSetting.WriteTimeout
	endless.DefaultMaxHeaderBytes = 1 << 20
	endPoint := fmt.Sprintf(":%d", config.ServerSetting.HttpPort)

	server := endless.NewServer(endPoint, routers.InitRouter())
	server.BeforeBegin = func(add string) {
		pid := syscall.Getpid()
		// log.Printf("Actual pid is %d", pid)
		f, err := os.Create("./main.pid")

		if err != nil {
			fmt.Println(err.Error())
		} else {
			_, err = f.Write([]byte(fmt.Sprint(pid)))
			log.Println(err)
			defer f.Close()

		}
	}
	return &server.Server
}

func main0() { //不使用endless的方法
	gin.SetMode(config.C.ServerSetting.RunMode)
	//log.Println(setting.ServerSetting.RunMode)
	routersInit := routers.InitRouter()
	readTimeout := config.C.ServerSetting.ReadTimeout
	writeTimeout := config.C.ServerSetting.WriteTimeout
	endPoint := fmt.Sprintf(":%d", config.C.ServerSetting.HttpPort)
	maxHeaderBytes := 1 << 20

	server := &http.Server{
		Addr:           endPoint,
		Handler:        routersInit,
		ReadTimeout:    readTimeout * time.Second,
		WriteTimeout:   writeTimeout * time.Second,
		MaxHeaderBytes: maxHeaderBytes,
	}

	log.Printf("[info] start http server listening %s", endPoint)

	go func() {
		// service connections
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	<-quit
	log.Println("Shutdown Server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatal("Server Shutdown:", err)
	}
	log.Println("Server exiting")
}
