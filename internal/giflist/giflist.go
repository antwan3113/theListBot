package giflist

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Initialize random seed on package init
func init() {
	rand.Seed(time.Now().UnixNano())
}

// GifList manages the mappings between 2-character codes and GIF URLs
type GifList struct {
	codeMap    map[string][]string // Changed from map[string]string to map[string][]string
	mutex      sync.RWMutex
	configFile string
}

// NewGifList creates a new GifList and loads mappings from a file if available
func NewGifList() *GifList {
	log.Println("Initializing GifList")

	// Determine configuration directory and file
	configDir := getConfigDir()
	configFile := filepath.Join(configDir, "gifcodes.json")

	list := &GifList{
		codeMap:    make(map[string][]string), // Changed to []string
		configFile: configFile,
	}

	// Ensure the config directory exists
	if err := os.MkdirAll(configDir, 0755); err != nil {
		log.Printf("Warning: Failed to create config directory: %v", err)
	}

	// Try to load existing mappings
	if err := list.LoadFromFile(); err != nil {
		log.Printf("No existing mappings found or error loading: %v", err)
		log.Println("Adding default example mappings")

		// Add some example mappings
		list.AddGif("gg", "https://media.giphy.com/media/3o7abldj0b3rxrZUxW/giphy.gif")
		list.AddGif("gg", "https://media.giphy.com/media/12XDYvMJNcmLgQ/giphy.gif") // Second URL for same code
		list.AddGif("ty", "https://media.giphy.com/media/KB8C86UMgLDThpt4WT/giphy.gif")

		// Save the default mappings
		if err := list.SaveToFile(); err != nil {
			log.Printf("Warning: Failed to save default mappings: %v", err)
		}
	}

	// Log the number of codes and total GIFs
	totalGifs := 0
	for _, urls := range list.codeMap {
		totalGifs += len(urls)
	}
	log.Printf("GifList initialized with %d codes and %d total GIFs", len(list.codeMap), totalGifs)

	return list
}

// getConfigDir determines the appropriate configuration directory
func getConfigDir() string {
	// Check for explicit config path in environment
	if configPath := os.Getenv("GIFLIST_CONFIG_PATH"); configPath != "" {
		return configPath
	}

	// Default to user home directory or current directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Printf("Could not determine home directory: %v", err)
		return "."
	}

	return filepath.Join(homeDir, ".thelistbot")
}

// LoadFromFile loads the code mappings from the JSON file
func (g *GifList) LoadFromFile() error {
	g.mutex.Lock()
	defer g.mutex.Unlock()

	// Check if the file exists
	if _, err := os.Stat(g.configFile); os.IsNotExist(err) {
		return fmt.Errorf("config file does not exist: %s", g.configFile)
	}

	// Read the file
	data, err := os.ReadFile(g.configFile)
	if err != nil {
		return fmt.Errorf("failed to read config file: %v", err)
	}

	// Parse the JSON
	if err := json.Unmarshal(data, &g.codeMap); err != nil {
		// Try to load legacy format (single URL per code)
		var legacyMap map[string]string
		if legacyErr := json.Unmarshal(data, &legacyMap); legacyErr == nil {
			log.Println("Detected legacy format, converting to multi-gif format")
			for code, url := range legacyMap {
				g.codeMap[code] = []string{url}
			}
		} else {
			return fmt.Errorf("failed to parse config file: %v", err)
		}
	}

	// Count total GIFs for logging
	totalGifs := 0
	for _, urls := range g.codeMap {
		totalGifs += len(urls)
	}
	log.Printf("Loaded %d code mappings with %d total GIFs from %s", len(g.codeMap), totalGifs, g.configFile)

	return nil
}

// SaveToFile saves the current code mappings to the JSON file
func (g *GifList) SaveToFile() error {
	g.mutex.RLock()
	defer g.mutex.RUnlock()

	// Serialize to JSON
	data, err := json.MarshalIndent(g.codeMap, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialize code mappings: %v", err)
	}

	// Write to file
	if err := os.WriteFile(g.configFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %v", err)
	}

	// Count total GIFs for logging
	totalGifs := 0
	for _, urls := range g.codeMap {
		totalGifs += len(urls)
	}
	log.Printf("Saved %d code mappings with %d total GIFs to %s", len(g.codeMap), totalGifs, g.configFile)

	return nil
}

