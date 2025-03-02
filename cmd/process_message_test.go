package main

import (
	"fmt"
	"regexp"
	"testing"
)

func TestProcessMessage(t *testing.T) {
	// Define test messages
	messages := []string{
		"gg",
		"GG everyone!",
		"ty for the help",
		" gg",           // has a space at start, should not match
		"hello gg",      // code not at start
		":gg:",          // starts with punctuation
		"!gg!",          // starts with punctuation
		"I'm saying gg", // not at start
		"gglong",        // part of longer word but at start - will match the prefix
		"notacode",
	}

	// Updated regex pattern to only match at the beginning of messages
	codePattern := regexp.MustCompile(`^(?i)([a-zA-Z]{2})(\b|$|[^a-zA-Z])`)

	fmt.Println("Testing start-of-message code pattern matching:")

	for _, msg := range messages {
		match := codePattern.FindStringSubmatch(msg)

		fmt.Printf("Message: %q\n", msg)
		if match == nil {
			fmt.Println("  No match at start of message")
		} else {
			if len(match) >= 2 {
				fmt.Printf("  Matched code at start: %q\n", match[1])
			} else {
				fmt.Printf("  Match found but invalid capture group\n")
			}
		}
		fmt.Println()
	}

	t.Skip("This is a debugging tool, not a real test")
}
