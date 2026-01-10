package database

import (
	"fmt"
	"hahacking/config"
	configuration "hahacking/config"
	"log"
	"os"
	"strings"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

type Config struct {
	Driver   string // 可选: "postgres", "mysql", "sqlite"
	DSN      string // 连接字符串
	LogLevel int    // GORM 日志级别
}

var (
	Conn     *gorm.DB
	LogLevel = logger.Info // 默认日志级别
)

// SetLogLevel 设置 GORM 日志级别
func SetLogLevel(level logger.LogLevel) {
	LogLevel = level
}

func Setup() {
	if Conn == nil {
		initDB()
	}
}
func GetConn() *gorm.DB {
	return Conn
}

func initDB() (*gorm.DB, error) {
	serverSetting := config.Get("server")
	mode := serverSetting["mode"].(string)
	if mode == "prd" {
		SetLogLevel(logger.Silent)
	}
	dbConfig := configuration.GetDatabaseInfo()
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer（日志输出的目标，前缀和日志包含的内容——译者注）
		logger.Config{
			SlowThreshold:             time.Second, // 慢 SQL 阈值
			LogLevel:                  LogLevel,
			IgnoreRecordNotFoundError: true,  // 忽略ErrRecordNotFound（记录未找到）错误
			Colorful:                  false, // 禁用彩色打印
		},
	)
	// log.Println("xxxxxxxxxxxxdbConfig", dbConfig)
	var dialector gorm.Dialector
	driver := dbConfig["driver"].(string)
	// dsn := dbConfig["dsn"].(string)
	// 选择数据库驱动并初始化
	var err error
	var db *gorm.DB
	// log.Printf("db driver %v", driver)
	switch driver {
	// case "postgres":
	// 	dialector = postgres.Open(dsn)
	// 	db, err = gorm.Open(dialector, &gorm.Config{})
	case "mysql":
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			dbConfig["user"].(string),
			dbConfig["password"].(string),
			dbConfig["host"].(string),
			dbConfig["port"].(string), dbConfig["database"].(string))
		mysqlConfig := mysql.Config{
			DSN:                       dsn,
			DefaultStringSize:         256,
			DisableDatetimePrecision:  true,
			DontSupportRenameIndex:    true,
			DontSupportRenameColumn:   true,
			SkipInitializeWithVersion: false,
		}
		dialector = mysql.New(mysqlConfig)
		db, err = gorm.Open(dialector, &gorm.Config{
			NamingStrategy: schema.NamingStrategy{
				TablePrefix:   "t_",                              // 表名前缀，`User`表为`t_users`
				SingularTable: true,                              // 使用单数表名
				NameReplacer:  strings.NewReplacer("CID", "Cid"), // 在转为数据库名称之前，使用NameReplacer更改结构/字段名称。
			},
			Logger: newLogger,
			// 禁用迁移时自动创建物理外键约束
			DisableForeignKeyConstraintWhenMigrating: true,
		})
		// 获取通用数据库对象 sql.DB，然后使用其配置连接池
		sqlDB, err := db.DB()
		if err != nil {
			fmt.Printf("get sql.DB error %v", err)
		}

		// SetMaxIdleConns 设置空闲连接池中连接的最大数量
		sqlDB.SetMaxIdleConns(10)
		// SetMaxOpenConns 设置打开数据库连接的最大数量
		sqlDB.SetMaxOpenConns(100)
		// SetConnMaxLifetime 设置了连接可复用的最大时间
		sqlDB.SetConnMaxLifetime(time.Hour)
	case "sqlite":
		// log.Println("xxxxxxxxxxxxdb sqlite", dsn)
		dsn := dbConfig["dsn"].(string)
		dialector = sqlite.Open(dsn)
		db, err = gorm.Open(dialector, &gorm.Config{
			NamingStrategy: schema.NamingStrategy{
				TablePrefix:   "t_",                              // 表名前缀，`User`表为`t_users`
				SingularTable: true,                              // 使用单数表名
				NameReplacer:  strings.NewReplacer("CID", "Cid"), // 在转为数据库名称之前，使用NameReplacer更改结构/字段名称。
			},
			Logger: newLogger,
		})
		// sqlDB, err := db.DB()
		// if err != nil {
		// 	panic("failed to get sql.DB")
		// }
		// sqlDB.SetMaxOpenConns(1) // SQLite typically works best with a single connection
		// sqlDB.SetMaxIdleConns(1)
	default:
		log.Fatal("Unsupported database driver:", driver)
	}
	// 连接数据库
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	Conn = db

	return db, err
}
