package gen

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"strings"
)

// AddHeaderAndFormatFiles 给文件数组中的每个文件添加头部注释、移除未使用的引用，并格式化代码
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

// addHeaderAndFormat 格式化指定的 Go 文件，并在文件头部添加指定注释，移除未使用的 import
func addHeaderAndFormat(filename, headerComment string) error {
	// 解析源文件
	fileSet := token.NewFileSet()
	node, err := parser.ParseFile(fileSet, filename, nil, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("failed to parse file: %v", err)
	}

	// 检查文件是否已经包含指定的注释
	if node.Doc == nil || !strings.Contains(node.Doc.Text(), headerComment) {
		// 添加头部注释
		fmt.Printf("Adding header comment to file: %s\n", filename)
		header := &ast.CommentGroup{
			List: []*ast.Comment{
				{Slash: node.Package, Text: headerComment},
			},
		}
		node.Comments = append([]*ast.CommentGroup{header}, node.Comments...)
		node.Doc = header
	}

	// 移除未使用的 import 语句
	removeUnusedImports(node)

	// 格式化 AST 并输出到缓冲区
	var buf bytes.Buffer
	if err := format.Node(&buf, fileSet, node); err != nil {
		return fmt.Errorf("failed to format code: %v", err)
	}

	// 将格式化后的代码写回文件
	if err := os.WriteFile(filename, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write formatted code to file: %v", err)
	}
	fmt.Printf("File formatted and updated successfully: %s\n", filename)

	return nil
}

// removeUnusedImports 移除未使用的 import 语句
func removeUnusedImports(f *ast.File) {
	// 创建一个 map 来存储所有导入的包名
	imports := make(map[string]*ast.ImportSpec)
	for _, imp := range f.Imports {
		pathValue := strings.Trim(imp.Path.Value, `"`)
		imports[pathValue] = imp
	}

	// 使用 ast.Inspect 遍历所有节点，找到所有被引用的包
	usedImports := make(map[string]struct{})
	ast.Inspect(f, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.SelectorExpr: // 查找所有包.名称 形式的调用
			if ident, ok := node.X.(*ast.Ident); ok {
				usedImports[ident.Name] = struct{}{}
			}
		case *ast.Ident: // 查找所有直接使用的标识符
			usedImports[node.Name] = struct{}{}
		}
		return true
	})

	// 移除未使用的 import
	var newImports []*ast.ImportSpec
	for path, spec := range imports {
		// 检查这个包名是否被使用了
		// 如果 import 使用了自定义别名 `import x "package/path"`，需要用 `spec.Name.Name` 来判断
		name := path
		if spec.Name != nil {
			name = spec.Name.Name
		}

		if _, used := usedImports[name]; used {
			newImports = append(newImports, spec) // 保留已使用的 import
		} else {
			fmt.Printf("Removing unused import: %s\n", path)
		}
	}
	f.Imports = newImports
}
