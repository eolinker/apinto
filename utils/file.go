package utils

import "os"

// GenFile 根据路径和数据生成文件
func GenFile(dir, fileName, data string) error {
	dir = "work/export/" + dir
	err := os.MkdirAll(dir, 0666)
	if err != nil {
		return err
	}
	file, err := os.OpenFile(dir+fileName, os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.WriteString(data)
	if err != nil {
		return err
	}
	return nil
}
