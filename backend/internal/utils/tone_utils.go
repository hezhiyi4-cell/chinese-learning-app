
package utils

import (
	"strings"
	"unicode"

	"github.com/mozillazg/go-pinyin"
)

// ToPinyin 把汉字转为带声调的拼音
// 例如：ToPinyin("你好") 返回 ["nǐ", "hǎo"]
func ToPinyin(text string) []string {
	a := pinyin.NewArgs()
	a.Style = pinyin.Tone

	var result []string
	for _, r := range text {
		if !unicode.Is(unicode.Han, r) {
			continue
		}
		pinYinResult := pinyin.Pinyin(string(r), a)
		if len(pinYinResult) > 0 && len(pinYinResult[0]) > 0 {
			result = append(result, pinYinResult[0][0])
		}
	}
	return result
}

// ExtractCharacters 只提取汉字
func ExtractCharacters(text string) []string {
	var result []string
	for _, r := range text {
		if unicode.Is(unicode.Han, r) {
			result = append(result, string(r))
		}
	}
	return result
}

// ToneError 表示一个声调错误
type ToneError struct {
	Index    int    `json:"index"`
	Expected string `json:"expected"`
	Actual   string `json:"actual"`
	ErrorType string `json:"errorType"`
}

// CompareTones 比较两个拼音序列，找出错误
func CompareTones(expectedChars []string, expectedPinyin []string, actualText string) []ToneError {
	var errors []ToneError
	
	actualPinyin := ToPinyin(actualText)
	actualChars := ExtractCharacters(actualText)

	maxLen := min(len(expectedChars), len(actualChars))
	
	for i := 0; i < maxLen; i++ {
		expected := expectedPinyin[i]
		actual := actualPinyin[i]
		
		if expected != actual {
			// 判断是声调错误还是其他错误
			expBase := strings.TrimRight(expected, "12345")
			actBase := strings.TrimRight(actual, "12345")
			
			errorType := "tone"
			if expBase != actBase {
				errorType = "initial_final"
			}
			
			errors = append(errors, ToneError{
				Index:    i,
				Expected: expected,
				Actual:   actual,
				ErrorType: errorType,
			})
		}
	}

	return errors
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// NormalizeText 标准化文本，只保留汉字
func NormalizeText(text string) string {
	var result []rune
	for _, r := range text {
		if unicode.Is(unicode.Han, r) {
			result = append(result, r)
		}
	}
	return string(result)
}

// CalculateScore 计算发音分数 (0-100)
func CalculateScore(expectedLen int, errors []ToneError) int {
	if expectedLen == 0 {
		return 0
	}
	
	correctCount := expectedLen - len(errors)
	if correctCount < 0 {
		correctCount = 0
	}
	
	score := int(float64(correctCount) / float64(expectedLen) * 100)
	return score
}
