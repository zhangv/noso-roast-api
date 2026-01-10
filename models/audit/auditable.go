package audit

import "reflect"

type Auditable interface {
	IsAuditable()
}

func IsAuditable(dest any) bool {
	// 处理指针和非指针
	if _, ok := dest.(Auditable); ok {
		return true
	}
	// 如果是指针，尝试解引用
	val := reflect.ValueOf(dest)
	if val.Kind() == reflect.Ptr && !val.IsNil() {
		if _, ok := val.Elem().Interface().(Auditable); ok {
			return true
		}
	}
	return false
}
