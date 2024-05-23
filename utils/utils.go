package utils

import (
	"strings"
	"unicode"
)

// ModelNameToTable converts a CamelCase model name to a snake_case table name
// with the last word in plural form.
func ModelNameToTable(modelName string) string {
	var result []rune
	var word []rune

	// Iterate through each character in the model name
	for i, r := range modelName {
		if unicode.IsUpper(r) {
			if i != 0 {
				result = append(result, '_')
			}
			// Append the current word to the result
			result = append(result, word...)
			// Clear the word
			word = []rune{unicode.ToLower(r)}
		} else {
			word = append(word, r)
		}
	}
	// Append the last word to the result
	result = append(result, word...)

	// Split the snake_case string by underscore
	parts := strings.Split(string(result), "_")

	// Pluralize the last word
	if len(parts) > 0 {
		parts[len(parts)-1] = ToPlural(parts[len(parts)-1])
	}

	// Join the parts back with underscores
	return strings.Join(parts, "_")
}

// ToPlural converts a singular noun to its plural form
func ToPlural(word string) string {
	// Simple pluralization rules
	if strings.HasSuffix(word, "y") && !IsVowel(word[len(word)-2]) {
		return word[:len(word)-1] + "ies"
	} else if strings.HasSuffix(word, "s") || strings.HasSuffix(word, "sh") || strings.HasSuffix(word, "ch") || strings.HasSuffix(word, "x") || strings.HasSuffix(word, "z") {
		return word + "es"
	} else {
		return word + "s"
	}
}

// IsVowel checks if a character is a vowel
func IsVowel(ch byte) bool {
	vowels := "aeiou"
	return strings.ContainsRune(vowels, rune(ch))
}

// CamelToColonHyphen converts a CamelCase string to a string
// where the first transition is marked with a colon and subsequent
// transitions are marked with hyphens. Consecutive uppercase letters
// are treated as a single word.
func CamelToColonHyphen(s string) string {
	var result []rune
	firstTransition := true
	consecutiveUpper := false

	for i, r := range s {
		if unicode.IsUpper(r) {
			if i != 0 && !consecutiveUpper {
				if firstTransition {
					result = append(result, ':')
					firstTransition = false
				} else {
					result = append(result, '-')
				}
			}
			result = append(result, unicode.ToLower(r))
			consecutiveUpper = true
		} else {
			result = append(result, r)
			consecutiveUpper = false
		}
	}
	return string(result)
}

// ConvertToCamelCase 将蛇形字符串转换为大写开头的驼峰式字符串
func ConvertToCamelCase(s string) string {
	// 分割字符串并转换为驼峰形式
	words := strings.Split(s, "_")
	for i, word := range words {
		words[i] = strings.Title(word)
	}
	result := strings.Join(words, "")

	return result
}
