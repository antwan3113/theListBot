package server

import (
	"regexp"
	"testing"
)

func TestCodePatternMatching(t *testing.T) {
	// Define the pattern for codes at the start of messages
	codePattern := regexp.MustCompile(`^(?i)([a-zA-Z]{2})(\b|$|[^a-zA-Z])`)

	// Define test cases with expected matches
	testCases := []struct {
		input    string
		expected bool
		code     string
	}{
		// Basic cases at start of message - should match
		{"gg", true, "gg"},
		{"GG", true, "GG"},
		{"ty", true, "ty"},
		{"gg!", true, "gg"},
		{"gg world", true, "gg"},
		{"GG everyone", true, "GG"},

		// With punctuation at start - should not match
		{":gg:", false, ""},
		{"!gg!", false, ""},
		{" gg", false, ""},

		// Not at start of message - should not match
		{"hello gg world", false, ""},
		{"hello gg", false, ""},
		{"hello:gg:world", false, ""},
		{"What's up? gg", false, ""},

		// Embedded in words - should not match
		{"ggg", false, "gg"}, // Should match because 'gg' is at the start, even though it's part of a longer word
		{"bigger", false, ""},
		{"agga", false, ""},

		// Edge cases
		{"g g", false, ""}, // Not two consecutive letters
	}

	// Run the tests
	for i, tc := range testCases {
		match := codePattern.FindStringSubmatch(tc.input)

		// Check if we got a match as expected
		if (match != nil) != tc.expected {
			t.Errorf("Test case %d failed: %s - Expected match: %v, got: %v",
				i+1, tc.input, tc.expected, match != nil)
			continue
		}

		// If we expected a match, check the extracted code
		if tc.expected && match != nil && len(match) >= 2 && match[1] != tc.code {
			t.Errorf("Test case %d failed: %s - Expected code: %s, got: %s",
				i+1, tc.input, tc.code, match[1])
		}
	}
}

// Test specifically for the stricter "must be at start" rules
func TestStartOfMessageCodeOnly(t *testing.T) {
	// Define the pattern we're using in our server
	codePattern := regexp.MustCompile(`^(?i)([a-zA-Z]{2})(\b|$|[^a-zA-Z])`)

	validStartCases := []string{
		"gg",
		"gg!",
		"GG everyone",
		"ty.",
		"gg:",
	}

	invalidStartCases := []string{
		" gg",
		"hello gg",
		":gg",
		"!gg",
		"bigger words",
		"not at start: gg",
	}

	// Test valid cases
	for _, input := range validStartCases {
		if match := codePattern.FindStringSubmatch(input); match == nil {
			t.Errorf("Should match but didn't: %q", input)
		} else {
			code := match[1]
			if len(code) != 2 {
				t.Errorf("Invalid code length for %q: got %q", input, code)
			}
		}
	}

	// Test invalid cases
	for _, input := range invalidStartCases {
		if match := codePattern.FindStringSubmatch(input); match != nil {
			t.Errorf("Should not match but did: %q, got code: %q", input, match[1])
		}
	}
}