// AddGif adds a GIF URL to a code's list, creates the code if it doesn't exist
func (g *GifList) AddGif(code string, gifURL string) error {
	if len(code) != 2 {
		log.Printf("Rejected invalid code length: %s (%d chars)", code, len(code))
		return fmt.Errorf("code must be exactly 2 characters")
	}

	g.mutex.Lock()

	// Check if this URL is already in the list for this code
	urls, found := g.codeMap[code]
	if found {
		for _, existingURL := range urls {
			if existingURL == gifURL {
				g.mutex.Unlock()
				log.Printf("URL already exists for code %s: %s", code, gifURL)
				return fmt.Errorf("URL already exists for this code")
			}
		}
		log.Printf("Adding new URL for existing code %s: %s", code, gifURL)
		g.codeMap[code] = append(g.codeMap[code], gifURL)
	} else {
		log.Printf("Creating new code %s with URL: %s", code, gifURL)
		g.codeMap[code] = []string{gifURL}
	}

	// Release the lock before file I/O
	g.mutex.Unlock()

	// Persist the change
	if err := g.SaveToFile(); err != nil {
		log.Printf("Warning: Failed to persist code change: %v", err)
	}

	return nil
}

// GetGif returns a randomly selected GIF URL for the given code
func (g *GifList) GetGif(code string) (string, bool) {
	g.mutex.RLock()
	defer g.mutex.RUnlock()

	urls, found := g.codeMap[code]
	if !found || len(urls) == 0 {
		return "", false
	}

	// If there's only one URL, return it
	if len(urls) == 1 {
		// Simplified logging
		return urls[0], true
	}

	// Otherwise, randomly select one
	selectedURL := urls[rand.Intn(len(urls))]
	// Simplified logging for random selection

	return selectedURL, true
}

// GetAllGifsForCode returns all GIF URLs for a given code
func (g *GifList) GetAllGifsForCode(code string) ([]string, bool) {
	g.mutex.RLock()
	defer g.mutex.RUnlock()

	urls, found := g.codeMap[code]
	return urls, found && len(urls) > 0
}

// RemoveGif removes a specific GIF URL from a code
func (g *GifList) RemoveGif(code string, gifURL string) bool {
	g.mutex.Lock()

	urls, exists := g.codeMap[code]
	if !exists {
		g.mutex.Unlock()
		log.Printf("Attempted to remove from non-existent code: %s", code)
		return false
	}

	// If no specific URL provided, remove all URLs for the code
	if gifURL == "" {
		delete(g.codeMap, code)
		log.Printf("Removed entire code: %s with %d GIFs", code, len(urls))

		g.mutex.Unlock()

		// Persist the change
		if err := g.SaveToFile(); err != nil {
			log.Printf("Warning: Failed to persist removal: %v", err)
		}

		return true
	}

	// Find and remove the specific URL
	found := false
	newURLs := make([]string, 0, len(urls))
	for _, url := range urls {
		if url == gifURL {
			found = true
		} else {
			newURLs = append(newURLs, url)
		}
	}

	if !found {
		g.mutex.Unlock()
		log.Printf("URL not found for code %s: %s", code, gifURL)
		return false
	}

	// If removing the last URL for this code, delete the code entirely
	if len(newURLs) == 0 {
		delete(g.codeMap, code)
		log.Printf("Removed last URL for code %s, deleting code", code)
	} else {
		g.codeMap[code] = newURLs
		log.Printf("Removed URL for code %s, %d URLs remaining", code, len(newURLs))
	}

	// Release lock before file I/O
	g.mutex.Unlock()

	// Persist the change
	if err := g.SaveToFile(); err != nil {
		log.Printf("Warning: Failed to persist removal: %v", err)
	}

	return true
}

// RemoveCode removes all GIFs associated with a code
func (g *GifList) RemoveCode(code string) bool {
	return g.RemoveGif(code, "")
}

// ListAllCodes returns all available codes
func (g *GifList) ListAllCodes() []string {
	g.mutex.RLock()
	defer g.mutex.RUnlock()

	codes := make([]string, 0, len(g.codeMap))
	for code := range g.codeMap {
		codes = append(codes, code)
	}
	// Removed verbose logging here
	return codes
}

// GetCodeDetails returns the number of GIFs for each code
type CodeDetails struct {
	Code     string
	GifCount int
}

// ListCodesWithCounts returns details about all codes and their GIF counts
func (g *GifList) ListCodesWithCounts() []CodeDetails {
	g.mutex.RLock()
	defer g.mutex.RUnlock()

	details := make([]CodeDetails, 0, len(g.codeMap))
	for code, urls := range g.codeMap {
		details = append(details, CodeDetails{
			Code:     code,
			GifCount: len(urls),
		})
	}

	return details
}
