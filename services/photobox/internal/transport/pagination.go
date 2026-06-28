package transport

import "strconv"

const MaxPageLimit int32 = 100

const defaultPageLimit int32 = 10

func ParseQueryInt32(raw string, defaultVal int32) int32 {
	if raw == "" {
		return defaultVal
	}
	v, err := strconv.ParseInt(raw, 10, 32)
	if err != nil {
		return defaultVal
	}
	return int32(v)
}

func ClampPaginationFromQuery(limitRaw, offsetRaw string) (int32, int32) {
	return ClampPagination(
		ParseQueryInt32(limitRaw, defaultPageLimit),
		ParseQueryInt32(offsetRaw, 0),
	)
}

func ClampPagination(limit, offset int32) (int32, int32) {
	if limit <= 0 {
		limit = defaultPageLimit
	}
	if limit > MaxPageLimit {
		limit = MaxPageLimit
	}
	if offset < 0 {
		offset = 0
	}
	return limit, offset
}
