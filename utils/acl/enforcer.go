package acl

import (
	"fmt"
	"hahacking/database"
	"log"
	"strings"

	"github.com/casbin/casbin/v2"
	gormadapter "github.com/casbin/gorm-adapter/v3"
)

var Enforcer *casbin.Enforcer
var MenuEnforcer *casbin.Enforcer

func Setup() {
	db := database.GetConn()
	adapter, _ := gormadapter.NewAdapterByDB(db)
	// 通过mysql适配器新建一个enforcer
	// 注意：可能出现数据库中有脏数据导致这里初始化失败的情况
	// Enforcer, err := casbin.NewEnforcer("config/rbac_model.conf", adapter)
	Enforcer, err := casbin.NewEnforcer("config/rbac_model_with_domain.conf", adapter)
	if err != nil {
		log.Println("create enforcer err", err)
	}
	// 日志记录
	Enforcer.EnableLog(true)

	Enforcer.AddFunction("isSuper", func(arguments ...interface{}) (i interface{}, e error) {
		if len(arguments) == 1 {
			user := arguments[0].(string)
			return IsSuperAdmin(user), nil
		}
		return nil, fmt.Errorf("superMatch error")
	})
	Enforcer.AddFunction("keyMatch", KeyMatchFunc) //支持通配符 *

	MenuEnforcer, _ = casbin.NewEnforcer("config/rbac_menu_model.conf", adapter)
	MenuEnforcer.EnableLog(true)
}
func KeyMatch(key1 string, key2 string) bool { //https://www.cnblogs.com/yjf512/p/12200206.html
	i := strings.Index(key2, "*")
	if i == -1 {
		return key1 == key2
	}

	if len(key1) > i {
		return key1[:i] == key2[:i]
	}
	return key1 == key2[:i]
}

func KeyMatchFunc(args ...interface{}) (interface{}, error) {
	name1 := args[0].(string)
	name2 := args[1].(string)

	return (bool)(KeyMatch(name1, name2)), nil
}

func IsSuperAdmin(userName string) bool { //只是为了演示添加方法
	var admins = []string{"root", "admin"}
	for _, user := range admins {
		if user == userName {
			return true
		}
	}
	return false
}
