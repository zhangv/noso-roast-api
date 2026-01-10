package audit

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"hahacking/database"
	"hahacking/utils"
	"log"
	"reflect"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

/*
TODO auditlog应该用另一个database，这样不会影响主数据库的性能，也方便定期归档
TODO 测试环境，需要把auditlog的表删除，然后重新运行测试
*/

type AuditLog struct {
	ID         uint    `gorm:"primaryKey"`
	TableName  string  `gorm:"type:varchar(255);not null"`
	ObjectName string  `gorm:"type:varchar(255);not null"`
	RecordID   uint    `gorm:"not null"`
	Action     string  `gorm:"type:varchar(50);not null"`
	OldData    *string `gorm:"type:json;null"`
	NewData    string  `gorm:"type:json"`
	ChangedBy  string  `gorm:"type:varchar(100)"`
	ChangedAt  int64   `gorm:"autoCreateTime"`
	IPAddress  string  `gorm:"type:varchar(45)"`
}

type SearchForm struct {
	utils.Pager
	TableName  string `form:"table_name"`
	ObjectName string `form:"object_name"`
	RecordID   uint   `form:"record_id"`
	Action     string `form:"action"`
}

func Search(q SearchForm) ([]AuditLog, int) {
	db := database.Conn
	var os []AuditLog
	count := int64(0)
	stmt := db.Model(&AuditLog{})
	if q.TableName != "" {
		stmt = stmt.Where("table_name = ?", q.TableName)
	}
	if q.ObjectName != "" {
		stmt = stmt.Where("object_name = ?", q.ObjectName)
	}
	if q.RecordID != 0 {
		stmt = stmt.Where("record_id = ?", q.RecordID)
	}
	if q.Action != "" {
		stmt = stmt.Where("action = ?", q.Action)
	}
	stmt.Count(&count).Find(&os)
	return os, int(count)
}

func pushToRedis(auditLog AuditLog) {
	database.Conn.Create(&auditLog) //同步处理先
}

/*
*
弃用！！由于更新时候情况比较复杂，无法拿到更新的id（主要是类似statement.Where("id = ?", id)这种），所以无法记录更新前的数据
@deprecated
*/
func RegisterAuditHooks(db *gorm.DB) {
	// 先移除已存在的钩子，避免重复注册
	db.Callback().Create().Remove("audit_create")
	db.Callback().Update().Remove("audit_update")
	db.Callback().Delete().Remove("audit_delete")

	//如果是auditlog的操作，跳过
	db.Callback().Create().After("gorm:after_create").Register("audit_create", func(db *gorm.DB) {
		if !IsAuditable(db.Statement.Dest) {
			log.Printf("object %v is not auditable", db.Statement.Dest)
			return
		}
		//获取对象名称而不是表名
		objid, objname := getRecordID(db)
		if objname == "AuditLog" {
			return
		}
		log.Printf("auditing object create name:%s, table:%s", objname, db.Statement.Table)

		auditLog := AuditLog{
			TableName:  db.Statement.Table,
			ObjectName: objname,
			RecordID:   objid,
			Action:     "CREATE",
			ChangedBy:  getUser(db),
			IPAddress:  getIP(db),
		}
		newData, _ := json.Marshal(db.Statement.Dest)
		auditLog.NewData = string(newData)
		pushToRedis(auditLog)
	})

	db.Callback().Update().Before("gorm:before_update").Register("audit_update", func(db *gorm.DB) {
		if !IsAuditable(db.Statement.Dest) {
			log.Printf("object %v is not auditable", db.Statement.Dest)
			return
		}

		objid, objname, tablename, ok := getUpdateID(db)
		log.Printf("object update id:%d, name:%s, table:%s, ok:%v", objid, objname, tablename, ok)
		if !ok {
			return
		}

		oldData, err := getRowAsMap(db, tablename, objid)
		if err != nil {
			// 处理未查到或报错
		}
		log.Println("oldData", oldData)
		auditLog := AuditLog{
			TableName:  tablename,
			ObjectName: objname,
			RecordID:   objid,
			Action:     "UPDATE",
			ChangedBy:  getUser(db),
			IPAddress:  getIP(db),
		}
		oldJSON, _ := json.Marshal(oldData)
		newData, _ := json.Marshal(db.Statement.Dest)
		oldJSONStr := string(oldJSON)
		auditLog.OldData = &oldJSONStr
		auditLog.NewData = string(newData)
		pushToRedis(auditLog)
	})

	db.Callback().Delete().Before("gorm:before_delete").Register("audit_delete", func(db *gorm.DB) {
		if !IsAuditable(db.Statement.Dest) {
			log.Printf("object %v is not auditable", db.Statement.Dest)
			return
		}

		objid, objname := getRecordID(db)
		if objname == "AuditLog" {
			return
		}
		log.Printf("auditing object create name:%s, table:%s", objname, db.Statement.Table)
		oldData := map[string]interface{}{}
		db.Statement.ConnPool.QueryRowContext(context.Background(), "SELECT * FROM "+db.Statement.Table+" WHERE id = ?", db.Statement.Dest).Scan(oldData)

		auditLog := AuditLog{
			TableName:  db.Statement.Table,
			ObjectName: objname,
			RecordID:   objid,
			Action:     "DELETE",
			ChangedBy:  getUser(db),
			IPAddress:  getIP(db),
		}
		oldJSON, _ := json.Marshal(oldData)
		oldJSONStr := string(oldJSON)
		auditLog.OldData = &oldJSONStr
		pushToRedis(auditLog)
	})
}
func getRowAsMap(db *gorm.DB, tablename string, id uint) (map[string]interface{}, error) {
	query := fmt.Sprintf("SELECT * FROM %s WHERE id = ?", tablename)
	// row := db.Statement.ConnPool.QueryRowContext(context.Background(), query, id)
	// 获取列名
	cols, err := db.Statement.ConnPool.QueryContext(context.Background(), query, id)
	if err != nil {
		return nil, err
	}
	defer cols.Close()
	columns, err := cols.Columns()
	if err != nil {
		return nil, err
	}
	values := make([]interface{}, len(columns))
	valuePtrs := make([]interface{}, len(columns))
	for i := range columns {
		valuePtrs[i] = &values[i]
	}
	if cols.Next() {
		if err := cols.Scan(valuePtrs...); err != nil {
			return nil, err
		}
		result := make(map[string]interface{})
		for i, col := range columns {
			var v interface{}
			val := values[i]
			b, ok := val.([]byte)
			if ok {
				v = string(b)
			} else {
				v = val
			}
			result[col] = v
		}
		return result, nil
	}
	return nil, sql.ErrNoRows
}

