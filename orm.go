package esql

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

var (
	// ErrNotMatchDestination is an error that indicates not matching destination to scan.
	// (ErrNotMatchDestination 是一个错误，表示不匹配要扫描的目标。)
	ErrNotMatchDestination = errors.New("not matching destination to scan")
	// ErrNotReadableValue is an error that indicates value is not addressable or interfaceable.
	// (ErrNotReadableValue 是一个错误，值不是指针或接口。)
	ErrNotReadableValue = errors.New("value not addressable or interfaceable")
	// ErrNotSettable is an error that indicates the passed in variable is not settable.
	// (ErrNotSettable 是一个错误，表示传入的变量不可设置。)
	ErrNotSettable = errors.New("passed in variable is not settable")
	// ErrUnsupportedValueType is an error that indicates unsupported unmarshal type.
	// (ErrUnsupportedValueType 是一个错误，表示不是支持反序列化的类型。)
	ErrUnsupportedValueType = errors.New("unsupported unmarshal type")
	// ErrRecordNotFound is an error that record not found
	// (rrRecordNotFound 是一个错误，表示未找到记录)
	ErrRecordNotFound = errors.New("record not found")
)

const tagName = "esql"

// sql.Rows默认已实现
type rowsScanner interface {
	Columns() ([]string, error)
	Err() error
	Next() bool
	Scan(v ...interface{}) error
}

// 通过tag获取字段名称和字段值
func getTaggedFieldValueMap(v reflect.Value) (map[string]interface{}, error) {
	rt := Deref(v.Type())
	size := rt.NumField()
	result := make(map[string]interface{}, size)

	for i := 0; i < size; i++ {
		field := rt.Field(i)
		// 判断是否是嵌入的结构体
		if field.Anonymous && Deref(field.Type).Kind() == reflect.Struct {
			inner, err := getTaggedFieldValueMap(reflect.Indirect(v).Field(i))
			if err != nil {
				return nil, err
			}

			for key, val := range inner {
				result[key] = val
			}

			continue
		}

		key := parseTagName(field)
		if len(key) == 0 {
			// 没标签，默认字段名下划线格式
			key = ConvertCamelToSnake(field.Name)
		}

		valueField := reflect.Indirect(v).Field(i)
		valueData, err := getValueInterface(valueField)
		if err != nil {
			return nil, err
		}

		result[key] = valueData
	}

	return result, nil
}

func getValueInterface(value reflect.Value) (interface{}, error) {
	switch value.Kind() {
	case reflect.Ptr:
		if !value.CanInterface() {
			return nil, ErrNotReadableValue
		}

		if value.IsNil() {
			// 返回value的实际类型
			baseValueType := Deref(value.Type())
			// 新建一个指向和value相同类型的新零值的指针，赋值给value
			value.Set(reflect.New(baseValueType))
		}

		return value.Interface(), nil
	default:
		if !value.CanAddr() || !value.Addr().CanInterface() {
			return nil, ErrNotReadableValue
		}
		// 创建空值返回
		return value.Addr().Interface(), nil
	}
}

// 将结构体字段类型变量映射到切片
func mapStructFieldsIntoSlice(v reflect.Value, columns []string, strict bool) ([]interface{}, error) {
	fields := unwrapFields(v)
	if strict && len(columns) < len(fields) {
		return nil, ErrNotMatchDestination
	}

	taggedMap, err := getTaggedFieldValueMap(v)
	if err != nil {
		return nil, err
	}

	values := make([]interface{}, len(columns))
	if len(taggedMap) == 0 {
		if len(fields) < len(values) {
			return nil, ErrNotMatchDestination
		}

		for i := 0; i < len(values); i++ {
			valueField := fields[i]
			valueData, err := getValueInterface(valueField)
			if err != nil {
				return nil, err
			}

			values[i] = valueData
		}
	} else {
		for i, column := range columns {
			if tagged, ok := taggedMap[column]; ok {
				values[i] = tagged
			} else {
				var anonymous interface{}
				values[i] = &anonymous
			}
		}
	}

	return values, nil
}

func parseTagName(field reflect.StructField) string {
	key := field.Tag.Get(tagName)
	if len(key) == 0 {
		return ""
	}

	options := strings.Split(key, ",")
	return strings.TrimSpace(options[0])
}

