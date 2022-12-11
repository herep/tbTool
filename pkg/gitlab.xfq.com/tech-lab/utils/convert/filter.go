package convert

import (
	"regexp"
)

// 过滤emoji
func RemoveEmoji(str string) string {
	reg := regexp.MustCompile(`[\x{1F600}-\x{1F6FF}|[\x{2600}-\x{26FF}]`)
	return reg.ReplaceAllString(str, "")
}
