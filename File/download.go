package file

import (
	"file/utils"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/rs/xid" // v2版本的，V3版本太复杂，没研究明白
	"github.com/tealeg/xlsx"
	"go.uber.org/zap"
)

//从xlsx中获取文件链接，拼接文件头
func ReadExcel(Urlstr chan string) chan string {

	// 打开 xlsx文件
	xlFile, err := xlsx.OpenFile("F:/WebDownload/200样本url.xlsx")
	if err != nil {
		fmt.Println("打开文件失败", err.Error())
		return nil
	}

	// 遍历 sheet 页
	for _, sheet := range xlFile.Sheets {
		// 行
		for _, row := range sheet.Rows {

			// 列
			// var temp Person

			// 将excel每一列文件读取放在字符串切片中
			var str string = row.Cells[0].String()
			// for _, cell := range row.Cells {
			// 	str = append(str, cell.String())
			// }

			// temp.fileUrl = str[0]
			if str == "fileUrl" {
				continue
			}
			head := "https://ocs-cn-south1.heytapcs.com"
			str = head + str
			Urlstr <- str
			operator(Urlstr)
		}

	}
	return Urlstr
}

//获取zip和txt的名称，以及文件名；传输给DownloadFileData进行下载
func operator(Urlstr chan string) {
	var uuid, zipPath, txtPath, fileurl string

	fileurl = <-Urlstr

	uuid = getUuid()
	zipPath = uuid + ".zip"
	txtPath = uuid + ".txt"

	zap.S().Infow("Downloading the file is begining", "fileurl", fileurl)

	err := DownloadFileData(fileurl, "ajfxio6r4qj1cz23", zipPath, txtPath)
	if err != nil {
		fmt.Printf("DownloadFileData error: %s\n", err.Error())
		os.Exit(1)
	}
}

//获取唯一的uuid
func getUuid() string {
	t := time.Now()
	guid := xid.NewWithTime(t)
	return guid.String()
}

//下载文件，解密解压缩，保留压缩文件和文本文档
func DownloadFileData(downPath string, fileSecret string, zipPath string, localPath string) error {
	buf, err := Download(downPath)
	if err != nil {
		return err
	}
	//解密
	// zap.S().Infow("Start decrypting file", "zipPath", zipPath)
	zipData, err := utils.AES_CBC_Decrypt(buf, []byte(fileSecret))
	if err != nil {
		return err
	}
	file, err := os.Create(zipPath)
	if err != nil {
		return err
	}
	_, err = file.Write(zipData)
	if err != nil {
		file.Close()
		return err
	}
	zap.S().Infow("Decrypt file is complete", "zipPath", zipPath)
	file.Close()
	//解压缩
	// zap.S().Infow("Start unzipping the file", "localPath", localPath)
	err = utils.Decompression(zipPath, localPath, buf)
	if err != nil {
		return err
	}

	zap.S().Infow("Unzipping file is complete", "localPath", localPath)

	errDelete := os.Remove(zipPath)
	if errDelete != nil {
		zap.S().Errorw("Failed to Delete Zip file", "zipPath", zipPath, "error", errDelete)
	}
	return nil
}

//下载文件
func Download(url string) ([]byte, error) {

	// Download the tar file.
	r, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	// Read in the entire contents of the file.
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	// Return the file.
	return body, nil
}
