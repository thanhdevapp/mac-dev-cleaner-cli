package scanner

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/thanhdevapp/dev-cleaner/pkg/types"
)

// DockerSystemDF represents docker system df output
type DockerSystemDF struct {
	Type        string `json:"Type"`
	TotalCount  int    `json:"TotalCount"`
	Active      int    `json:"Active"`
	Size        string `json:"Size"`
	Reclaimable string `json:"Reclaimable"`
}

// parseDockerSize converts Docker size strings like "1.5GB" to bytes
func parseDockerSize(sizeStr string) int64 {
	sizeStr = strings.TrimSpace(sizeStr)
	if sizeStr == "" || sizeStr == "0B" {
		return 0
	}

	// Remove any parenthetical info like "(100%)"
	if idx := strings.Index(sizeStr, " "); idx > 0 {
		sizeStr = sizeStr[:idx]
	}

	var multiplier int64 = 1
	var value float64

	// Determine unit
	sizeStr = strings.ToUpper(sizeStr)
	if strings.HasSuffix(sizeStr, "KB") {
		multiplier = 1024
		sizeStr = strings.TrimSuffix(sizeStr, "KB")
	} else if strings.HasSuffix(sizeStr, "MB") {
		multiplier = 1024 * 1024
		sizeStr = strings.TrimSuffix(sizeStr, "MB")
	} else if strings.HasSuffix(sizeStr, "GB") {
		multiplier = 1024 * 1024 * 1024
		sizeStr = strings.TrimSuffix(sizeStr, "GB")
	} else if strings.HasSuffix(sizeStr, "TB") {
		multiplier = 1024 * 1024 * 1024 * 1024
		sizeStr = strings.TrimSuffix(sizeStr, "TB")
	} else if strings.HasSuffix(sizeStr, "B") {
		sizeStr = strings.TrimSuffix(sizeStr, "B")
	}

	// Parse numeric value
	_, err := fmt.Sscanf(sizeStr, "%f", &value)
	if err != nil {
		return 0
	}

	return int64(value * float64(multiplier))
}

// isDockerAvailable checks if Docker daemon is running
func isDockerAvailable() bool {
	cmd := exec.Command("docker", "info")
	err := cmd.Run()
	return err == nil
}

// ScanDocker scans for Docker artifacts using docker CLI
func (s *Scanner) ScanDocker() []types.ScanResult {
	var results []types.ScanResult

	// Check if Docker is available
	if !isDockerAvailable() {
		// Docker not installed or not running - skip silently
		return results
	}

	// Get Docker disk usage
	cmd := exec.Command("docker", "system", "df", "--format", "{{json .}}")
	output, err := cmd.Output()
	if err != nil {
		return results
	}

	// Parse each line of JSON output
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}

		var df DockerSystemDF
		if err := json.Unmarshal([]byte(line), &df); err != nil {
			continue
		}

		// Only include if there's reclaimable space
		reclaimSize := parseDockerSize(df.Reclaimable)
		if reclaimSize == 0 {
			continue
		}

		// Create result for each Docker resource type
		var name string
		switch df.Type {
		case "Images":
			name = "Docker Images (unused)"
		case "Containers":
			name = "Docker Containers (stopped)"
		case "Local Volumes":
			name = "Docker Volumes (unused)"
		case "Build Cache":
			name = "Docker Build Cache"
		default:
			name = "Docker " + df.Type
		}

		results = append(results, types.ScanResult{
			Path:      "docker:" + strings.ToLower(strings.ReplaceAll(df.Type, " ", "-")),
			Type:      types.TypeDocker,
			Size:      reclaimSize,
			FileCount: df.TotalCount - df.Active,
			Name:      name,
		})
	}

	return results
}
