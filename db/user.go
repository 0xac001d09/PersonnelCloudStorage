package db

import (
	mydb "filestore-server/db/mysql"
	"fmt"
)

// 用户注册，写库
func UserSignup(username string, passwd string) bool {
	stmt, err := mydb.DBConn().Prepare("insert ignore into tbl_user (`user_name`,`user_pwd`) values (?,?)")
	if err != nil {
		fmt.Println("fail to insert,err:", err.Error())
		return false
	}
	defer stmt.Close()

	res, err := stmt.Exec(username, passwd)
	if err != nil {
		fmt.Println("fail to insert,err:", err.Error())
		return false
	}
	// 重复注册的情况
	if ra, err := res.RowsAffected(); err == nil && ra > 0 {
		return true
	}
	return false
}

// 用户登陆查库,判断密码是否一致
func UserSignin(username string, encpwd string) bool {
	stmt, err := mydb.DBConn().Prepare("select * from tbl_user where user_name=? limit 1")
	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	rows, err := stmt.Query(username)
	if err != nil {
		fmt.Println(err.Error())
		return false
	} else if rows == nil {
		fmt.Println("user not found:" + username)
		return false
	}

	// 对比 pwd,parseRows 来将查询到的 rows 转成元素为 map 类型的数组
	pRows := mydb.ParseRows(rows)
	if len(pRows) > 0 && string(pRows[0]["user_pwd"].([]byte)) == encpwd {
		return true
	}
	fmt.Println(pRows[0]["user_pwd"])
	return false
}

// 更新用户登陆的 token
func UpdateToken(username string, token string) bool {
	stmt, err := mydb.DBConn().Prepare("replace into tbl_user_token(`user_name`,`user_token`) values (?,?)")
	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	defer stmt.Close()

	_, err = stmt.Exec(username, token)
	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	return true
}

// 返回对应用户信息的一条记录
func GetUserInfo(username string) (User, error) {
	user := User{}

	stmt, err := mydb.DBConn().Prepare("select user_name,signup_at from tbl_user where user_name=? limit 1")
	if err != nil {
		fmt.Println(err.Error())
		return user, err
	}
	defer stmt.Close()

	err = stmt.QueryRow(username).Scan(&user.Username, &user.SignupAt)
	if err != nil {
		return user, err
	}
	return user, nil
}

type User struct {
	Username   string
	Email      string
	Phone      string
	SignupAt   string
	LastActive string
	Status     int
}
