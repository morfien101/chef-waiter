package internalstate

import "regexp"

func extractVersion(in string) string {
	re := regexp.MustCompile(`([0-9]+\.[0-9]+\.[0-9]+)`)
	versionNumber := re.FindString(in)
	return versionNumber
}
