package conv

import (
	"strconv"
	"strings"
)

func StringToInt64(s string) (int64, error) {
	newData, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, err
	}

	return newData, nil
}

func GenerateSlug(name string) string {
	slug := strings.ToLower(name)
	slug = strings.ReplaceAll(slug, " ", "-")
	return slug
}


func StringToUUI(s string) string {
	if s == "" {
		return ""
	}
	s = strings.TrimSpace(s)
	if len(s) < 36 {
		return ""
	}
	return s[:36]
}

func BoolToString(b bool) string {
    if b {
        return "Published"
    } else {
        return "Unpublished"
    }
}

