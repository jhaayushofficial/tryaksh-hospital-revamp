package logging

import "log/slog"

// MaskPhone reduces a phone number to its last four digits so that logs can be
// correlated with a patient without storing the full number.
//
//	+919229333922 -> ******3922
func MaskPhone(phone string) string {
	const keep = 4
	runes := []rune(phone)
	if len(runes) <= keep {
		return "****"
	}
	masked := make([]rune, 0, len(runes))
	for range runes[:len(runes)-keep] {
		masked = append(masked, '*')
	}
	return string(append(masked, runes[len(runes)-keep:]...))
}

// Phone builds a log attribute holding a masked phone number.
func Phone(key, phone string) slog.Attr {
	return slog.String(key, MaskPhone(phone))
}
