package util

import (
	"regexp"
	"strings"
)

//We limit to four digits... Git commits tend to be only nummeric as well
var versionWithOptionalMinorAndPatch = regexp.MustCompile(`^[0-9]{1,5}(\.[0-9]+(\.[0-9]+)?)?(\+([0-9A-Za-z]+))?$`)
var versionWithMinorAndPatch = regexp.MustCompile(`^[0-9]+\.[0-9]+\.[0-9]+$|^[0-9]+\.[0-9]+\.[0-9]+\+([0-9A-Za-z]+)$`)
var versionMeta = regexp.MustCompile(`\+([0-9A-Za-z]+)$`)

//GetVersionWithoutMetadata :
func GetVersionWithoutMetadata(versionString string) string {
	matches := versionMeta.FindStringSubmatch(versionString)
	if matches == nil {
		return versionString
	}
	return strings.Replace(versionString, "+"+matches[1], "", -1)
}

//IsFullSemanticVersion :
func IsFullSemanticVersion(versionString string) bool {
	if versionWithMinorAndPatch.MatchString(versionString) {
		return true
	}
	return false
}
