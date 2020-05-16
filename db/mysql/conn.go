package mysql

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"os"
	"strings"
)

var db *sql.DB

// 参考：https://www.jianshu.com/p/ee87e989f149
const (
	userName = "root"
	password = "root"
	ip       = "127.0.0.1"
	port     = "32768"
	dbName   = "fileserver"
)

func init() {
	path := strings.Join([]string{userName, ":", password, "@tcp(", ip, ":", port, ")/", dbName, "?charset=utf8"}, "")
	// 设置驱动和数据库连接
	db, _ = sql.Open("mysql", path)
	fmt.Println(&db)
	// 设置最大连接数
	db.SetMaxOpenConns(100)
	// 验证连接
	err := db.Ping()
	if err != nil {
		fmt.Printf("fail to connect to mysql,err:" + err.Error())
		// 出错强制退出
		os.Exit(1)
	}
	fmt.Println("mysql connect success")
}

// 外部调用接口：返回数据库连接对象
func DBConn() *sql.DB {
	return db
}

// 将查询到的 rows 转成元素为 map 类型的数组
func ParseRows(rows *sql.Rows) []map[string]interface{} {
	columns, _ := rows.Columns()
	scanArgs := make([]interface{}, len(columns))
	values := make([]interface{}, len(columns))
	for j := range values {
		scanArgs[j] = &values[j]
	}

	record := make(map[string]interface{})
	records := make([]map[string]interface{}, 0)
	for rows.Next() {
		//将行数据保存到record字典
		err := rows.Scan(scanArgs...)
		checkErr(err)

		for i, col := range values {
			if col != nil {
				record[columns[i]] = col
			}
		}
		records = append(records, record)
	}
	return records
}

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
		panic(err)
	}
}
