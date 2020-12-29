package utils

import "github.com/microcosm-cc/bluemonday"

// Sanitize string
func SanitizeString(str string) string {
	ugcPolicy := bluemonday.UGCPolicy()
	return ugcPolicy.Sanitize(str)
}