// 把单条数据scan到v
func unmarshalRow(v interface{}, scanner rowsScanner, strict bool) error {
	if !scanner.Next() {
		if err := scanner.Err(); err != nil {
			return err
		}
		return ErrRecordNotFound
	}

	rv := reflect.ValueOf(v)
	// 判断是否是指针
	if err := ValidatePtr(rv); err != nil {
		return err
	}

	// 获取具体类型
	rte := reflect.TypeOf(v).Elem()
	rve := rv.Elem()
	switch rte.Kind() {
	case reflect.Bool,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64,
		reflect.String:
		if !rve.CanSet() {
			return ErrNotSettable
		}

		return scanner.Scan(v)
	case reflect.Struct:
		columns, err := scanner.Columns()
		if err != nil {
			return err
		}

		values, err := mapStructFieldsIntoSlice(rve, columns, strict)
		if err != nil {
			return err
		}
		// 扫描到结构体的每个字段值
		return scanner.Scan(values...)
	default:
		return ErrUnsupportedValueType
	}
}

// 把多条数据scan到v
func unmarshalRows(v interface{}, scanner rowsScanner, strict bool) error {
	rv := reflect.ValueOf(v)
	if err := ValidatePtr(rv); err != nil {
		return err
	}

	rt := reflect.TypeOf(v)
	rte := rt.Elem()
	rve := rv.Elem()
	if !rve.CanSet() {
		return ErrNotSettable
	}

	switch rte.Kind() {
	case reflect.Slice:
		ptr := rte.Elem().Kind() == reflect.Ptr
		appendFn := func(item reflect.Value) {
			if ptr {
				// 将item添加到切片rve，并重新赋值给rev
				rve.Set(reflect.Append(rve, item))
			} else {
				// 如果元素不是指针则创建一个指向元素的指针并添加到切片rve，并重新赋值给rev
				rve.Set(reflect.Append(rve, reflect.Indirect(item)))
			}
		}
		fillFn := func(value interface{}) error {
			if err := scanner.Scan(value); err != nil {
				return err
			}

			appendFn(reflect.ValueOf(value))
			return nil
		}

		base := Deref(rte.Elem())
		switch base.Kind() {
		case reflect.Bool,
			reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
			reflect.Float32, reflect.Float64,
			reflect.String:
			for scanner.Next() {
				value := reflect.New(base)
				if err := fillFn(value.Interface()); err != nil {
					return err
				}
			}
		case reflect.Struct:
			columns, err := scanner.Columns()
			if err != nil {
				return err
			}

			for scanner.Next() {
				value := reflect.New(base)
				values, err := mapStructFieldsIntoSlice(value, columns, strict)
				if err != nil {
					return err
				}

				if err := scanner.Scan(values...); err != nil {
					return err
				}

				appendFn(value)
			}
		default:
			return ErrUnsupportedValueType
		}

		return nil
	default:
		return ErrUnsupportedValueType
	}
}

// 解析v，返回其所有字段
func unwrapFields(v reflect.Value) []reflect.Value {
	var fields []reflect.Value
	indirect := reflect.Indirect(v)

	for i := 0; i < indirect.NumField(); i++ {
		child := indirect.Field(i)
		if !child.CanSet() {
			continue
		}

		if child.Kind() == reflect.Ptr && child.IsNil() {
			baseValueType := Deref(child.Type())
			child.Set(reflect.New(baseValueType))
		}

		child = reflect.Indirect(child)
		childType := indirect.Type().Field(i)
		// 如果tag值为"-"则忽略
		tagValue := childType.Tag.Get(tagName)
		if tagValue == "-" {
			continue
		}
		if child.Kind() == reflect.Struct && childType.Anonymous {
			fields = append(fields, unwrapFields(child)...)
		} else {
			fields = append(fields, child)
		}
	}

	return fields
}

// Deref 取消引用类型，如果是指针类型，则返回其元素类型
func Deref(t reflect.Type) reflect.Type {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	return t
}

// ValidatePtr 验证v是否是有效指针
func ValidatePtr(v reflect.Value) error {
	// 必须先检测是否是指针，再判断是不是空指针
	if !v.IsValid() || v.Kind() != reflect.Ptr || v.IsNil() {
		return fmt.Errorf("not a valid pointer: %v", v)
	}

	return nil
}
