# esql

`esql`意为`easy sql`，一个轻量的`sql`库，定位不是传统的`ORM`框架（`GORM`已经足够全面了），目的是在标准库`database/sql` 
基础上提供对象映射等功能补充。因为`database/sql`的学习成本已经非常低，通常只是无法进行对象映射而显得繁琐。
## 主要功能 
- 对象映射
- 自动化事务
- 通过命令行/函数调用，生成表对应的模型文件（暂时只支持MySQL）
- 通过结构体获取查询字段和更新字段
- 开发日志接口，自定义日志输出

## 安装
- 安装命令行工具
```bash
go install github.com/cyj19/esql/cmd/esql@latest
```
- 安装`esql`库
```bash
go get -u github.com/cyj19/esql@latest
```

## 使用方法
- 连接
```
dataSource := "root:123456@tcp(127.0.0.1:3306)/test?charset=utf8&parseTime=True&loc=Local"
db, err := esql.Open(esql.Mysql, dataSource, nil)
if err != nil {
    log.Fatal(err)
}

err = db.Ping()
if err != nil {
    log.Fatal(err)
}
```
- 查询单条记录
```
userFieldNames := esql.RawFieldNames(User{})
userFields := esql.RawQueryFields(userFieldNames)
var user User
query := fmt.Sprintf("select %s from user where id=?", userFields)
err := db.QueryRow(&user, query, 2)
if err != nil {
    log.Fatal(err)
}

log.Println(user)
```
- 查询多条记录
```
userFieldNames := esql.RawFieldNames(&User{})
userFields := esql.RawQueryFields(userFieldNames)
var rows []*User
query := fmt.Sprintf("select %s from user", userFields)
err := db.QueryRows(&rows, query)
if err != nil {
    log.Fatal(err)
}

for _, row := range rows {
    log.Printf("%+v \n", row)
}
```

- 执行
```
user := User{
    Name: "ccc",
}
query := "insert into user(`name`) values(?)"
result, err := db.Exec(query, user.Name)
if err != nil {
    log.Fatal(err)
}

id, _ := result.LastInsertId()
user.ID = int(id)

log.Println(user)
```
- 事务  
自动化事务
```
err := db.Transaction(func(tx *esql.Tx) error {
    userFieldNames := esql.RawFieldNames(User{})
    userFields := esql.RawQueryFields(userFieldNames)
    var user User
    query := fmt.Sprintf("select %s from user where id=?", userFields)
    err := tx.QueryRow(&user, query, 2)
    if err != nil {
        return err
    }

    userFieldsWithPlaceHolder :=  esql.RawUpdateFieldsWithPlaceHolder(userFieldNames, "`id`")
    query = fmt.Sprintf("update user set %s where `id`=?", userFieldsWithPlaceHolder)
    result, err := tx.Exec(query, user.Name+"1", user.Age+1,  2)
    if err != nil {
        return err
    }

    log.Println(result.RowsAffected())
    return nil

})

if err != nil {
    log.Fatal(err)
}
```
手动事务操作
```
// 开启事务
tx, err := db.Begin()
if err != nil {
    log.Fatal(err)
}

....
if err != nil {
    // 回滚事务
    tx.Rollback()
    return err
}


// 提交事务
tx.Commit()
```
- 代码生成  
命令行工具
```bash
dev@virtual-dev:~$esql
Usage of esql:
  -db string
        the database name
  -dsn string
        the dataSource
  -ip string
        the database ip (default "127.0.0.1")
  -mode string
        the database drive (default "mysql")
  -p string
        the database password
  -path string
        the path to save file (default "./")
  -port int
        the database port (default 3306)
  -tag
        the generated structure needs to be tagged
  -u string
        the database user (default "root")


dev@virtual-dev:~$ esql -db test -u root -p 123456 -path ./model
```
方法调用
```
err := db.GenStructByTable(esql.Mysql, "test", "./model", false)
if err != nil {
    log.Println(err)
}
```

- 获取查询/更新字段
```
// 调用以下方法即可

// 获取结构体字段集
esql.RawFieldNames(in interface{}, postgreSql ...bool) []string

// 获取查询字段
esql.RawQueryFields(fieldNames []string) string

// 获取带表名前缀的查询字段
esql.RawQueryFieldsWithPrefix(fieldNames []string, table string) string

// 获取带占位符的更新字段
esql.RawUpdateFieldsWithPlaceHolder(fieldNames []string, str ...string) string
```
- 自定义日志
```
// 实现esql.Logger
type CustomLogger struct {
}

...

// 传入自定义的日志即可
db, err := esql.Open(esql.Mysql, dataSource, &CustomLogger)
if err != nil {
    log.Fatal(err)
}



```


## 参考
`esql`参考了[go-zero](https://github.com/zeromicro/go-zero) `sql`库对象映射的设计，特此感谢。  
