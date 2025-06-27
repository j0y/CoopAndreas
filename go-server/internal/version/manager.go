package version

import (
	"fmt"
	"strconv"
	"strings"

	"coopandreas-server/internal/types"
)

// Manager handles version validation logic
type Manager struct {
	minClientVersion string
	maxClientVersion string
}

// NewManager creates a new version manager
func NewManager() *Manager {
	return &Manager{
		minClientVersion: types.MinClientVersion,
		maxClientVersion: types.ServerVersion, // Assume server version is max supported
	}
}

// VersionInfo represents parsed version information
type VersionInfo struct {
	Major int
	Minor int
	Patch int
	Tag   string // e.g., "alpha", "beta", "rc1"
}

// ParseVersion parses a version string like "0.1.1-alpha"
func (m *Manager) ParseVersion(version string) (*VersionInfo, error) {
	// Remove null bytes and trim
	cleanVersion := strings.TrimRight(version, "\x00")
	cleanVersion = strings.TrimSpace(cleanVersion)
	
	if cleanVersion == "" {
		return nil, fmt.Errorf("empty version string")
	}
	
	// Split on dash to separate version from tag
	parts := strings.Split(cleanVersion, "-")
	versionPart := parts[0]
	
	var tag string
	if len(parts) > 1 {
		tag = parts[1]
	}
	
	// Parse version numbers (x.y.z)
	versionNumbers := strings.Split(versionPart, ".")
	if len(versionNumbers) < 2 || len(versionNumbers) > 3 {
		return nil, fmt.Errorf("invalid version format: %s", cleanVersion)
	}
	
	major, err := strconv.Atoi(versionNumbers[0])
	if err != nil {
		return nil, fmt.Errorf("invalid major version: %s", versionNumbers[0])
	}
	
	minor, err := strconv.Atoi(versionNumbers[1])
	if err != nil {
		return nil, fmt.Errorf("invalid minor version: %s", versionNumbers[1])
	}
	
	patch := 0
	if len(versionNumbers) == 3 {
		patch, err = strconv.Atoi(versionNumbers[2])
		if err != nil {
			return nil, fmt.Errorf("invalid patch version: %s", versionNumbers[2])
		}
	}
	
	return &VersionInfo{
		Major: major,
		Minor: minor,
		Patch: patch,
		Tag:   tag,
	}, nil
}

// CompareVersions compares two version info structs
// Returns: -1 if v1 < v2, 0 if v1 == v2, 1 if v1 > v2
func (m *Manager) CompareVersions(v1, v2 *VersionInfo) int {
	if v1.Major != v2.Major {
		if v1.Major < v2.Major {
			return -1
		}
		return 1
	}
	
	if v1.Minor != v2.Minor {
		if v1.Minor < v2.Minor {
			return -1
		}
		return 1
	}
	
	if v1.Patch != v2.Patch {
		if v1.Patch < v2.Patch {
			return -1
		}
		return 1
	}
	
	// If version numbers are equal, compare tags
	return m.compareTags(v1.Tag, v2.Tag)
}

// compareTags compares version tags (alpha < beta < rc < release)
func (m *Manager) compareTags(tag1, tag2 string) int {
	tagWeight := map[string]int{
		"":      100, // release version
		"rc":    80,
		"beta":  60,
		"alpha": 40,
		"dev":   20,
	}
	
	weight1 := m.getTagWeight(tag1, tagWeight)
	weight2 := m.getTagWeight(tag2, tagWeight)
	
	if weight1 < weight2 {
		return -1
	}
	if weight1 > weight2 {
		return 1
	}
	return 0
}

// getTagWeight returns the weight of a tag for comparison
func (m *Manager) getTagWeight(tag string, weights map[string]int) int {
	tag = strings.ToLower(tag)
	
	// Check for exact matches first
	if weight, exists := weights[tag]; exists {
		return weight
	}
	
	// Check for partial matches (e.g., "rc1", "beta2")
	for tagType, weight := range weights {
		if tagType != "" && strings.HasPrefix(tag, tagType) {
			return weight
		}
	}
	
	// Unknown tag, treat as development version
	return weights["dev"]
}

// ValidateClientVersion checks if a client version is compatible
func (m *Manager) ValidateClientVersion(clientVersionStr string) (bool, string, error) {
	clientVersion, err := m.ParseVersion(clientVersionStr)
	if err != nil {
		return false, fmt.Sprintf("Invalid version format: %s", clientVersionStr), err
	}
	
	minVersion, err := m.ParseVersion(m.minClientVersion)
	if err != nil {
		return false, "Server configuration error", err
	}
	
	maxVersion, err := m.ParseVersion(m.maxClientVersion)
	if err != nil {
		return false, "Server configuration error", err
	}
	
	// Check if client version is too old
	if m.CompareVersions(clientVersion, minVersion) < 0 {
		return false, fmt.Sprintf("Client version too old. Minimum required: %s", m.minClientVersion), nil
	}
	
	// Check if client version is too new (future version)
	if m.CompareVersions(clientVersion, maxVersion) > 0 {
		return false, fmt.Sprintf("Client version too new. Maximum supported: %s", m.maxClientVersion), nil
	}
	
	// Version is compatible
	return true, "Version compatible", nil
}

// GetVersionString returns a formatted version string
func (m *Manager) GetVersionString(info *VersionInfo) string {
	version := fmt.Sprintf("%d.%d.%d", info.Major, info.Minor, info.Patch)
	if info.Tag != "" {
		version += "-" + info.Tag
	}
	return version
}

// IsDevVersion checks if this is a development version
func (m *Manager) IsDevVersion(version *VersionInfo) bool {
	devTags := []string{"dev", "alpha", "beta"}
	lowerTag := strings.ToLower(version.Tag)
	
	for _, devTag := range devTags {
		if strings.HasPrefix(lowerTag, devTag) {
			return true
		}
	}
	return false
}
