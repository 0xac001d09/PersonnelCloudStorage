// 处理文件上传接口
package handler

import (
	"encoding/json"
	dao "filestore-server/db"
	"filestore-server/meta"
	"filestore-server/util"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"time"
)

/**
定义输入参数
1、用于向用户返回数据的resp对象
2、用于接收用户请求的request对象的指针
*/

// 上传文件接口
func UploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		// 1、如果用户是请求，加载文件返回上传的http页面
		file, err := ioutil.ReadFile("./static/view/index.html")
		if err != nil {
			io.WriteString(w, "internal server error")
			return
		}
		// 成功就把文件内容返回出去
		io.WriteString(w, string(file))

	} else if r.Method == "POST" {
		// 1、如果是POST请求，要接收用户上传的文件流，存储到本地目录
		// 文件句柄、文件头、错误信息
		file, head, err := r.FormFile("file")
		if err != nil {
			fmt.Printf("fail to get data,err:%s", err.Error())
			return
		}
		defer file.Close()

		fileMeta := meta.FileMeta{
			FileName: head.Filename,
			Location: "/Users/zhangye/gotemp/" + head.Filename,
			UploadAt: time.Now().Format("2006-01-02 15:04:05"),
		}

		// 2、创建的新的文件路径
		newFile, err := os.Create(fileMeta.Location)
		if err != nil {
			fmt.Printf("fail to create file,err:%s", err.Error())
			return
		}
		// 不要忘记在函数退出之前关闭文件资源
		defer newFile.Close()

		// 3、将内存中的文件拷贝到新的文件的 buffer 中去，返回写入大小与 err
		fileMeta.FileSize, err = io.Copy(newFile, file)
		if err != nil {
			fmt.Printf("fail to copy file to buffer,err:%s", err.Error())
			return
		}

		// 从头开始，文件指针偏移 0
		newFile.Seek(0, 0)
		fileMeta.FileSha1 = util.FileSha1(newFile)
		//meta.UpdateFileMeta(fileMeta)
		meta.UpdateFileMetaDB(fileMeta)

		// 更新 user-file 表记录
		r.ParseForm()
		username := r.Form.Get("username")
		suc := dao.OnUserFileUploadFinished(username, fileMeta.FileSha1, fileMeta.FileName, fileMeta.FileSize)
		if suc {
			http.Redirect(w, r, "/static/view/home.html", http.StatusFound)
		} else {
			w.Write([]byte("Upload Failed"))
		}

		// 4、文件写入成功，给定成功信息
		http.Redirect(w, r, "/file/upload/suc", http.StatusFound)
	}
}

// 上传已完成 handler
func UploadSucceedHandler(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "Upload finished!")
}

// 获取文件元信息接口
func GetFileMetaHandler(w http.ResponseWriter, r *http.Request) {
	// 解析，入参第一个为 fileHash
	r.ParseForm()
	filehash := r.Form["filehash"][0]
	//fileMeta := meta.GetFileMeta(filehash)
	fileMeta, err := meta.GetFileMetaDB(filehash)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// 转成json返回客户端
	data, err := json.Marshal(fileMeta)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(data)
}

// 批量查询文件元信息
func FileQueryHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	limitCnt, _ := strconv.Atoi(r.Form.Get("limit"))
	username := r.Form.Get("username")
	//fileMetas, _ := meta.GetLastFileMetasDB(limitCnt)
	userFiles, err := dao.QueryUserFileMetas(username, limitCnt)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	data, err := json.Marshal(userFiles)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(data)
}

// 文件下载接口
func DownloadHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	fsha1 := r.Form.Get("filehash")
	fileMeta := meta.GetFileMeta(fsha1)

	// 拿到下载位置
	file, err := os.Open(fileMeta.Location)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer file.Close()

	// 测试小文件直接读了
	data, err := ioutil.ReadAll(file)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	// 告诉浏览器下载成功
	w.Header().Set("Content-Type", "application/octetc-stream")
	w.Header().Set("content-disposition", "attachment;fimename=\""+fileMeta.FileName+"\"")
	w.Write(data)
}

// 更新文件元信息接口（文件重命名
func FileMetaUpdateHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	// 解析请求的参数列表，客户端带三个参数，操作类型、文件唯一标识（filesha1）、文件新名字
	opType := r.Form.Get("op")
	fileSha1 := r.Form.Get("filehash")
	newFileName := r.Form.Get("fimename")

	// 如果不是重命名操作，报错
	if opType != "0" {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	// 只支持 post 请求
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// 获取当前文件元信息
	curFileMeta := meta.GetFileMeta(fileSha1)
	curFileMeta.FileName = newFileName
	meta.UpdateFileMeta(curFileMeta)

	data, err := json.Marshal(curFileMeta)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(data)
	w.WriteHeader(http.StatusOK)
}

// 删除文件接口
func FileDeleteHandler(w http.ResponseWriter, r *http.Request) {
	// 只接收一个参数 filesha1
	r.ParseForm()
	fileSha1 := r.Form.Get("filehash")
	fileMeta := meta.GetFileMeta(fileSha1)

	// 物理上删除
	os.Remove(fileMeta.Location)

	// 删除索引
	meta.RemoveFileMeta(fileSha1)

	w.WriteHeader(http.StatusOK)
}

// 尝试秒传接口
func TryFastUploadHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	// 1. 解析请求参数
	username := r.Form.Get("username")
	filehash := r.Form.Get("filehash")
	filename := r.Form.Get("filename")
	filesize, _ := strconv.Atoi(r.Form.Get("filesize"))

	// 2. 从文件表中查询相同hash的文件记录
	fileMeta, err := meta.GetFileMetaDB(filehash)
	if err != nil {
		fmt.Println(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// 3. 查不到记录则返回秒传失败 (fileMeta == nil 不行)
	if fileMeta == (meta.FileMeta{}) {
		resp := util.RespMsg{
			Code: -1,
			Msg:  "秒传失败，请访问普通上传接口",
		}
		w.Write(resp.JSONBytes())
		return
	}

	// 4. 之前上传过，则将文件信息写入用户文件表，返回成功
	suc := dao.OnUserFileUploadFinished(username, filehash, filename, int64(filesize))
	if suc {
		resp := util.RespMsg{
			Code: 0,
			Msg:  "秒传成功",
		}
		w.Write(resp.JSONBytes())
		return
	}
	resp := util.RespMsg{
		Code: -2,
		Msg:  "秒传失败，请稍后重试",
	}
	w.Write(resp.JSONBytes())
	return
}
