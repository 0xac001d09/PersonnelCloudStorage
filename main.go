package main

import (
	"filestore-server/handler"
	"fmt"
	"net/http"
)

func main() {

	// 建立路由规则
	http.HandleFunc("/file/upload",handler.UploadHandler)
	http.HandleFunc("/file/upload/suc",handler.UploadSucceedHandler)


	// 端口监听
	err := http.ListenAndServe(":8080", nil)
	if err!= nil {
		fmt.Printf("fail to start server,err:%s",err.Error())
	}

}