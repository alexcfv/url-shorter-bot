package validators

import "testing"

func TestIsValidURL(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "Valid HTTP URL",
			input:    "http://example.com",
			expected: true,
		},
		{
			name:     "Valid HTTPS URL",
			input:    "https://example.com",
			expected: true,
		},
		{
			name:     "Missing protocol",
			input:    "example.com",
			expected: false,
		},
		{
			name:     "Invalid protocol",
			input:    "ftp://example.com",
			expected: false,
		},
		{
			name:     "URL with path",
			input:    "http://example.com/path",
			expected: true,
		},
		{
			name:     "URL with query",
			input:    "http://example.com?query=1",
			expected: true,
		},
		{
			name:     "URL with port",
			input:    "http://example.com:8080",
			expected: true,
		},
		{
			name:     "Subdomain",
			input:    "https://sub.example.com",
			expected: true,
		},
		{
			name:     "Invalid characters",
			input:    "https://exam ple.com",
			expected: true,
		},
		{
			name:     "Trailing slash",
			input:    "https://example.com/",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidURL(tt.input)
			if result != tt.expected {
				t.Errorf("IsValidURL(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestShortToHash(t *testing.T) {
	tests := []struct {
		input string
	}{
		{"https://example.com"},
		{"https://example.com/page"},
		{"bebebbebebeb"},
		{"123456"},
		{""},
	}

	seen := make(map[uint32]string)

	for _, tt := range tests {
		hash := ShortToHash(tt.input)

		if prev, exists := seen[hash]; exists && prev != tt.input {
			t.Errorf("Hash collision: %q and %q both produce %d", prev, tt.input, hash)
		}
		seen[hash] = tt.input
	}
}

func TestShortToHash_Consistency(t *testing.T) {
	cases := []struct {
		input    string
		expected uint32
	}{
		{"https://example.com", ShortToHash("https://example.com")},
		{"test", ShortToHash("test")},
		{"", ShortToHash("")},
	}

	for _, c := range cases {
		got := ShortToHash(c.input)
		if got != c.expected {
			t.Errorf("ShortToHash(%q) = %v; want %v", c.input, got, c.expected)
		}
	}
}
