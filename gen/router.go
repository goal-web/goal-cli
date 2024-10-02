package gen

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"strings"
)

type RouterCollector struct {
	ImportPath string
	UsageName  string
}

func GenRouters(output string, services []*Service) {
	var routers []*RouterCollector
	for _, service := range services {
		if service.Controller {
			routers = append(routers, &RouterCollector{
				ImportPath: strings.Replace(service.ImportPath, "services", "controllers", 1),
				UsageName:  fmt.Sprintf("%sRouter(router)", service.UsageName),
			})
		}
	}

	for _, router := range routers {
		GenRouter(strings.Join([]string{output, "controllers", "kernel.go"}, "/"), router.ImportPath, router.UsageName)
	}
}

// GenRouter 通过指定的 importPath 和 usage 动态修改指定 Go 文件中的路由注册函数
func GenRouter(filename, importPath, usage string) {
	// 检查并创建文件，如果文件不存在则创建
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		fmt.Printf("File %s does not exist. Creating a new file...\n", filename)
		createInitialFile(filename)
	}

	// 解析文件
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		fmt.Println("Error parsing file:", err)
		return
	}

	// 获取现有的 import 别名和包路径
	imports := getImportAliases(node)

	// 检查 importPath 是否已存在，存在则直接使用已有的别名
	alias, exists := imports[importPath]
	if !exists {
		// 如果 importPath 不存在，则生成合适的 alias
		alias = generateUniqueAlias(importPath, imports)
		fmt.Printf("Adding import: %s as %s\n", importPath, alias)
		addImport(node, importPath, alias)
		imports[importPath] = alias // 记录新增的 import
	} else {
		fmt.Printf("Import already exists: %s as %s\n", importPath, alias)
	}

	// 确定实际使用的包名（可能是 alias）
	actualUsage := strings.Replace(usage, strings.Split(usage, ".")[0], alias, 1)

	// 查找 Register 函数，并插入新的路由调用
	modified := false
	ast.Inspect(node, func(n ast.Node) bool {
		if fn, ok := n.(*ast.FuncDecl); ok && fn.Name.Name == "Register" {
			// 查找是否已经存在相同的调用，避免重复插入
			for _, stmt := range fn.Body.List {
				if exprStmt, ok := stmt.(*ast.ExprStmt); ok {
					if call, ok := exprStmt.X.(*ast.CallExpr); ok {
						// 格式化现有的调用，判断是否已存在
						if formatNode(call, fset) == actualUsage {
							fmt.Printf("Router call already exists: %s\n", actualUsage)
							return false
						}
					}
				}
			}

			// 插入新的调用语句
			fmt.Printf("Inserting router call: %s\n", actualUsage)
			newCallStmt := createRouterCallStmt(actualUsage)
			fn.Body.List = append(fn.Body.List, newCallStmt)
			modified = true
		}
		return true
	})

	// 如果文件被修改，则写回文件
	if modified {
		fmt.Println("File modified, saving changes...")
		f, err := os.Create(filename)
		if err != nil {
			fmt.Println("Error creating file:", err)
			return
		}
		defer f.Close()

		if err := printer.Fprint(f, fset, node); err != nil {
			fmt.Println("Error printing file:", err)
		} else {
			fmt.Println("File successfully updated.")
		}
	} else {
		fmt.Println("No modifications made to the file.")
	}
}

// 获取文件中现有的 import 语句及其别名
func getImportAliases(node *ast.File) map[string]string {
	imports := make(map[string]string)
	for _, imp := range node.Imports {
		importPath := strings.Trim(imp.Path.Value, `"`)
		if imp.Name != nil {
			imports[importPath] = imp.Name.Name
		} else {
			imports[importPath] = getPackageName(importPath)
		}
	}
	return imports
}

// 生成唯一的 import 别名
func generateUniqueAlias(importPath string, imports map[string]string) string {
	baseAlias := getPackageName(importPath)
	alias := baseAlias
	i := 1

	// 如果别名已存在，生成唯一别名（例如 user1, user2 等）
	for containsValue(imports, alias) {
		alias = fmt.Sprintf("%s%d", baseAlias, i)
		i++
	}
	return alias
}

// 检查 map 中是否已包含某个值
func containsValue(m map[string]string, value string) bool {
	for _, v := range m {
		if v == value {
			return true
		}
	}
	return false
}

// 获取路径的包名（假设包名是路径的最后一个部分）
func getPackageName(importPath string) string {
	parts := strings.Split(importPath, "/")
	return parts[len(parts)-1]
}

// 添加 import 语句（包含别名）
func addImport(node *ast.File, importPath, alias string) {
	var newImport *ast.ImportSpec
	if alias != getPackageName(importPath) {
		// 需要别名时，设置 Name 字段
		newImport = &ast.ImportSpec{
			Name: &ast.Ident{Name: alias},
			Path: &ast.BasicLit{Kind: token.STRING, Value: fmt.Sprintf(`"%s"`, importPath)},
		}
	} else {
		newImport = &ast.ImportSpec{
			Path: &ast.BasicLit{Kind: token.STRING, Value: fmt.Sprintf(`"%s"`, importPath)},
		}
	}

	// 查找现有的 import 语句块
	if len(node.Imports) == 0 {
		// 如果没有 import 语句块，创建一个新的 import 语句
		node.Decls = append([]ast.Decl{&ast.GenDecl{
			Tok:   token.IMPORT,
			Specs: []ast.Spec{newImport},
		}}, node.Decls...)
	} else {
		// 如果已存在 import 语句块，则插入到现有的 import 语句中
		for _, decl := range node.Decls {
			if genDecl, ok := decl.(*ast.GenDecl); ok && genDecl.Tok == token.IMPORT {
				genDecl.Specs = append(genDecl.Specs, newImport)
				break
			}
		}
	}
}

// 格式化 AST 节点为字符串表示形式
func formatNode(node ast.Node, fset *token.FileSet) string {
	var sb strings.Builder
	if err := printer.Fprint(&sb, fset, node); err != nil {
		return ""
	}
	return sb.String()
}

// 创建新的 router 调用语句
func createRouterCallStmt(usage string) ast.Stmt {
	expr, err := parser.ParseExpr(usage)
	if err != nil {
		fmt.Println("Error creating new call expression:", err)
		return nil
	}
	return &ast.ExprStmt{X: expr}
}

// 创建初始 Go 文件的内容
func createInitialFile(filename string) {
	initialContent := `package controllers

import (
	"github.com/goal-web/contracts"
)

// Register 注册路由函数
func Register(router contracts.HttpRouter) {
	// 在这里添加您的路由注册逻辑
}
`
	err := os.WriteFile(filename, []byte(initialContent), 0644)
	if err != nil {
		fmt.Printf("Error creating initial file: %v\n", err)
	} else {
		fmt.Printf("Initial file %s created successfully.\n", filename)
	}
}
