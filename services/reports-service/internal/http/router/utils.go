package router

import "strings"

// parseAllowedOrigins mirrors auth-service: empty raw → ["*"]; otherwise trimmed CSV.
func parseAllowedOrigins(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return []string{"*"}
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		origin := strings.TrimSpace(part)
		if origin == "" {
			continue
		}
		out = append(out, origin)
	}
	if len(out) == 0 {
		return []string{"*"}
	}
	return out
}

// corsAllowCredentials is true when origins are explicit (browsers forbid credentials with "*").
func corsAllowCredentials(origins []string) bool {
	return len(origins) > 0 &&
		(len(origins) != 1 || origins[0] != "*")
}
