package {{.Package}}

{{ if .Imports }}
import (
    {{ range $key, $value := .Imports }}"{{ $key }}"
    {{ end }}
)
{{ end }}

var (
    {{ .Name }}FieldNames = esql.RawFieldNames(&{{ .Name }}{})
    // 查询字段
    {{ .Name }}Fields = esql.RawQueryFields({{ .Name }}FieldNames)
    // 更新字段
    {{ .Name }}FieldsWithPlaceHolder = esql.RawUpdateFieldsWithPlaceHolder({{ .Name }}FieldNames, `id`)
)

type {{ .Name }} struct {
{{ range .Fields }} {{ .CamelName }} {{ .Type }} {{ if .HasTag }} `esql:"{{ .Name }}"` {{ end }}
{{ end }}
}

func ({{ .Name }}) TableName() string {
	return "{{ .Table }}"
}
