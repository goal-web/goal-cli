package gen

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

type ExtractServiceTemp struct {
	Suffix   string
	Template string
	List     []*Service
}

type Method struct {
	Name                string
	InputImportPackage  string   // biz/request
	OutputImportPackage string   // biz/models
	InputUsageName      string   // 包含包名的完整类型，例如：requests.LoginRequest
	OutputUsageName     string   // 包含包名的完整类型，例如：models.UserModel
	Method              []string // http 方法，控制器才有
	Path                string   // http 路径，控制器才有
	Middlewares         []string
}

type Service struct {
	Name        string
	Methods     []*Method
	PackageName string // 包名，例如：auth
	ImportPath  string // 引用名，例如：pro/biz/services/auth
	UsageName   string // 使用方法，例如：auth.AuthService
	Filename    string
	Template    string

	Middlewares []string
	Controller  bool
	Prefix      string
}

// GenServices 生成 service 代码
func GenServices(baseOutputDir, basePackage string, tmpl *template.Template, services []*Service) []string {
	var files []string
	for _, svc := range services {
		files = append(files, GenService(baseOutputDir, basePackage, tmpl, svc, DetermineServiceImports(svc)))
		fmt.Printf("生成服务文件：%s\n", filepath.Join(baseOutputDir, svc.Filename))

		if svc.Controller {
			svc.Filename = strings.Replace(svc.Filename, "services", "controllers", 1)
			svc.UsageName = strings.Replace(svc.UsageName, filepath.Base(svc.ImportPath), "svc", 1)
			svc.Template = "controller"
			files = append(files, GenService(baseOutputDir, basePackage, tmpl, svc, DetermineServiceImports(svc)))
			fmt.Printf("生成控制器文件：%s\n", filepath.Join(baseOutputDir, svc.Filename))
		}

	}
	return files
}

// SDKServices 生成 service 代码
func SDKServices(baseOutputDir, basePackage string, tmpl *template.Template, services []*Service) []string {
	var files []string
	for _, svc := range services {

		if svc.Controller {
			svc.Filename = strings.ReplaceAll(svc.Filename, ".go", ".ts")
			svc.UsageName = strings.Replace(svc.UsageName, filepath.Base(svc.ImportPath), "svc", 1)
			svc.Template = "controller"
			files = append(files, GenService(baseOutputDir, basePackage, tmpl, svc, DetermineTsServiceImports(svc)))
			fmt.Printf("生成控制器文件：%s\n", filepath.Join(baseOutputDir, svc.Filename))
		}

	}
	return files
}

func GenService(baseOutputDir, basePackage string, tmpl *template.Template, svc *Service, imports []Import) string {
	outputPath := filepath.Join(baseOutputDir, svc.Filename)
	err := os.MkdirAll(filepath.Dir(outputPath), os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}

	outFile, err := os.Create(outputPath)
	if err != nil {
		log.Fatal(err)
	}

	// 执行模板，传入 moduleName、outputPackageName 和 imports
	err = tmpl.ExecuteTemplate(outFile, svc.Template, map[string]interface{}{
		"Package":      svc.PackageName,
		"ImportPath":   svc.ImportPath,
		"UsageName":    svc.UsageName,
		"Middlewares":  svc.Middlewares,
		"Name":         svc.Name,
		"Methods":      svc.Methods,
		"Prefix":       svc.Prefix,
		"Imports":      imports,
		"ResponsePath": fmt.Sprintf("%s/response", basePackage),
	})
	if err != nil {
		log.Fatal(err)
	}
	outFile.Close()
	return outputPath
}
