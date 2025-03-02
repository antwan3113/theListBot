package giflist

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestGifListPersistence(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "giflist_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Set environment variable for config path
	os.Setenv("GIFLIST_CONFIG_PATH", tempDir)
	defer os.Unsetenv("GIFLIST_CONFIG_PATH")

	// Create a new GifList
	list := NewGifList()

	// Add some test codes with multiple GIFs
	list.AddGif("t1", "https://test1.gif")
	list.AddGif("t1", "https://test1b.gif") // Second URL for same code
	list.AddGif("t2", "https://test2.gif")

	// Create a new list that should load from the same file
	list2 := NewGifList()

	// Verify the codes were loaded
	urls, found := list2.GetAllGifsForCode("t1")
	if !found {
		t.Error("Code t1 should be found but wasn't")
	}
	if len(urls) != 2 {
		t.Errorf("Code t1 should have 2 URLs, got %d", len(urls))
	}

	// Check if both URLs for t1 are present
	urlsFound := make(map[string]bool)
	for _, url := range urls {
		urlsFound[url] = true
	}
	if !urlsFound["https://test1.gif"] || !urlsFound["https://test1b.gif"] {
		t.Errorf("Not all URLs for t1 were loaded. Found: %v", urlsFound)
	}

	// Test removing a specific URL
	if removed := list2.RemoveGif("t1", "https://test1.gif"); !removed {
		t.Error("Failed to remove specific URL from code t1")
	}

	// Verify that only one URL remains for t1
	urls, found = list2.GetAllGifsForCode("t1")
	if !found || len(urls) != 1 || urls[0] != "https://test1b.gif" {
		t.Errorf("After removal, t1 should have only one URL. Found: %v", urls)
	}

	// Test removing an entire code
	if removed := list2.RemoveCode("t1"); !removed {
		t.Error("Failed to remove code t1")
	}

	// Verify code was completely removed
	urls, found = list2.GetAllGifsForCode("t1")
	if found || len(urls) > 0 {
		t.Errorf("Code t1 should be completely removed, found: %v", urls)
	}

	// Verify file format
	configFile := filepath.Join(tempDir, "gifcodes.json")
	data, err := os.ReadFile(configFile)
	if err != nil {
		t.Fatalf("Failed to read config file: %v", err)
	}

	var codeMap map[string][]string
	if err := json.Unmarshal(data, &codeMap); err != nil {
		t.Fatalf("Failed to parse config file: %v", err)
	}

	if len(codeMap["t2"]) != 1 || codeMap["t2"][0] != "https://test2.gif" {
		t.Errorf("Unexpected content in config file: %v", codeMap)
	}
}

func TestRandomSelectionFromMultipleGifs(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "giflist_random_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	os.Setenv("GIFLIST_CONFIG_PATH", tempDir)
	defer os.Unsetenv("GIFLIST_CONFIG_PATH")

	// Create GifList with many URLs for one code
	list := NewGifList()

	// Add test URLS
	testURLs := []string{
		"https://test-url-1.gif",
		"https://test-url-2.gif",
		"https://test-url-3.gif",
		"https://test-url-4.gif",
		"https://test-url-5.gif",
	}

	for _, url := range testURLs {
		list.AddGif("rx", url)
	}

	// Check all URLs were added
	urls, found := list.GetAllGifsForCode("rx")
	if !found {
		t.Error("Code rx should be found")
	}
	if len(urls) != len(testURLs) {
		t.Errorf("Expected %d URLs for rx, got %d", len(testURLs), len(urls))
	}

	// Test random selection by getting many samples
	// This isn't a perfect randomness test but can detect obvious issues
	selected := make(map[string]int)
	iterations := 100

	for i := 0; i < iterations; i++ {
		url, found := list.GetGif("rx")
		if !found {
			t.Fatal("GetGif failed to find code rx")
		}
		selected[url]++
	}

	// We should have selected most URLs at least once
	if len(selected) < 3 {
		t.Errorf("Random selection doesn't seem very random, only %d distinct URLs selected out of %d",
			len(selected), len(testURLs))
	}

	// Check if the distribution looks reasonable (not a perfect test)
	expectedAvg := float64(iterations) / float64(len(testURLs))
	for url, count := range selected {
		if count == 0 || float64(count) < expectedAvg*0.5 || float64(count) > expectedAvg*2.0 {
			t.Logf("Warning: URL %s has suspicious distribution: %d/%d selections",
				url, count, iterations)
		}
	}
}
