package combo

import (
	"sync"
	"time"
)

type ComboTracker struct {
	dailyCounts     map[string]int
	lastUsed        map[string]map[string]time.Time // userID -> (code -> timestamp)
	userCombos      map[string]int                  // userID -> combo count
	mu              sync.Mutex
	consecutiveTime time.Duration
}

// ComboEvent is a struct to hold the combo level, the message, and the GIF URL
type ComboEvent struct {
	Level   int
	Message string
	GifURL  string
}

func NewComboTracker(consecutiveTime time.Duration) *ComboTracker {
	return &ComboTracker{
		dailyCounts:     make(map[string]int),
		lastUsed:        make(map[string]map[string]time.Time),
		userCombos:      make(map[string]int),
		consecutiveTime: consecutiveTime,
	}
}

func (c *ComboTracker) RecordCode(userID string, code string) (int, int, *ComboEvent) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Update daily counts
	c.dailyCounts[code]++

	// Initialize user's last used map if it doesn't exist
	if _, ok := c.lastUsed[userID]; !ok {
		c.lastUsed[userID] = make(map[string]time.Time)
	}

	// Check for consecutive usage
	lastUsedTime, ok := c.lastUsed[userID][code]
	if ok && time.Since(lastUsedTime) <= c.consecutiveTime {
		c.userCombos[userID]++
	} else {
		c.userCombos[userID] = 1
	}

	// Update last used time
	c.lastUsed[userID][code] = time.Now()

	// Determine combo event
	var comboEvent *ComboEvent
	switch c.userCombos[userID] {
	case 2:
		comboEvent = &ComboEvent{Level: 2, Message: "He's heating up....", GifURL: "https://media.tenor.com/HZ7yDEjwlsgAAAAM/hes-heating-up.gif"}
	case 3:
		comboEvent = &ComboEvent{Level: 3, Message: "Hes on fire!", GifURL: "https://i.imgur.com/FQKnDp9.gif"}
	case 4:
		comboEvent = &ComboEvent{Level: 4, Message: "BOOMSHAKALAKA", GifURL: "https://media.tenor.com/J_mncMNX5A8AAAAM/nbajam-boomshakalaka.gif"}
	case 5:
		comboEvent = &ComboEvent{Level: 5, Message: "C C C COMBO BREAKER", GifURL: "https://media.tenor.com/homlsrzxig8AAAAM/kung-fu-nuts.gif"}
		c.userCombos[userID] = 0 // Reset combo after breaker
	}

	return c.dailyCounts[code], c.userCombos[userID], comboEvent
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

func (c *ComboTracker) ResetDailyCounts() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.dailyCounts = make(map[string]int)
}
