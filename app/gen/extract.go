package gen

import (
	"fmt"
	"github.com/emicklei/proto"
	"path/filepath"
	"strings"
)

// Proto 数据结构定义
type Proto struct {
	Messages   map[string][]*Message
	Services   map[string]*ExtractServiceTemp
	Enums      []*Enum
	References []*Proto
}

func ExtractModel(msg *proto.Message, basePackage, dir string) *Message {
	// 处理模型和请求实体
	var fields []*Field
	var relations []*Field
	var primaryKey string
	for _, element := range msg.Elements {
		if field, ok := element.(*proto.NormalField); ok {
			if primaryKey == "" {
				primaryKey = field.Name
			}

			var fieldItem = &Field{
				Repeated:  field.Repeated,
				Comment:   field.Comment,
				Name:      ToCamelCase(field.Name),
				Type:      field.Type,
				JSONName:  field.Name,
				UsageName: field.Type,
				GoType:    GetComment(field.Comment, "@goType", ""),
			}

			if field.Comment != nil {
				var commentTexts []string
				for _, line := range field.Comment.Lines {
					if strings.HasPrefix(line, "@gotag:") {
						fieldItem.Tags = strings.TrimPrefix(line, "@gotag:")
					} else if strings.HasPrefix(line, "@pk") {
						primaryKey = field.Name
					} else if strings.HasPrefix(line, "@ptr") {
						fieldItem.Ptr = true
					} else {
						commentTexts = append(commentTexts, "//"+line)
					}
				}

				if len(commentTexts) > 0 {
					fieldItem.Comments = strings.Join(commentTexts, "\n")
				}
			}
			if (HasComment(field.Comment, "@belongsTo") ||
				HasComment(field.Comment, "@hasOne") ||
				HasComment(field.Comment, "@hasMany")) && strings.HasSuffix(field.Type, "Model") {
				fieldItem.IsModel = true
				relations = append(relations, fieldItem)
			} else {
				fields = append(fields, fieldItem)
			}
		}
	}
	var midDir, tmlp = "models", "model"

	importPath := strings.Join(trim(basePackage, midDir), "/")
	usageName := fmt.Sprintf("%s.%s", filepath.Base(importPath), msg.Name)

	message := Message{
		IsModel:         true,
		Comment:         msg.Comment,
		Template:        tmlp,
		PrimaryKey:      primaryKey,
		TableName:       ConvertCamelToSnake(replaceSuffix(msg.Name, "Model")),
		RawName:         replaceSuffix(msg.Name, "Model"),
		Name:            msg.Name,
		Fields:          fields,
		Relations:       relations,
		ImportPath:      importPath,
		UsageName:       usageName,
		Authenticatable: HasComment(msg.Comment, "@authenticatable"),
		FilePath:        strings.Join(trim(midDir, replaceSuffix(msg.Name, "Model")+"_gen.go"), "/"),
	}

	if msg.Comment != nil {
		for _, line := range msg.Comment.Lines {
			if strings.HasPrefix(line, "@table:") {
				message.TableName = strings.Trim(strings.TrimPrefix(line, "@table:"), " ")
			} else if strings.HasPrefix(line, "@softDelete") {
				message.SoftDelete = strings.Trim(strings.TrimPrefix(line, "@softDelete"), ": ")
			} else {
				message.Comments = append(message.Comments, "//"+line)
			}
		}
	}
	usagePackageMap[msg.Name] = &message
	return &message
}

// 避免死循环解析
var parsedProtoMap = make(map[string]*Proto)

