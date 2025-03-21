package gen

import (
	"fmt"
	"path/filepath"
	"strings"
)

type Import struct {
	Pkg   string
	Alias string
}

func DetermineMessageImports(message *Message) []Import {
	importsSet := make(map[string]string)
	usageMap := map[string]string{}
	fields := append(message.Fields, message.Relations...)
	var base string

	for i, field := range fields {
		if message.IsModel && field.IsModel {
			continue
		}
		if alias, exists := usageMap[field.ImportPath]; exists {
			field.UsageName = strings.ReplaceAll(field.UsageName, filepath.Base(field.ImportPath), alias)
		} else {
			if msg, exists := usagePackageMap[field.Type]; exists && msg.ImportPath != message.ImportPath {
				if HasComment(msg.Comment, "@goType") {
					continue
				}

				base = filepath.Base(msg.ImportPath)
				if path, exists := importsSet[base]; exists && path != msg.ImportPath {
					base = fmt.Sprintf("%s%d", base, i)
					usageMap[base] = msg.ImportPath
				}
				importsSet[base] = msg.ImportPath
				field.UsageName = fmt.Sprintf("%s.%s", base, field.Type)
			}
		}

	}
	var imports []Import
	for alias, pkg := range importsSet {
		imp := Import{
			Pkg:   pkg,
			Alias: alias,
		}
		imports = append(imports, imp)
	}
	return imports
}

func DetermineTsMessageImports(message *Message) []Import {
	importsSet := make(map[string]string)
	usageMap := map[string]string{}
	fields := append(message.Fields, message.Relations...)
	var base string

	for i, field := range fields {
		if message.IsModel && field.IsModel {
			continue
		}

		if msg, ok := usagePackageMap[field.Type]; ok {
			importsSet[field.Type] = "../" + strings.TrimSuffix(msg.FilePath, ".go")
			continue
		}

		if alias, exists := usageMap[field.ImportPath]; exists {
			field.UsageName = strings.ReplaceAll(field.UsageName, filepath.Base(field.ImportPath), alias)
		} else {
			if msg, exists := usagePackageMap[field.Type]; exists && msg.ImportPath != message.ImportPath {
				base = field.Name
				if path, exists := importsSet[base]; exists && path != msg.ImportPath {
					base = fmt.Sprintf("%s%d", base, i)
					usageMap[base] = msg.ImportPath
				}
				importsSet[base] = msg.ImportPath
				field.UsageName = base
			}
		}

	}
	var imports []Import
	for alias, pkg := range importsSet {
		imp := Import{
			Pkg:   pkg,
			Alias: alias,
		}
		imports = append(imports, imp)
	}
	return imports
}

func DetermineServiceImports(service *Service) []Import {
	var base string
	var svcImportsSets = make(map[string]string)
	// key 是 pkg
	var tmpMap = make(map[string]string)

	for i, method := range service.Methods {
		var inputAlias, outputAlias string

		if ia, exists := tmpMap[method.InputImportPackage]; exists {
			inputAlias = ia
			svcImportsSets[ia] = method.InputImportPackage
		} else {
			inputAlias = filepath.Base(method.InputImportPackage)
			if ia, exists = svcImportsSets[inputAlias]; exists {
				inputAlias = inputAlias + fmt.Sprintf("%d", i)
			}
			svcImportsSets[inputAlias] = method.InputImportPackage
			tmpMap[method.InputImportPackage] = inputAlias
		}
		method.InputUsageName = fmt.Sprintf("%s.%s", inputAlias, Last(strings.Split(method.InputUsageName, ".")))

		inputMsg, _ := usagePackageMap[Last(strings.Split(method.InputUsageName, "."))]
		if inputMsg != nil && HasComment(inputMsg.Comment, "@goType") {
			method.InputUsageName = GetComment(inputMsg.Comment, "@goType", method.InputUsageName)
		}

		outputMsg, _ := usagePackageMap[Last(strings.Split(method.OutputUsageName, "."))]
		if oa, exists := tmpMap[method.OutputImportPackage]; exists {
			outputAlias = oa
			svcImportsSets[oa] = method.OutputImportPackage
			tmpMap[method.OutputImportPackage] = outputAlias
		} else {
			outputAlias = filepath.Base(method.OutputImportPackage)
			if oa, exists = svcImportsSets[outputAlias]; exists {
				outputAlias = outputAlias + fmt.Sprintf("%d", i)
			}

			svcImportsSets[outputAlias] = method.OutputImportPackage
			tmpMap[method.OutputImportPackage] = outputAlias
		}
		method.OutputUsageName = fmt.Sprintf("%s.%s", outputAlias, Last(strings.Split(method.OutputUsageName, ".")))

		if outputMsg != nil && HasComment(outputMsg.Comment, "@goType") {
			method.OutputUsageName = GetComment(outputMsg.Comment, "@goType", method.OutputUsageName)
		}
	}

	var imports []Import

	for alias, pkg := range svcImportsSets {
		imp := Import{
			Pkg:   pkg,
			Alias: alias,
		}
		imports = append(imports, imp)
	}

	fmt.Println("DetermineServiceImports:", base, svcImportsSets, tmpMap)
	return imports
}

func DetermineTsServiceImports(service *Service) []Import {
	var base string
	svcImportsSet = make(map[string]string)

	for i, method := range service.Methods {
		if alias, exists := svcUsageMap[method.InputImportPackage]; exists {
			method.InputUsageName = strings.ReplaceAll(method.InputUsageName, filepath.Base(method.InputImportPackage), alias)
		} else {
			base = strings.Split(method.InputUsageName, ".")[1]
			method.InputUsageName = base
			method.InputImportPackage = fmt.Sprintf("../%s", strings.TrimSuffix(usagePackageMap[base].FilePath, ".go"))

			if importPackage, exists := svcImportsSet[base]; exists && importPackage != method.InputImportPackage {
				alias = fmt.Sprintf("%s%d", base, i)
				svcImportsSet[alias] = method.InputImportPackage
				method.InputUsageName = strings.ReplaceAll(method.InputUsageName, base, alias)
				svcUsageMap[method.InputImportPackage] = alias
			} else {
				svcImportsSet[base] = method.InputImportPackage
			}
		}

		if alias, exists := svcUsageMap[method.OutputImportPackage]; exists {
			method.OutputUsageName = strings.ReplaceAll(method.OutputUsageName, filepath.Base(method.OutputImportPackage), alias)
		} else {
			base = strings.Split(method.OutputUsageName, ".")[1]
			method.OutputUsageName = base
			method.OutputImportPackage = fmt.Sprintf("../%s", strings.TrimSuffix(usagePackageMap[base].FilePath, ".go"))

			if importPackage, exists := svcImportsSet[base]; exists && importPackage != method.OutputImportPackage {
				alias = fmt.Sprintf("%s%d", base, i)
				svcImportsSet[alias] = method.OutputImportPackage
				method.OutputUsageName = strings.ReplaceAll(method.OutputUsageName, base, alias)
				svcUsageMap[method.OutputImportPackage] = alias
			} else {
				svcImportsSet[base] = method.OutputImportPackage
			}
		}
	}
	var imports []Import
	for alias, pkg := range svcImportsSet {
		imp := Import{Pkg: pkg, Alias: alias}
		imports = append(imports, imp)
	}
	return imports
}

var svcImportsSet = make(map[string]string)
var svcUsageMap = map[string]string{}
