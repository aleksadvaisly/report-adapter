package main

import (
	"regexp"
	"strconv"
	"strings"
)

var xmlTagPattern = regexp.MustCompile(`<[^>]+>`)

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func parseFloat(value string) float64 {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0
	}
	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0
	}
	return parsed
}

func parseTRXDuration(value string) float64 {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0
	}
	parts := strings.Split(value, ":")
	if len(parts) != 3 {
		return 0
	}
	hours := parseFloat(parts[0])
	minutes := parseFloat(parts[1])
	seconds := parseFloat(parts[2])
	return hours*3600 + minutes*60 + seconds
}

func stripXMLTags(value string) string {
	value = xmlTagPattern.ReplaceAllString(value, "")
	return strings.TrimSpace(value)
}

func intFromValue(value interface{}) int {
	switch typed := value.(type) {
	case float64:
		return int(typed)
	case int:
		return typed
	case int64:
		return int(typed)
	default:
		return 0
	}
}

func boolFromValue(value interface{}) (bool, bool) {
	switch typed := value.(type) {
	case bool:
		return typed, true
	default:
		return false, false
	}
}
