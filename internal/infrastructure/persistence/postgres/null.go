package postgres

func nullableString(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}
