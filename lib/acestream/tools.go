package acestream

import "regexp"

func IsValidAceID(aceID string) bool {
	return regexp.MustCompile(`^[a-fA-F0-9]{40}$`).MatchString(aceID)
}
