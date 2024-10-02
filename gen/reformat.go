package gen

import (
	"bytes"
	"fmt"
	"go/format"
	"io/ioutil"
	"os"
	"strings"
)

// AddHeaderAndFormatFiles 给文件数组中的每个文件添加头部注释并格式化代码
// 参数:
// - files: 文件路径数组，表示需要处理的文件列表
// - headerComment: 需要添加的文件头部注释内容
func AddHeaderAndFormatFiles(files []string, headerComment string) error {
	for _, file := range files {
		// 检查文件是否存在
		if _, err := os.Stat(file); os.IsNotExist(err) {
			fmt.Printf("File does not exist: %s\n", file)
			continue
		}

		// 检查是否为 Go 文件
		if !strings.HasSuffix(file, ".go") {
			fmt.Printf("Skipping non-Go file: %s\n", file)
			continue
		}

		// 格式化并添加注释
		err := addHeaderAndFormat(file, headerComment)
		if err != nil {
			fmt.Printf("Failed to process file %s: %v\n", file, err)
			return err
		}
	}

	return nil
}

// addHeaderAndFormat 格式化指定的 Go 文件，并在文件头部添加指定注释
func addHeaderAndFormat(filename, headerComment string) error {
	// 读取文件内容
	src, err := ioutil.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read file: %v", err)
	}

	// 检查文件是否已经包含指定的注释
	if !bytes.HasPrefix(src, []byte(headerComment)) {
		// 在文件内容前添加注释
		fmt.Printf("Adding header comment to file: %s\n", filename)
		src = append([]byte(headerComment), src...)
	}

	// 格式化 Go 代码
	formattedSrc, err := format.Source(src)
	if err != nil {
		return fmt.Errorf("failed to format code in file %s: %v", filename, err)
	}

	// 检查是否有必要写回文件
	if !bytes.Equal(src, formattedSrc) {
		fmt.Printf("Formatting file: %s\n", filename)

		// 将格式化后的内容写回文件
		err = ioutil.WriteFile(filename, formattedSrc, 0644)
		if err != nil {
			return fmt.Errorf("failed to write formatted code to file: %v", err)
		}
		fmt.Printf("File formatted and updated successfully: %s\n", filename)
	} else {
		fmt.Printf("File is already formatted and includes header: %s\n", filename)
	}

	return nil
}
