package esql

import (
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"text/template"
)

const (
	Mysql    = "mysql"
	Postgres = "postgres"
	SQLite   = "sqlite3"
)

type Table struct {
	Name string `esql:"table_name"`
}

type TableField struct {
	Name string `esql:"Field"`
	Type string `esql:"Type"`
}

type Field struct {
	Name      string
	CamelName string
	Type      string
	HasTag    bool
}

type StructInfo struct {
	Table   string
	Package string
	Name    string
	Fields  []Field
	Imports map[string]struct{}
}

//go:embed struct.tpl
var structTemplate string

func GenStructByTable(mode, dsn, dbName, savePath string, hasTag bool) error {
	db, err := Open(mode, dsn, nil)
	if err != nil {
		return err
	}

	err = db.Ping()
	if err != nil {
		return err
	}

	return db.GenStructByTable(mode, dbName, savePath, hasTag)
}

func genStructByMysqlTable(db *DB, dbName, savePath string, hasTag bool) error {
	var tables []Table
	query := "select table_name from information_schema.tables where table_schema=?"
	err := db.QueryRows(&tables, query, dbName)
	if err != nil {
		return err
	}

	if len(tables) > 0 {
		if savePath == "./" || savePath == "." || savePath == "" {
			savePath, _ = os.Getwd()
			savePath = strings.ReplaceAll(savePath, "\\", "/")
		}

		dirs := strings.Split(savePath, "/")
		pack := dirs[len(dirs)-1]
		errCh := make(chan error, len(tables))
		wg := &sync.WaitGroup{}

		for _, table := range tables {
			wg.Add(1)
			go func(table string) {
				defer wg.Done()
				errCh <- genFileByMysqlTable(db, table, savePath, pack, hasTag)
			}(table.Name)
		}

		wg.Wait()
		// 统一格式化
		cmd := exec.Command("go", "fmt", savePath)
		cmd.Run()

		close(errCh)
		for err := range errCh {
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// Generate model files by Mysql table (通过Mysql表生成模型文件)
func genFileByMysqlTable(db *DB, table, savePath, pack string, hasTag bool) error {
	var fs []TableField
	err := db.QueryRows(&fs, fmt.Sprintf("show columns from %s", table))
	if err != nil {
		return err
	}
	if len(fs) > 0 {
		structInfo := StructInfo{
			Table:   table,
			Package: pack,
			Name:    ConvertToCamel(table),
			Fields:  make([]Field, 0, len(fs)),
			Imports: make(map[string]struct{}),
		}

		structInfo.Imports["github.com/cyj19/esql"] = struct{}{}

		for _, v := range fs {
			field := Field{Name: v.Name, CamelName: ConvertToCamel(v.Name), HasTag: hasTag}
			ot := v.Type
			v.Type = strings.Split(v.Type, "(")[0]
			switch v.Type {
			case "int", "tinyint", "smallint":
				field.Type = "int"
				if strings.Contains(ot, "unsigned") {
					field.Type = "uint"
				}
			case "bigint":
				field.Type = "int64"
				if strings.Contains(ot, "unsigned") {
					field.Type = "uint64"
				}
			case "char", "varchar", "longtext", "text", "tinytext":
				field.Type = "string"
			case "date", "datetime", "timestamp":
				field.Type = "time.Time"
				structInfo.Imports["time"] = struct{}{}
			case "double", "float":
				field.Type = "float64"
			default:
				// 其他类型当成string处理
				field.Type = "string"
			}

			structInfo.Fields = append(structInfo.Fields, field)
		}

		// 解析模板
		tmpl, err := template.New("structTemplate").Parse(structTemplate)
		if err != nil {
			return err
		}

		// 表名作为文件名
		filename := savePath + "/" + table + ".go"
		f, err := os.Create(filename)
		if err != nil {
			return err
		}
		defer f.Close()

		return tmpl.Execute(f, structInfo)
	}

	return ErrRecordNotFound
}

func genStructByPostgresSqlTable(db *DB, dbName, savePath string, hasTag bool) error {
	var tables []Table
	query := "select tablename as table_name from pg_tables where schemaname='public'"
	err := db.QueryRows(&tables, query)
	if err != nil {
		return err
	}

	//TODO

	return nil
}

func genStructBySQLiteTable(db *DB, dbName, savePath string, hasTag bool) error {
	//TODO

	return nil
}
