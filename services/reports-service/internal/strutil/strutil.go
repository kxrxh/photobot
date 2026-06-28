package strutil

import "strings"

func TrimPtr(p *string) string {
	if p == nil {
		return ""
	}
	return strings.TrimSpace(*p)
}

func TrimNonEmpty(items []string) []string {
	if len(items) == 0 {
		return nil
	}
	out := make([]string, 0, len(items))
	for _, item := range items {
		if s := strings.TrimSpace(item); s != "" {
			out = append(out, s)
		}
	}
	return out
}

func EmptyAsDash(s string) string {
	if strings.TrimSpace(s) == "" {
		return "-"
	}
	return s
}

func PtrOrDash(p *string) string {
	if p == nil || strings.TrimSpace(*p) == "" {
		return "-"
	}
	return *p
}

func IsHTTPURL(u string) bool {
	u = strings.ToLower(strings.TrimSpace(u))
	return strings.HasPrefix(u, "http://") || strings.HasPrefix(u, "https://")
}

func IsDataURL(u string) bool {
	return strings.HasPrefix(strings.ToLower(strings.TrimSpace(u)), "data:")
}
