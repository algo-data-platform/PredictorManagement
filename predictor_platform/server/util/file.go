package util

import (
	"io"
	"os"
)

// 文件中写入数据
func WriteFile(file_name string, data []byte) (int, error) {
	fl, err := os.OpenFile(file_name, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return 0, err
	}
	defer fl.Close()
	n, err := fl.Write(data)
	if err == nil && n < len(data) {
		err = io.ErrShortWrite
	}
	return n, err
}
