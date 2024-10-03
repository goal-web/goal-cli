package gen

import (
	"bufio"
	"fmt"
	"github.com/emicklei/proto"
	"os"
	"path/filepath"
	"strings"
)

func getComments(comment *proto.Comment, name string, defaultValue string) []string {
	var list []string
	if comment != nil {
		for _, line := range comment.Lines {
			if strings.HasPrefix(line, name) {
				value := strings.TrimPrefix(strings.TrimPrefix(line, name), ":")
				if value == "" {
					value = defaultValue
				}
				list = append(list, value)
			}
		}
	}
	return list
}

// 提取数据的函数

func trim(paths ...string) []string {
	var newPaths []string
	for _, path := range paths {
		if path != "." && path != "" {
			newPaths = append(newPaths, path)
		}
	}
	return newPaths

}

// GetModuleNameAndDir 获取模块名和模块根目录
func GetModuleNameAndDir(startDir string) (string, string, error) {
	dir, err := filepath.Abs(startDir)
	if err != nil {
		return "", "", err
	}
	for {
		goModPath := filepath.Join(dir, "go.mod")
		if _, err := os.Stat(goModPath); err == nil {
			// 找到 go.mod 文件，读取模块名
			file, err := os.Open(goModPath)
			if err != nil {
				return "", "", fmt.Errorf("无法打开 go.mod 文件：%v", err)
			}
			defer file.Close()

			scanner := bufio.NewScanner(file)
			for scanner.Scan() {
				line := strings.TrimSpace(scanner.Text())
				if strings.HasPrefix(line, "module") {
					parts := strings.Fields(line)
					if len(parts) >= 2 {
						return parts[1], dir, nil
					}
				}
			}
			if err := scanner.Err(); err != nil {
				return "", "", fmt.Errorf("读取 go.mod 文件出错：%v", err)
			}
			return "", "", fmt.Errorf("在 go.mod 中未找到模块名")
		}

		// 检查是否到达了根目录
		parentDir := filepath.Dir(dir)
		if parentDir == dir {
			break
		}
		dir = parentDir
	}
	return "", "", fmt.Errorf("在目录及其上级目录中未找到 go.mod 文件")
}
