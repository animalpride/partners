package middleware

// MaskToken returns the first 8 characters of a token followed by "..." for
// safe logging. If the token is 8 characters or shorter it is returned as-is.
func MaskToken(token string) string {
	if len(token) <= 8 {
		return token
	}
	return token[:8] + "..."
}
