package main

import (
	file "file/File"
	"file/belog"
	"fmt"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	// === 初始化日志 ===
	logPath := "log.txt"
	_, err := belog.InitZapLogger(
		logPath, 500, zapcore.InfoLevel)
	if err != nil {
		fmt.Printf("InitZapLogger error: %s\n", err.Error())
		os.Exit(1)
	}
	defer zap.L().Sync()
	defer zap.S().Sync()

	// === 下载、解密、解压缩 ===
	zap.S().Infow("file download is begining")
	Urlstr := make(chan string, 200)
	file.ReadExcel(Urlstr)
	zap.S().Infow("file is ending")

	// // === 加密文件 ===
	// zap.S().Infow("file encryption is begining")
	// originalpath := "C:/Users/JD/Desktop/1.txt"
	// key := "ajfxio6r4qj1cz23"
	// afterPath := "C:/Users/JD/Desktop/1"
	// byt, err := os.ReadFile(originalpath)
	// if err != nil {
	// 	fmt.Printf("Read file error: %s\n", err.Error())
	// }

	// fl, err := utils.AES_CBC_Encrypt(byt, []byte(key))
	// if err != nil {
	// 	fmt.Printf("Encryption file error: %s\n", err.Error())
	// }
	// file, err := os.Create(afterPath)
	// if err != nil {
	// 	fmt.Printf("Create file error: %s\n", err.Error())
	// }
	// _, err = file.Write(fl)
	// if err != nil {
	// 	fmt.Printf("Write file error: %s\n", err.Error())
	// }
	// zap.S().Infow("file is ending")
}
