package db

import (
	"fmt"
	"strings"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"

	_ "github.com/go-sql-driver/mysql"
)

var (
	Mysql *gorm.DB
)

func init() {
	var err error
	dsn := "root:root@(127.0.0.1:3306)/biasurago?charset=utf8&parseTime=True&loc=Local"
	Mysql, err = gorm.Open(mysql.New(mysql.Config{
		DSN:                       dsn,   // "root:root@tcp(127.0.0.1:3306)/family?charset=utf8&parseTime=True&loc=Local", // DSN data source name
		DefaultStringSize:         256,   // string 类型字段的默认长度
		DisableDatetimePrecision:  true,  // 禁用 datetime 精度，MySQL 5.6 之前的数据库不支持
		DontSupportRenameIndex:    true,  // 重命名索引时采用删除并新建的方式，MySQL 5.7 之前的数据库和 MariaDB 不支持重命名索引
		DontSupportRenameColumn:   true,  // 用 `change` 重命名列，MySQL 8 之前的数据库和 MariaDB 不支持重命名列
		SkipInitializeWithVersion: false, // 根据当前 MySQL 版本自动配置
	}), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   "t_",                              // 表名前缀，`User`表为`t_users`
			SingularTable: true,                              // 使用单数表名，启用该选项后，`User` 表将是`user`
			NameReplacer:  strings.NewReplacer("CID", "Cid"), // 在转为数据库名称之前，使用NameReplacer更改结构/字段名称。
		},
	})
	if err != nil {
		fmt.Println("connect DB error")
		panic(err)
	}
}