/*
TODO 这里需要优化，因为Update传入的map，无法拿到id，所以无法记录更新前的数据
*/
func getUpdateID(db *gorm.DB) (uint, string, string, bool) {
	sql := db.Dialector.Explain(db.Statement.SQL.String(), db.Statement.Vars...)
	log.Printf("即将执行的SQL: %s", sql)
	log.Printf("SQL: %s", db.Statement.SQL.String())
	log.Printf("Vars: %v", db.Statement.Vars)
	// 1. struct 类型
	if db.Statement.Schema != nil && len(db.Statement.Schema.PrimaryFields) > 0 {
		for _, pf := range db.Statement.Schema.PrimaryFields {
			val := reflect.ValueOf(db.Statement.Dest)
			if val.Kind() == reflect.Ptr {
				val = val.Elem()
			}
			if val.Kind() == reflect.Map {
				idVal := val.MapIndex(reflect.ValueOf("ID"))
				if idVal.IsValid() && idVal.Kind() == reflect.Uint {
					return uint(idVal.Uint()), "map", db.Statement.Table, true
				}
				idVal = val.MapIndex(reflect.ValueOf("id"))
				if idVal.IsValid() && idVal.Kind() == reflect.Uint {
					return uint(idVal.Uint()), "map", db.Statement.Table, true
				}
			} else {
				value, zero := pf.ValueOf(context.Background(), val)
				if !zero {
					if id, ok := value.(uint); ok {
						return id, db.Statement.Schema.Name, db.Statement.Table, true
					}
				}
			}
		}
	}
	// 2. map 类型
	val := reflect.ValueOf(db.Statement.Dest)
	if val.Kind() == reflect.Map {
		idVal := val.MapIndex(reflect.ValueOf("ID"))
		if idVal.IsValid() && idVal.Kind() == reflect.Uint {
			return uint(idVal.Uint()), "map", db.Statement.Table, true
		}
		idVal = val.MapIndex(reflect.ValueOf("id"))
		if idVal.IsValid() && idVal.Kind() == reflect.Uint {
			return uint(idVal.Uint()), "map", db.Statement.Table, true
		}
	}
	// 3. WHERE 子句
	if clauseWhere, ok := db.Statement.Clauses["WHERE"]; ok {
		if expr, ok := clauseWhere.Expression.(clause.Expr); ok {
			for _, v := range expr.Vars {
				if id, ok := v.(uint); ok {
					// 对象名无法直接获得，返回空
					return id, "", db.Statement.Table, true
				}
			}
		}
	}
	return 0, "", db.Statement.Table, false
}
func getRecordID(db *gorm.DB) (uint, string) {
	if dest, ok := db.Statement.Dest.(interface{ GetID() uint }); ok {
		log.Println("dest.GetID()", dest.GetID())
		return dest.GetID(), ""
	}
	// 反射获取主键值
	// val := reflect.ValueOf(db.Statement.Dest).Elem()
	obj := db.Statement.Dest
	log.Println("obj", obj)
	var objname string

	val := reflect.ValueOf(obj)
	//这里可能是Update传入的map，val.Kind()是map - 比如order.AfterCreate() - 修改ordersn，或者单独修改batch或者bean的库存值
	log.Println("val.Kind()", val.Kind())
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
		if val.Kind() == reflect.Ptr { //在来一次
			val = val.Elem()
		}
		log.Println("val.Kind() struct", val.Kind())
		objname = val.Type().Name()
	} else {
		objname = reflect.TypeOf(obj).Name()
	}
	idField := val.FieldByName("ID")
	if idField.IsValid() && idField.Type().Kind() == reflect.Uint {
		return uint(idField.Uint()), objname
	}
	return 0, objname
}
func getUser(db *gorm.DB) string {
	if ctx, ok := db.Statement.Context.Value("user").(string); ok {
		return ctx
	}
	return "anonymous"
}

func getIP(db *gorm.DB) string {
	if ctx, ok := db.Statement.Context.(*gin.Context); ok {
		return ctx.ClientIP()
	}
	return "unknown"
}
