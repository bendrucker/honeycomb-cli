package command

import (
	"fmt"
	"strings"
)

// ValidateEnum reports whether value is one of allowed, returning a consistent
// error naming the flag and listing the accepted values otherwise. An empty
// value passes; callers gate on presence separately.
func ValidateEnum(flag, value string, allowed []string) error {
	if value == "" {
		return nil
	}
	for _, a := range allowed {
		if value == a {
			return nil
		}
	}
	return fmt.Errorf("invalid --%s %q: must be one of %s", flag, value, strings.Join(allowed, ", "))
}
