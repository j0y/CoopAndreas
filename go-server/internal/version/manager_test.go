package version

import (
	"testing"
)

func TestParseVersion(t *testing.T) {
	manager := NewManager()
	
	tests := []struct {
		input    string
		expected VersionInfo
		hasError bool
	}{
		{
			input: "0.1.1-alpha",
			expected: VersionInfo{
				Major: 0,
				Minor: 1,
				Patch: 1,
				Tag:   "alpha",
			},
			hasError: false,
		},
		{
			input: "1.2.3",
			expected: VersionInfo{
				Major: 1,
				Minor: 2,
				Patch: 3,
				Tag:   "",
			},
			hasError: false,
		},
		{
			input: "2.0-beta",
			expected: VersionInfo{
				Major: 2,
				Minor: 0,
				Patch: 0,
				Tag:   "beta",
			},
			hasError: false,
		},
		{
			input:    "invalid",
			expected: VersionInfo{},
			hasError: true,
		},
		{
			input:    "",
			expected: VersionInfo{},
			hasError: true,
		},
	}
	
	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			result, err := manager.ParseVersion(test.input)
			
			if test.hasError {
				if err == nil {
					t.Errorf("Expected error for input %s, but got none", test.input)
				}
				return
			}
			
			if err != nil {
				t.Errorf("Unexpected error for input %s: %v", test.input, err)
				return
			}
			
			if result.Major != test.expected.Major ||
				result.Minor != test.expected.Minor ||
				result.Patch != test.expected.Patch ||
				result.Tag != test.expected.Tag {
				t.Errorf("For input %s, expected %+v, got %+v", test.input, test.expected, *result)
			}
		})
	}
}

func TestCompareVersions(t *testing.T) {
	manager := NewManager()
	
	v0_1_0_alpha := &VersionInfo{Major: 0, Minor: 1, Patch: 0, Tag: "alpha"}
	v0_1_0_beta := &VersionInfo{Major: 0, Minor: 1, Patch: 0, Tag: "beta"}
	v0_1_0 := &VersionInfo{Major: 0, Minor: 1, Patch: 0, Tag: ""}
	v0_1_1 := &VersionInfo{Major: 0, Minor: 1, Patch: 1, Tag: ""}
	v0_2_0 := &VersionInfo{Major: 0, Minor: 2, Patch: 0, Tag: ""}
	v1_0_0 := &VersionInfo{Major: 1, Minor: 0, Patch: 0, Tag: ""}
	
	tests := []struct {
		name     string
		v1       *VersionInfo
		v2       *VersionInfo
		expected int
	}{
		{"same version", v0_1_0, v0_1_0, 0},
		{"alpha < beta", v0_1_0_alpha, v0_1_0_beta, -1},
		{"beta < release", v0_1_0_beta, v0_1_0, -1},
		{"patch difference", v0_1_0, v0_1_1, -1},
		{"minor difference", v0_1_1, v0_2_0, -1},
		{"major difference", v0_2_0, v1_0_0, -1},
		{"reverse comparison", v1_0_0, v0_2_0, 1},
	}
	
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := manager.CompareVersions(test.v1, test.v2)
			if result != test.expected {
				t.Errorf("Expected %d, got %d for %+v vs %+v", test.expected, result, test.v1, test.v2)
			}
		})
	}
}

func TestValidateClientVersion(t *testing.T) {
	manager := NewManager()
	
	tests := []struct {
		name         string
		clientVer    string
		expectValid  bool
		expectError  bool
	}{
		{"valid current version", "0.1.1", true, false},
		{"valid release version", "0.1.1", true, false},
		{"too old version", "0.0.1", false, false},
		{"future version", "1.0.0", false, false},
		{"invalid format", "invalid", false, true},
	}
	
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			valid, _, err := manager.ValidateClientVersion(test.clientVer)
			
			if test.expectError && err == nil {
				t.Errorf("Expected error for %s, but got none", test.clientVer)
			}
			
			if !test.expectError && err != nil {
				t.Errorf("Unexpected error for %s: %v", test.clientVer, err)
			}
			
			if valid != test.expectValid {
				t.Errorf("For %s, expected valid=%v, got valid=%v", test.clientVer, test.expectValid, valid)
			}
		})
	}
}
