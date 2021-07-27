package collector

import (
	"fmt"
	"strings"
)

func splitTopic(topic string) (string, string, string, error) {
	substrings := strings.Split(topic, "/")
	if len(substrings) < 4 || len(filterEmpty(substrings)) < 4 {
		return "", "", "", fmt.Errorf("invalid topic name '%s'", topic)
	}
	return substrings[1], substrings[2], substrings[3], nil
}

func filterEmpty(substrings []string) []string {
	newSubstrings := make([]string, 0, 4)
	for _, substring := range substrings {
		if len(strings.TrimSpace(substring)) != 0 {
			newSubstrings = append(newSubstrings, substring)
		}
	}
	return newSubstrings
}
