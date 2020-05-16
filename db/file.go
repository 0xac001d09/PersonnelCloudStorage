package db

import (
	"database/sql"
	_ "database/sql"
	mydb "filestore-server/db/mysql"
	"fmt"
)

// 文件信息写入数据库接口 : 成功返回 true，否则false
func OnFileFinished(filehash string, filename string, filesize int64, fileaddr string) bool {
	conn := mydb.DBConn()

	stmt, err := conn.Prepare(
		"insert ignore " +
			"into tbl_file (`file_sha1`,`file_name`,`file_size`,`file_addr`,`status`) " +
			"values (?,?,?,?,1)")

	if err != nil {
		fmt.Println("fail to prepare statement,err:", err.Error())
		return false
	}
	defer stmt.Close()

	res, err := stmt.Exec(filehash, filename, filesize, fileaddr)
	if err != nil {
		fmt.Println("fail to execute sql,err:", err.Error())
		return false
	}
	if rf, err := res.RowsAffected(); err != nil {
		// 如果sql执行成功，但没有产生一条新的记录，那说明插重了
		if rf <= 0 {
			fmt.Printf("file with hash %s has been uploaded before", filehash)
		}
		return true
	}
	return false
}

// 从数据库获取文件元信息接口
func GetFileMeta(filehash string) (*TableFile, error) {
	stmt, err := mydb.DBConn().Prepare("select file_sha1,file_name,file_size,file_addr from tbl_file where file_sha1=? and status=1 limit 1")
	if nil != err {
		fmt.Println(err.Error())
		return nil, err
	}
	defer stmt.Close()

	tfile := TableFile{}
	// 这个方法总是会返回一个非空的值， 而它引起的错误则会被推延到数据行的 Scan 方法被调用为止。
	// scan就是拿到结果来赋值
	// 这句话相当于 err := db.QueryRow("SELECT username FROM users WHERE id=?", id).Scan(&username)
	err = stmt.QueryRow(filehash).Scan(&tfile.FileHash, &tfile.FileName, &tfile.FileSize, &tfile.FileAddr)
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}
	return &tfile, nil
}

type TableFile struct {
	FileHash string
	FileName sql.NullString
	FileSize sql.NullInt64
	FileAddr sql.NullString
}
