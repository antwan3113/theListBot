package combo

import (
	"encoding/json"
	"log"
	"os"
	"sync"
	"time"
)

type ComboTracker struct {
	dailyCounts     map[string]int
	lifetimeCounts  map[string]int
	lastUsedTime    time.Time
	lastUsedCode    string
	globalCombo     int
	mu              sync.Mutex
	consecutiveTime time.Duration
	filePath        string
}

// ComboEvent is a struct to hold the combo level, the message, and the GIF URL
type ComboEvent struct {
	Level   int
	Message string
	GifURL  string
}

func NewComboTracker(consecutiveTime time.Duration, filePath string) *ComboTracker {
	c := &ComboTracker{
		dailyCounts:     make(map[string]int),
		lifetimeCounts:  make(map[string]int),
		globalCombo:     0,
		consecutiveTime: consecutiveTime,
		filePath:        filePath,
	}
	c.loadLifetimeCounts()
	return c
}

func (c *ComboTracker) RecordCode(userID string, code string) (int, int, *ComboEvent) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Update daily counts
	c.dailyCounts[code]++
	c.lifetimeCounts[code]++

	// Check for consecutive usage
	if !c.lastUsedTime.IsZero() && time.Since(c.lastUsedTime) <= c.consecutiveTime {
		if c.lastUsedCode != code {
			c.globalCombo = 1
		} else {
			c.globalCombo++
		}
	} else {
		c.globalCombo = 1
	}

	// Update last used time and code
	c.lastUsedTime = time.Now()
	c.lastUsedCode = code

	// Determine combo event
	var comboEvent *ComboEvent
	switch c.globalCombo {
	case 2:
		comboEvent = &ComboEvent{Level: 2, Message: "He's heating up....", GifURL: "https://media.tenor.com/HZ7yDEjwlsgAAAAM/hes-heating-up.gif"}
	case 3:
		comboEvent = &ComboEvent{Level: 3, Message: "Hes on fire!", GifURL: "https://i.imgur.com/FQKnDp9.gif"}
	case 4:
		comboEvent = &ComboEvent{Level: 4, Message: "BOOMSHAKALAKA", GifURL: "https://media.tenor.com/J_mncMNX5A8AAAAM/nbajam-boomshakalaka.gif"}
	case 5:
		comboEvent = &ComboEvent{Level: 5, Message: "C C C COMBO BREAKER", GifURL: "https://media.tenor.com/homlsrzxig8AAAAM/kung-fu-nuts.gif"}
		c.globalCombo = 0 // Reset combo after breaker
	}

	return c.dailyCounts[code], c.globalCombo, comboEvent
}

func (c *ComboTracker) GetDailyCounts() map[string]int {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Create a copy to avoid race conditions
	counts := make(map[string]int)
	for code, count := range c.dailyCounts {
		counts[code] = count
	}
	return counts
}

// GetLifetimeCounts returns a copy of the lifetime counts map.
func (c *ComboTracker) GetLifetimeCounts() map[string]int {
	c.mu.Lock()
	defer c.mu.Unlock()

	counts := make(map[string]int)
	for code, count := range c.lifetimeCounts {
		counts[code] = count
	}
	return counts
}

func (c *ComboTracker) ResetDailyCounts() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.dailyCounts = make(map[string]int)
}

// loadLifetimeCounts loads the lifetime counts from the JSON file.
func (c *ComboTracker) loadLifetimeCounts() {
	c.mu.Lock()
	defer c.mu.Unlock()

	file, err := os.Open(c.filePath)
	if err != nil {
		log.Printf("Error opening lifetime counts file: %v", err)
		return
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&c.lifetimeCounts)
	if err != nil {
		log.Printf("Error decoding lifetime counts: %v", err)
		return
	}

	log.Println("Successfully loaded lifetime counts from file")
}

// saveLifetimeCounts saves the lifetime counts to the JSON file.
func (c *ComboTracker) saveLifetimeCounts() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	file, err := os.Create(c.filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	err = encoder.Encode(c.lifetimeCounts)
	if err != nil {
		return err
	}

	log.Println("Successfully saved lifetime counts to file")
	return nil
}

// Stop saves the lifetime counts when the bot shuts down.
func (c *ComboTracker) Stop() {
	if err := c.saveLifetimeCounts(); err != nil {
		log.Printf("Error saving lifetime counts on shutdown: %v", err)
	}
}
