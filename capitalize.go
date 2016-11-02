package capitalize

import (
	"strings"
)

//Format is used to capitalize a string
func Format(msg string) string {
	return strings.ToUpper(msg)
}
