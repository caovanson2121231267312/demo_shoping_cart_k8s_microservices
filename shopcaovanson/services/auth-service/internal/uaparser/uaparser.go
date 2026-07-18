package uaparser

import (
	"strings"
)

// Parse extracts coarse browser / OS / device hints from a User-Agent string.
func Parse(ua string) (browser, osName, device string) {
	ua = strings.TrimSpace(ua)
	if ua == "" {
		return "Unknown", "Unknown", "Unknown"
	}
	lower := strings.ToLower(ua)

	switch {
	case strings.Contains(lower, "edg/"):
		browser = "Edge"
	case strings.Contains(lower, "chrome/") && !strings.Contains(lower, "edg/"):
		browser = "Chrome"
	case strings.Contains(lower, "firefox/"):
		browser = "Firefox"
	case strings.Contains(lower, "safari/") && !strings.Contains(lower, "chrome/"):
		browser = "Safari"
	case strings.Contains(lower, "opr/") || strings.Contains(lower, "opera"):
		browser = "Opera"
	default:
		browser = "Other"
	}

	switch {
	case strings.Contains(lower, "windows"):
		osName = "Windows"
	case strings.Contains(lower, "android"):
		osName = "Android"
	case strings.Contains(lower, "iphone") || strings.Contains(lower, "ipad") || strings.Contains(lower, "ios"):
		osName = "iOS"
	case strings.Contains(lower, "mac os") || strings.Contains(lower, "macintosh"):
		osName = "macOS"
	case strings.Contains(lower, "linux"):
		osName = "Linux"
	default:
		osName = "Other"
	}

	switch {
	case strings.Contains(lower, "mobile") || strings.Contains(lower, "android") || strings.Contains(lower, "iphone"):
		device = "Mobile"
	case strings.Contains(lower, "ipad") || strings.Contains(lower, "tablet"):
		device = "Tablet"
	default:
		device = "Desktop"
	}
	return browser, osName, device
}
