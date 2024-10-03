package gen

import (
	"fmt"
	"github.com/emicklei/proto"
	"github.com/goal-web/collection"
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

	if field.Ptr {
		str = "*" + str
	}
	return str
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

func GetComment(comment *proto.Comment, name string, defaultValue string) string {
	if comment != nil {
		for _, line := range comment.Lines {
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

// ToTags 生成 tag
func ToTags(f *Field) string {
	if strings.Contains(f.Tags, "json:") {
		return f.Tags
	}
	if f.Tags != "" && !strings.HasPrefix(f.Tags, " ") {
		f.Tags = " " + f.Tags
	}
	return fmt.Sprintf(`json:"%s"%s`, f.JSONName, f.Tags)
}

var (
	// 记录类型到包的映射
	usagePackageMap = make(map[string]Message)
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