func ExtractProto(pwd string, def *proto.Proto, basePackage string, dir string) *Proto {
	var models []*Message
	var dataList []*Message
	var requests []*Message
	var results []*Message
	var references []*Proto
	var data *Proto
	parsedProtoMap[def.Filename] = data

	for _, element := range def.Elements {
		switch v := element.(type) {
		case *proto.Option:
			if v.Name == "go_package" {
				dir = v.Constant.Source
				fmt.Printf("读取到包名：%s\n", dir)
			}
		case *proto.Import:
			protoFile := filepath.Join(pwd, v.Filename)
			if subPro, exists := parsedProtoMap[protoFile]; exists {
				references = append(references, subPro)
			} else {
				subProf := ParseProto(protoFile)
				subPro = ExtractProto(pwd, subProf, basePackage, dir)
				parsedProtoMap[def.Filename] = subPro
				references = append(references, subPro)
			}
		}
	}

	// 遍历 proto 文件的元素
	for _, elem := range def.Elements {
		if e, ok := elem.(*proto.Message); ok {

			var midDir, tmlp string
			if strings.HasSuffix(e.Name, "Model") {
				models = append(models, ExtractModel(e, basePackage, dir))
				continue
			} else if strings.HasSuffix(e.Name, "Req") || strings.HasSuffix(e.Name, "Request") {
				midDir = "requests"
				tmlp = "request"
			} else if strings.HasSuffix(e.Name, "Result") {
				midDir = "results"
				tmlp = "result"
			} else {
				midDir = "models"
				tmlp = "data"
			}

			// 处理模型和请求实体
			var fields []*Field
			var primaryKey string
			for _, element := range e.Elements {
				if field, ok := element.(*proto.NormalField); ok {
					if primaryKey == "" {
						primaryKey = field.Name
					}

					var fieldItem = &Field{
						Repeated:  field.Repeated,
						Comment:   field.Comment,
						Name:      ToCamelCase(field.Name),
						Type:      field.Type,
						JSONName:  field.Name,
						UsageName: field.Type,
						GoType:    GetComment(field.Comment, "@goType", ""),
					}
					if field.Comment != nil {
						var commentTexts []string
						for _, line := range field.Comment.Lines {
							if strings.HasPrefix(line, "@gotag:") {
								fieldItem.Tags = " " + strings.TrimPrefix(line, "@gotag:")
							} else if strings.HasPrefix(line, "@ptr") {
								fieldItem.Ptr = true
							} else {
								commentTexts = append(commentTexts, "//"+line)
							}
						}

						if len(commentTexts) > 0 {
							fieldItem.Comments = strings.Join(commentTexts, "\n")
						}
					}
					fields = append(fields, fieldItem)
				}
			}

			importPath := strings.Join(trim(basePackage, midDir, dir), "/")
			usageName := fmt.Sprintf("%s.%s", filepath.Base(importPath), e.Name)

			msg := Message{
				Comment:    e.Comment,
				Template:   tmlp,
				PrimaryKey: primaryKey,
				TableName:  ConvertCamelToSnake(replaceSuffix(e.Name, "Request", "Req")),
				RawName:    replaceSuffix(e.Name, "Request", "Req"),
				Name:       e.Name,
				Fields:     fields,
				ImportPath: importPath,
				UsageName:  usageName,
				FilePath:   strings.Join(trim(midDir, dir, replaceSuffix(e.Name, "Model", "Request", "Req")+"_gen.go"), "/"),
			}

			usagePackageMap[e.Name] = &msg

			if strings.HasSuffix(e.Name, "Req") || strings.HasSuffix(e.Name, "Request") {
				requests = append(requests, &msg)
			} else if strings.HasSuffix(e.Name, "Result") {
				results = append(results, &msg)
			} else {
				dataList = append(dataList, &msg)
			}
		}
	}

	// 返回提取的数据
	data = &Proto{
		Messages: map[string][]*Message{
			"models":   models,
			"dataList": dataList,
			"requests": requests,
			"results":  results,
		},
		Services:   ExtractServices(def, basePackage, dir),
		Enums:      ExtractEnums(def, basePackage, dir),
		References: references,
	}
	return data
}

func ExtractServices(def *proto.Proto, basePackage string, dir string) map[string]*ExtractServiceTemp {
	var services = map[string]*ExtractServiceTemp{
		"services": {
			Suffix:   "Service",
			Template: "service",
			List:     make([]*Service, 0),
		},
		"controllers": {
			Suffix:   "Controller",
			Template: "controller",
			List:     make([]*Service, 0),
		},
	}
	for _, elem := range def.Elements {
		if e, ok := elem.(*proto.Service); ok {
			for path, temp := range services {
				if strings.HasSuffix(e.Name, temp.Suffix) {
					var methods []*Method
					for _, se := range e.Elements {
						if rpc, ok := se.(*proto.RPC); ok {

							method := &Method{
								Name:                rpc.Name,
								InputUsageName:      usagePackageMap[rpc.RequestType].UsageName,
								InputImportPackage:  usagePackageMap[rpc.RequestType].ImportPath,
								OutputUsageName:     usagePackageMap[rpc.ReturnsType].UsageName,
								OutputImportPackage: usagePackageMap[rpc.ReturnsType].ImportPath,

								Method:      strings.Split(GetComment(rpc.Comment, "@method", "Post"), ","),
								Path:        GetComment(rpc.Comment, "@path", fmt.Sprintf("/%s", rpc.Name)),
								Middlewares: getComments(rpc.Comment, "@middleware", ""),
							}

							methods = append(methods, method)
						}
					}

					importPath := strings.Join(trim(basePackage, path, dir), "/")
					usageName := fmt.Sprintf("%s.%s", filepath.Base(importPath), e.Name)

					temp.List = append(temp.List, &Service{
						Middlewares: getComments(e.Comment, "@middleware", ""),
						Controller:  HasComment(e.Comment, "@controller"),
						Prefix:      GetComment(e.Comment, "@controller", ""),

						Name:        e.Name,
						Methods:     methods,
						Template:    temp.Template,
						PackageName: filepath.Base(importPath),
						ImportPath:  importPath,
						UsageName:   usageName,
						Filename:    strings.Join(trim(path, dir, replaceSuffix(e.Name, temp.Suffix)+"_gen.go"), "/"),
					})
				}
			}

		}
	}
	return services
}
