package utils

import (
	"archive/zip"
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"go.uber.org/zap"
)

type MemoryPool struct {
	Chs chan []byte
}

func InitMemory(size int) *MemoryPool {
	mp := &MemoryPool{Chs: make(chan []byte, size)}
	for i := 0; i < cap(mp.Chs); i++ {
		mp.Chs <- make([]byte, 5*1024*1024)
	}
	return mp
}

func DownloadFileData(downPath string, fileSecret string, zipPath string, localPath string, memoryItem []byte) error {
	tic := time.Now()

	// zap.S().Infow("Start downloading files", "downPath", downPath)
	request, err := http.NewRequest("GET", downPath, nil)
	if err != nil {
		return err
	}
	//超时机制
	client := http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(request)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		zap.S().Errorw("The status code returns an error", "StatusCode", resp.StatusCode, "RespBody", resp.Body)
		return fmt.Errorf("the status code: %d returns an error", resp.StatusCode)
	}

	beginIndex := 0
	memoryLen := len(memoryItem)
	for {
		length, err := resp.Body.Read(memoryItem[beginIndex:memoryLen])
		beginIndex += length
		if err != nil {
			break
		}
		if len(memoryItem) < beginIndex {
			return fmt.Errorf("insufficient memorypool size, memoryPool: %d < respbody: %d", len(memoryItem), beginIndex)
		}
	}
	zap.S().Infow("Downloading the file is complete", "file length:", beginIndex, "downPath:", downPath) //下载完成 文件大小，文件名

	if (beginIndex % 16) != 0 {
		return fmt.Errorf("the bodylength: %d is not a multiple of 16", beginIndex)
	}

	//解密
	// zap.S().Infow("Start decrypting file", "zipPath", zipPath)
	zipData, err := AES_CBC_Decrypt(memoryItem[:beginIndex], []byte(fileSecret))
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
	err = Decompression(zipPath, localPath, memoryItem)
	if err != nil {
		return err
	}

	zap.S().Infow("Unzipping file is complete", "localPath", localPath)

	errDelete := os.Remove(zipPath)
	if errDelete != nil {
		zap.S().Errorw("Failed to Delete Zip file", "zipPath", zipPath, "error", errDelete)
	}

	toc := time.Now()
	zap.S().Infow("fileUtil", "TIME", toc.Sub(tic).Milliseconds())
	return nil
}

// Padding 对明文进行填充
func Padding(plainText []byte, blockSize int) []byte {
	//计算要填充的长度
	n := blockSize - len(plainText)%blockSize
	//对原来的明文填充n个n
	temp := bytes.Repeat([]byte{byte(n)}, n)
	plainText = append(plainText, temp...)
	return plainText
}

// UnPadding 对密文删除填充
func UnPadding(cipherText []byte) []byte {
	if len(cipherText) != 0 {
		//取出密文最后一个字节end
		end := cipherText[len(cipherText)-1]
		//删除填充
		cipherText = cipherText[:len(cipherText)-int(end)]
		return cipherText
	} else {
		return nil
	}
}

// AES_CBC_Encrypt 加密
func AES_CBC_Encrypt(plainText []byte, key []byte) ([]byte, error) {
	//指定加密算法，返回一个AES算法的Block接口对象
	block, err := aes.NewCipher([]byte(base64.StdEncoding.EncodeToString(key)))
	if err != nil {

		return []byte{}, err
	}
	//进行填充
	plainText = Padding(plainText, block.BlockSize())
	//指定初始向量vi,长度和block的块尺寸一致
	//iv := []byte("ASDFGHJKLASDFGHJ")
	//指定分组模式，返回一个BlockMode接口对象
	blockMode := cipher.NewCBCEncrypter(block, key)
	//加密连续数据库
	cipherText := make([]byte, len(plainText))
	blockMode.CryptBlocks(cipherText, plainText)
	//返回密文
	return cipherText, nil
}

// AES_CBC_Decrypt 解密
func AES_CBC_Decrypt(cipherText []byte, key []byte) ([]byte, error) {
	//指定解密算法，返回一个AES算法的Block接口对象
	block, err := aes.NewCipher([]byte(base64.StdEncoding.EncodeToString(key)))
	if err != nil {
		return []byte{}, err
	}
	//指定初始化向量IV,和加密的一致
	//指定分组模式，返回一个BlockMode接口对象
	iv := []byte("ajfxio6r4qj1cz23")
	blockMode := cipher.NewCBCDecrypter(block, iv)
	//解密
	// plainText := make([]byte, len(cipherText))
	// plainText := ch[:len(cipherText)]
	blockMode.CryptBlocks(cipherText, cipherText)
	//删除填充
	// cipherText = UnPadding(cipherText)
	return cipherText, nil
}

// Decompression 解压缩
func Decompression(zipPath string, localPath string, memoryItem []byte) error {
	// const BUF_SIZE = 1024 * 1024
	// var buf []byte = make([]byte, BUF_SIZE)

	zipReader, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer zipReader.Close()

	txt_f, err := os.Create(localPath)
	if err != nil {
		zap.S().Errorw("create undecompress file error", "FileName", zipPath, "error", err)
		return err
	}

	defer txt_f.Close()

	for _, f := range zipReader.File {
		r, err := f.Open()
		if err != nil {
			zap.S().Errorw("open zip file error", "FileName", zipPath, "error", err)
			return err
		}

		for {
			ret, err := r.Read(memoryItem)
			if (err == io.EOF) && (ret == 0) {
				zap.S().Infow("File reading end", "file", zipPath)
				break
			}
			_, wErr := txt_f.Write(memoryItem[0:ret])
			if wErr != nil {
				zap.S().Errorw("write undecompress file error", "FileName", zipPath, "error", wErr)
				break
			}
		}
		r.Close()
		break
	}

	zap.S().Infow("The file is decompressed", "FileName", zipPath)
	return nil
}

// func Decompression(zipPath string, localPath string) error {
// 	var data []byte

// 	zipReader, err := zip.OpenReader(zipPath)
// 	if err != nil {
// 		return err
// 	}
// 	defer zipReader.Close()
// 	for _, f := range zipReader.File {
// 		r, err := f.Open()
// 		if err != nil {
// 			r.Close()
// 			return err
// 		}
// 		buf, err := ioutil.ReadAll(r) //读固定大小数据
// 		if err != nil {
// 			r.Close()
// 			return err
// 		}
// 		data = append(data, buf...)
// 		r.Close()
// 	}

// 	if data == nil {
// 		return fmt.Errorf("The file unzipped is empty, dataLength:", len(data))
// 	}

// 	err = ioutil.WriteFile(localPath, data, os.ModePerm)
// 	if err != nil {
// 		return err
// 	}
// 	zap.S().Infow("The file is decompressed", "FileName", zipPath, "Size", len(data))
// 	return nil
// }
