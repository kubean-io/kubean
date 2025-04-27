package util

import "strings"

func UnifyVersions(versions []string) []string {
	seen := make(map[string]struct{})
	result := make([]string, 0)

	for _, v := range versions {
		cleaned := strings.TrimPrefix(v, "v")
		if _, exists := seen[cleaned]; !exists {
			seen[cleaned] = struct{}{}
			result = append(result, cleaned)
		}
	}

	return result
}
