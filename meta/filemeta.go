package meta

import mydb "filestore-server/db"

/*
文件属性(元信息)结构体
*/
type FileMeta struct {
	// 文件的唯一标识 FileSha1，当然可以用MD5等
	FileSha1 string
	// 文件名
	FileName string
	// 文件大小
	FileSize int64
	// 文件存储路径
	Location string
	// 文件上传时间
	UploadAt string
}

//  map : 存储上传信息,key 为FileSha1
var fileMetas map[string]FileMeta

func init() {
	fileMetas = make(map[string]FileMeta)

}

// 新增/更新文件 metadata
func UpdateFileMeta(fmeta FileMeta) {
	fileMetas[fmeta.FileSha1] = fmeta
}

// 更新文件元信息到 mysql 中
func UpdateFileMetaDB(fMeta FileMeta) bool {
	return mydb.OnFileFinished(fMeta.FileSha1, fMeta.FileName, fMeta.FileSize, fMeta.Location)
}

// 获取文件的 metadata
func GetFileMeta(filesha1 string) FileMeta {
	return fileMetas[filesha1]
}

// 从数据库获取文件元信息
func GetFileMetaDB(fileSha1 string) (FileMeta, error) {
	tableFile, err := mydb.GetFileMeta(fileSha1)
	if err != nil {
		return FileMeta{}, err
	}
	// tablefile -> filemeta,要把 sql 中的取出来
	fMeta := FileMeta{
		FileSha1: tableFile.FileHash,
		FileSize: tableFile.FileSize.Int64,
		FileName: tableFile.FileName.String,
		Location: tableFile.FileAddr.String,
	}
	return fMeta, nil
}

// 批量获取文件源信息列表
//func GetFileMetas(count int) []FileMeta {
//}

func RemoveFileMeta(fileSha1 string) {
	// TODO 需要考虑安全判断
	delete(fileMetas, fileSha1)
}
