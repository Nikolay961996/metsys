package buildinfo

import "fmt"

var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

func PrintHello() {
	fmt.Printf("Build version: %s\n", defaultIfEmpty(buildVersion))
	fmt.Printf("Build date: %s\n", defaultIfEmpty(buildDate))
	fmt.Printf("Build commit: %s\n", defaultIfEmpty(buildCommit))
}

func defaultIfEmpty(s string) string {
	if s == "" {
		return "N/A"
	}
	return s
}
