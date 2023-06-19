package main

import (
	"fmt"
	"os"
)

func main() {
	exePath, err := os.Executable()
	if err != nil {
		fmt.Println("获取可执行文件路径失败：", err)
		return
	}

	fmt.Println("当前可执行文件路径为：", exePath)
}
