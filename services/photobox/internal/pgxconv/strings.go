package pgxconv

// NonEmptyStringPtr returns nil for empty strings (optional query / sort fields).
func NonEmptyStringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
