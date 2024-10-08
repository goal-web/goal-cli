package gen

import (
	"fmt"
	"github.com/emicklei/proto"
	"github.com/goal-web/collection"
	"github.com/goal-web/supports/utils"
	"regexp"
	"strings"
	"unicode"
)

// ConvertCamelToSnake converts a string from CamelCase to snake_case
func ConvertCamelToSnake(str string) string {
	// Regular expression to find the positions where an uppercase letter is preceded by a lowercase letter
	var matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
	// Regular expression to find the positions where an uppercase letter is preceded by another uppercase letter and followed by a lowercase letter
	var matchAllCap = regexp.MustCompile("([a-z0-9])([A-Z])")

	// Insert underscores before capital letters
	snake := matchFirstCap.ReplaceAllString(str, "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")

	// Convert the entire string to lowercase
	snake = strings.ToLower(snake)

	// Split the string by underscores to identify the last word
	words := strings.Split(snake, "_")
	lastWord := words[len(words)-1]

	// Pluralize the last word
	words[len(words)-1] = pluralize(lastWord)

	// Join the words back together
	return strings.Join(words, "_")
}

func pluralize(word string) string {
	if strings.HasSuffix(word, "y") {
		// If the word ends with 'y', replaceSuffix 'y' with 'ies'
		return strings.TrimSuffix(word, "y") + "ies"
	} else if strings.HasSuffix(word, "s") || strings.HasSuffix(word, "x") || strings.HasSuffix(word, "z") {
		// If the word ends with 's', 'x', or 'z', add 'es'
		return word + "es"
	} else if strings.HasSuffix(word, "f") {
		// If the word ends with 'f', replaceSuffix 'f' with 'ves'
		return strings.TrimSuffix(word, "f") + "ves"
	} else if strings.HasSuffix(word, "fe") {
		// If the word ends with 'fe', replaceSuffix 'fe' with 'ves'
		return strings.TrimSuffix(word, "fe") + "ves"
	} else {
		// For most other cases, just add 's'
		return word + "s"
	}
}

func replaceSuffix(content string, trim ...string) string {
	for _, s := range trim {
		content = strings.TrimSuffix(content, s)
	}
	return content
}

// ToSnakeCase 将驼峰命名转换为蛇形命名
func ToSnakeCase(str string) string {
	var result []rune
	for i, r := range str {
		if i > 0 && unicode.IsUpper(r) && (unicode.IsLower(rune(str[i-1])) || (i+1 < len(str) && unicode.IsLower(rune(str[i+1])))) {
			result = append(result, '_')
		}
		result = append(result, unicode.ToLower(r))
	}
	return string(result)
}

// ToCamelCase 将字符串转换为大写开头的驼峰命名（PascalCase）
func ToCamelCase(s string) string {
	s = strings.ReplaceAll(s, "_", " ")
	s = strings.Title(s)
	s = strings.ReplaceAll(s, " ", "")
	return s
}

// GoType 将 Proto 类型映射为 Go 类型
func GoType(field *Field) string {

	if field.GoType != "" {
		return field.GoType
	}

	str := field.UsageName
	if str == "" {
		str = field.Type
	}

	if field.Ptr || field.IsModel {
		str = "*" + str
	}
	if field.Repeated {
		str = "[]" + str
	}
	return str
}

// FieldMsg 将 Proto 类型映射为 Go 类型
func FieldMsg(field *Field) *Message {
	return usagePackageMap[field.Type]
}

// SubString 切割字符串
func SubString(str string, start int, nums ...int) string {
	runes := []rune(str)
	strLen := len(runes)
	num := utils.DefaultValue(nums, strLen)
	if start >= strLen {
		return ""
	}
	if num < 0 {
		return string(runes[start : strLen+num])
	}
	if start+num >= strLen || num == 0 {
		return string(runes[start:])
	}
	return string(runes[start : start+num])
}

func HasComment(comment *proto.Comment, name string) bool {
	if comment != nil {
		for _, line := range comment.Lines {
			if strings.HasPrefix(line, name) {
				return true
			}
		}
	}
	return false
}

func Sub(v, t int) int {
	return v - t
}

func HasMsgComment(msg *Message, name string) bool {
	if msg.Comment != nil {
		if HasComment(msg.Comment, name) {
			return true
		}
	}
	for _, field := range msg.Fields {
		if HasComment(field.Comment, name) {
			return true
		}
	}
	for _, field := range msg.Relations {
		if HasComment(field.Comment, name) {
			return true
		}
	}

	return false
}

