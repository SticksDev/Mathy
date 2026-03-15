package utils

// Common embed colors used across commands.
const (
	ColorSuccess = 0x57F287 // Green
	ColorError   = 0xED4245 // Red
	ColorWarn    = 0xFFAA00 // Amber
	ColorInfo    = 0x3498DB // Blue
	ColorPurple  = 0x9B59B6 // Purple
	ColorOrange  = 0xFF6600 // Orange
)

// Truncate shortens a string to maxLen, appending "..." if truncated.
func Truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
