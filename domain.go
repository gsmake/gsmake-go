package gsmake

import "strings"

//  domain strings
const (
	DomainDefault = "task|golang"
	DomainTask    = "task"
)

// ParseDomain domain parse
func ParseDomain(src string, domainDefault string) []string {

	if src == "" {
		src = domainDefault
	}

	return strings.Split(src, "|")
}
