package main

import (
	"filestore-server/handler"
	"fmt"
	"net/http"
)

func main() {

	// 建立路由规则
	http.HandleFunc("/file/upload", handler.UploadHandler)
	http.HandleFunc("/file/upload/suc", handler.UploadSucceedHandler)
	http.HandleFunc("/file/meta", handler.GetFileMetaHandler)
	http.HandleFunc("/file/download", handler.DownloadHandler)
	http.HandleFunc("/file/update", handler.FileMetaUpdateHandler)
	http.HandleFunc("/file/delete", handler.FileDeleteHandler)

	// 端口监听
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Printf("fail to start server,err:%s", err.Error())
	}

}
