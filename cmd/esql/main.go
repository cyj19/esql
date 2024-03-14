package main

import (
	"flag"
	"fmt"
	"github.com/cyj19/esql"
	_ "github.com/go-sql-driver/mysql"
	"log"
)

var (
	mode     = flag.String("mode", esql.Mysql, "the database drive")
	ip       = flag.String("ip", "127.0.0.1", "the database ip")
	port     = flag.Int("port", 3306, "the database port")
	user     = flag.String("u", "root", "the database user")
	password = flag.String("p", "", "the database password")
	dsn      = flag.String("dsn", "", "the dataSource")
	dbName   = flag.String("db", "", "the database name")
	tag      = flag.Bool("tag", false, "the generated structure needs to be tagged")
	savePath = flag.String("path", "./", "the path to save file")
)

func main() {

	flag.Parse()

	// 如果没有提供参数，则打印帮助信息
	if *mode == "" || *dbName == "" {
		flag.Usage()
		return
	}

	var dataSource string
	if *dsn != "" {
		dataSource = *dsn
	} else {
		switch *mode {
		case esql.Mysql:
			dataSource = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8", *user, *password, *ip, *port, *dbName)
		case esql.Postgres:
			dataSource = fmt.Sprintf("user=%s password=%s dbname=%s port=%s sslmode=disable", *user, *password, *dbName, *port)
		default:
			//TODO
		}

	}

	err := esql.GenStructByTable(*mode, dataSource, *dbName, *savePath, *tag)
	if err != nil {
		log.Fatal(err)
	}
}