func GetComment(comment *proto.Comment, name string, defaultValue string) string {
	if comment != nil {
		for _, line := range comment.Lines {
			line = strings.TrimPrefix(line, " ")
			if strings.HasPrefix(line, name) {
				value := strings.TrimPrefix(strings.TrimPrefix(line, name), ":")
				if value == "" {
					value = defaultValue
				}
				return value
			}
		}
	}
	return defaultValue
}

func StringJoin(str ...string) string {
	var s string
	for _, s2 := range str {
		s += s2
	}
	return s
}

func GetIndexComment(comment *proto.Comment, name string, index int, defaultValue string) string {
	values := trim(strings.Split(GetComment(comment, name, ""), ",")...)
	if len(values) > index {
		return values[index]
	}
	return defaultValue
}

// ToTags 生成 tag
func ToTags(f *Field) string {
	var tags = []string{
		GetComment(f.Comment, "@goTag", ""),
	}

	if !strings.Contains(tags[0], "json:") {
		tags = append(tags, fmt.Sprintf(`json:"%s"`, f.JSONName))
	}

	if !strings.Contains(tags[0], "db:") {
		if f.Parent != nil && HasComment(f.Parent.Comment, "@timestamps") {
			createdAt := GetIndexComment(f.Parent.Comment, "@timestamps", 0, "created_at")
			updatedAt := GetIndexComment(f.Parent.Comment, "@timestamps", 1, "updated_at")

			if f.JSONName == createdAt {
				tags = append(tags, fmt.Sprintf(`db:"%s;type:timestamp;default CURRENT_TIMESTAMP;"`, f.JSONName))
			}
			if f.JSONName == updatedAt {
				tags = append(tags, fmt.Sprintf(`db:"%s;type:timestamp;DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP;"`, f.JSONName))
			}
		} else {
			tags = append(tags, fmt.Sprintf(
				`db:"%s;type:%s;not null;%s"`,
				f.JSONName,
				DBType(f),
				utils.IfString(HasComment(f.Comment, "@pk") || (f.Parent != nil && !HasMsgComment(f.Parent, "@pk") && f.Index == 0), "primary key", ""),
			),
			)
		}
	}

	return strings.TrimPrefix(strings.Join(tags, " "), " ")
}

func DBType(f *Field) string {
	if usagePackageMap[f.Type] != nil {
		return "json"
	} else if HasComment(f.Comment, "@carbon") {
		return "timestamp"
	} else if f.Parent != nil && HasComment(f.Parent.Comment, "@timestamps") {
		createdAt := GetIndexComment(f.Parent.Comment, "@timestamps", 0, "created_at")
		updatedAt := GetIndexComment(f.Parent.Comment, "@timestamps", 1, "updated_at")
		if f.JSONName == createdAt || f.JSONName == updatedAt {
			return "timestamp"
		}
	}

	var SQLTypeMap = map[string]string{
		"double":   "DOUBLE",          // 双精度浮点数
		"float":    "FLOAT",           // 单精度浮点数
		"int32":    "INT",             // 32 位整型
		"int64":    "BIGINT",          // 64 位整型
		"uint32":   "INT UNSIGNED",    // 32 位无符号整型
		"uint64":   "BIGINT UNSIGNED", // 64 位无符号整型
		"sint32":   "INT",             // 32 位有符号整型（优化负数）
		"sint64":   "BIGINT",          // 64 位有符号整型（优化负数）
		"fixed32":  "INT",             // 32 位固定长度整数
		"fixed64":  "BIGINT",          // 64 位固定长度整数
		"sfixed32": "INT",             // 32 位有符号固定长度整数
		"sfixed64": "BIGINT",          // 64 位有符号固定长度整数
		"bool":     "BOOLEAN",         // 布尔类型
		"string":   "VARCHAR(255)",    // 字符串类型
		"bytes":    "BLOB",            // 二进制数据
	}
	if value, ok := SQLTypeMap[f.Type]; ok {
		return value
	}
	return "varchar(255)"
}

var (
	// 记录类型到包的映射
	usagePackageMap = make(map[string]*Message)
)

func ToComments(name string, comments []string) string {
	if len(comments) == 0 {
		return ""
	}
	c := strings.Join(append([]string{
		fmt.Sprintf("// %s ", name),
	}, collection.New(comments).Each(func(i int, s string) string {
		if strings.HasPrefix(s, "//") {
			return s
		}
		return "// " + s
	}).ToArray()...), "\n")

	return c
}

func ToMiddlewares(middlewares []string) string {
	if len(middlewares) == 0 {
		return ""
	}

	return "," + strings.Join(collection.New(middlewares).Each(func(i int, s string) string {
		return fmt.Sprintf(`"%s"`, s)
	}).ToArray(), ",")
}

// NotContains 返回 true，如果字符串中不包含指定的子串
func NotContains(str, substr string) bool {
	return !strings.Contains(str, substr)
}
