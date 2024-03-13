package esql

import (
	"fmt"
	"reflect"
	"strings"
	"unicode"
)

// 获取结构体中的字段，只接受结构体/指针
func RawFieldNames(in interface{}, postgreSql ...bool) []string {
	out := make([]string, 0)
	v := reflect.ValueOf(in)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	var pg bool
	if len(postgreSql) > 0 {
		pg = postgreSql[0]
	}

	if v.Kind() != reflect.Struct {
		panic(fmt.Errorf("only accepts structs; got %T", v))
	}

	typ := v.Type()
	for i := 0; i < v.NumField(); i++ {
		// gets a StructField
		fi := typ.Field(i)
		vi := v.Field(i)
		if fi.Type.Kind() == reflect.Ptr {
			vi = vi.Elem()
		}
		if vi.Kind() == reflect.Struct {
			for j := 0; j < vi.NumField(); j++ {
				typJ := vi.Type()
				fj := typJ.Field(j)
				tagv := fj.Tag.Get(tagName)
				switch tagv {
				case "-":
					continue
				case "":
					// 默认驼峰字段名转为下划线格式
					tagv = ConvertCamelToSnake(fi.Name)
					if pg {
						out = append(out, tagv)
					} else {
						out = append(out, fmt.Sprintf("`%s`", tagv))
					}
				default:
					if len(tagv) == 0 {
						tagv = ConvertCamelToSnake(fi.Name)
					}
					if pg {
						out = append(out, tagv)
					} else {
						out = append(out, fmt.Sprintf("`%s`", tagv))
					}
				}
			}

			continue
		}

		tagv := fi.Tag.Get(tagName)
		switch tagv {
		case "-":
			continue
		case "":
			// 默认驼峰字段名转为下划线格式
			tagv = ConvertCamelToSnake(fi.Name)
			if pg {
				out = append(out, tagv)
			} else {
				out = append(out, fmt.Sprintf("`%s`", tagv))
			}
		default:
			if len(tagv) == 0 {
				tagv = ConvertCamelToSnake(fi.Name)
			}
			if pg {
				out = append(out, tagv)
			} else {
				out = append(out, fmt.Sprintf("`%s`", tagv))
			}
		}
	}

	return out
}

// 移除字段
func RemoveFieldName(strings []string, strs ...string) []string {
	out := append([]string(nil), strings...)

	for _, str := range strs {
		var n int
		for _, v := range out {
			if v != str {
				out[n] = v
				n++
			}
		}
		out = out[:n]
	}

	return out
}

// 为字段添加表名前缀
func AddFieldPrefix(strs []string, prefix string) []string {
	out := make([]string, 0, len(strs))
	for i := range strs {
		out = append(out, prefix+"."+strs[i])
	}

	return out
}

// 驼峰转下划线格式
func ConvertCamelToSnake(s string) string {
	var result string
	var words []string
	lastIndex := 0
	rs := []rune(s)

	for i := 0; i < len(rs); i++ {
		if i > 0 && unicode.IsUpper(rs[i]) {
			words = append(words, string(rs[lastIndex:i]))
			lastIndex = i
		}
	}

	words = append(words, string(rs[lastIndex:]))

	for k, word := range words {
		if k > 0 {
			result += "_"
		}

		result += strings.ToLower(word)
	}

	return result
}

// 大写驼峰格式
func ConvertToCamel(s string) string {
	temp := strings.Split(s, "_") // 有下划线的，需要拆分
	var str string
	for i := 0; i < len(temp); i++ {
		b := []rune(temp[i])
		for j := 0; j < len(b); j++ {
			if j == 0 {
				// 首字母大写转换
				b[j] -= 32
				str += string(b[j])
			} else {
				str += string(b[j])
			}
		}
	}
	return str
}

// 获取查询字段
func RawQueryFields(fieldNames []string) string {
	return strings.Join(fieldNames, ",")
}

// 获取带表名前缀的查询字段
func RawQueryFieldsWithPrefix(fieldNames []string, table string) string {
	return strings.Join(AddFieldPrefix(fieldNames, table), ",")
}

// 获取带占位符的更新字段
func RawUpdateFieldsWithPlaceHolder(fieldNames []string, str ...string) string {
	return strings.Join(RemoveFieldName(fieldNames, str...), "=?,") + "=?"
}
