package server

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// GetPort returns the port from RDR_PORT env var or default
func GetPort() int {
	if port := os.Getenv("RDR_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			return p
		}
	}
	return 8080 // default
}

// GetPodCacheDuration returns cache duration from POD_CACHE_DURATION env var
// Default: 15 minutes
// Format: "15m", "30m", "1h", "5m30s"
func GetPodCacheDuration() time.Duration {
	if dur := os.Getenv("POD_CACHE_DURATION"); dur != "" {
		if d, err := time.ParseDuration(dur); err == nil {
			return d
		}
		fmt.Printf("Warning: Invalid POD_CACHE_DURATION format '%s', using default 15m\n", dur)
	}
	return 15 * time.Minute // default
}

// GetMaxRDBSize returns max RDB file size from MAX_RDB_SIZE env var
// Default: 10GB
// Format: "10Gb", "50Gb", "500Mb", "5Gb"
func GetMaxRDBSize() int64 {
	if size := os.Getenv("MAX_RDB_SIZE"); size != "" {
		if s, err := parseSize(size); err == nil {
			return s
		}
		fmt.Printf("Warning: Invalid MAX_RDB_SIZE format '%s', using default 10Gb\n", size)
	}
	return 10 * 1024 * 1024 * 1024 // 10GB default
}

// parseSize converts size string (e.g., "10Gb", "500Mb") to bytes
func parseSize(s string) (int64, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, fmt.Errorf("empty size string")
	}

	// Extract number and unit
	var num float64
	var unit string

	// Try different units
	units := []string{"Gb", "GB", "Mb", "MB", "Kb", "KB", "b", "B"}
	for _, u := range units {
		if strings.HasSuffix(s, u) {
			numStr := strings.TrimSuffix(s, u)
			var err error
			num, err = strconv.ParseFloat(numStr, 64)
			if err != nil {
				return 0, fmt.Errorf("invalid number: %v", err)
			}
			unit = strings.ToLower(u)
			break
		}
	}

	if unit == "" {
		return 0, fmt.Errorf("no unit specified (use Gb, Mb, Kb)")
	}

	// Convert to bytes
	var multiplier int64
	switch unit {
	case "gb":
		multiplier = 1024 * 1024 * 1024
	case "mb":
		multiplier = 1024 * 1024
	case "kb":
		multiplier = 1024
	case "b":
		multiplier = 1
	default:
		return 0, fmt.Errorf("unknown unit: %s", unit)
	}

	return int64(num * float64(multiplier)), nil
}

// FormatSize converts bytes to human-readable format
func FormatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGT"[exp])
}
