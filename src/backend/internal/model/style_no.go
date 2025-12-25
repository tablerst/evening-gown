package model

import (
	"errors"
	"regexp"
	"strings"
)

// StyleNo is a human-facing model identifier.
// We normalize it to upper-case and restrict chars so it is safe for URLs and object keys.
//
// Allowed format:
// - A-Z, 0-9
// - optional '-' separators (no leading/trailing '-', no consecutive '--')
// Examples: 9001, AB-001, SS25-DR-01
const (
	StyleNoMaxLen = 64
)

var (
	errInvalidStyleNo = errors.New("invalid styleNo")
	styleNoRe         = regexp.MustCompile(`^[A-Z0-9]+(?:-[A-Z0-9]+)*$`)
)

func IsValidStyleNo(styleNo string) bool {
	styleNo = strings.TrimSpace(styleNo)
	if styleNo == "" || len(styleNo) > StyleNoMaxLen {
		return false
	}
	styleNo = strings.ToUpper(styleNo)
	return styleNoRe.MatchString(styleNo)
}

// NormalizeStyleNo trims and upper-cases a styleNo and validates it.
func NormalizeStyleNo(raw string) (string, error) {
	s := strings.TrimSpace(raw)
	if s == "" {
		return "", errInvalidStyleNo
	}
	if len(s) > StyleNoMaxLen {
		return "", errInvalidStyleNo
	}
	s = strings.ToUpper(s)
	if !styleNoRe.MatchString(s) {
		return "", errInvalidStyleNo
	}
	return s, nil
}
