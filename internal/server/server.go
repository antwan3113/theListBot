package server

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"syscall"
	"theListBot/internal/combo" // Import the combo package
	"theListBot/internal/giflist"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

type Server struct {
	discordSession *discordgo.Session
	gifList        *giflist.GifList
	comboTracker   *combo.ComboTracker // Add the combo tracker
	done           chan os.Signal
}

func NewServer() *Server {
	// Define the file path for lifetime counts
	filePath := "lifetime_counts.json"

	return &Server{
		gifList:      giflist.NewGifList(),
		comboTracker: combo.NewComboTracker(600*time.Second, filePath), // Pass the file path
		done:         make(chan os.Signal, 1),
	}
}

func (s *Server) Start() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	token := os.Getenv("DISCORD_TOKEN")
	if token == "" {
		log.Fatal("No DISCORD_TOKEN provided")
	}

	// Create Discord session with proper intents to read messages
	s.discordSession, err = discordgo.New("Bot " + token)
	if err != nil {
		log.Fatalf("Error creating Discord session: %v", err)
	}

	// Set required intents to receive message events
	s.discordSession.Identify.Intents = discordgo.IntentsGuildMessages | discordgo.IntentsDirectMessages

	// Register message handler
	s.discordSession.AddHandler(s.messageHandler)

	err = s.discordSession.Open()
	if err != nil {
		log.Fatalf("Error opening connection to Discord: %v", err)
	}

	log.Println("Bot is now running and listening for commands. Press CTRL+C to exit.")

	// Start the daily reset ticker
	go s.dailyResetTicker()

	// Wait for a termination signal
	signal.Notify(s.done, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-s.done

	// Clean up before exiting
	s.Stop()
}

func (s *Server) Stop() {
	if s.discordSession != nil {
		log.Println("Closing Discord session...")
		s.discordSession.Close()
	}
	// Stop the combo tracker to save lifetime counts
	s.comboTracker.Stop()
	log.Println("Server shutdown complete")
}

// dailyResetTicker resets the daily counts at midnight
func (s *Server) dailyResetTicker() {
	now := time.Now()
	midnight := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
	duration := midnight.Sub(now)

	// Wait until midnight
	time.Sleep(duration)

	ticker := time.NewTicker(24 * time.Hour)
	for range ticker.C {
		s.comboTracker.ResetDailyCounts()
		log.Println("Daily counts reset at midnight")
	}
}

// messageHandler processes Discord message events
func (s *Server) messageHandler(session *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore messages from the bot itself
	if m.Author.ID == session.State.User.ID {
		return
	}

	// Only log commands and matched codes, not regular messages
	// Regular message logging removed here

	// Handle admin commands for managing GIFs
	if strings.HasPrefix(m.Content, "!list") {
		log.Printf("Command from %s: %s", m.Author.Username, m.Content)
		s.handleListCommand(session, m)
		return
	}

	// Handle combo commands
	if strings.HasPrefix(m.Content, "!counts") {
		log.Printf("Command from %s: %s", m.Author.Username, m.Content)
		s.handleComboCommand(session, m)
		return
	}

	// Check for 2-character codes in the message
	s.processMessageForCodes(session, m)
}

// processMessageForCodes looks for 2-character codes at the start of messages
func (s *Server) processMessageForCodes(session *discordgo.Session, m *discordgo.MessageCreate) {
	// Remove verbose logging for every message check
	// log.Printf("Checking message for codes: %s", m.Content)

	// Updated regex to only match at the beginning of the message
	// ^(?i) - Start of string + case insensitive
	// ([a-zA-Z]{2}) - Two letters as our code
	// (\b|$|[^a-zA-Z]) - Must be followed by word boundary, end of string, or non-letter
	codePattern := regexp.MustCompile(`^(?i)([a-zA-Z0-9\.]+)$`)

	// Debug regex logging removed
	// log.Printf("Using regex pattern: %s", codePattern.String())

	// Find the match at the start of the message
	match := codePattern.FindStringSubmatch(m.Content)

	if len(match) >= 2 {
		code := strings.ToLower(match[1]) // Extract the code and convert to lowercase

		// Only log when we've identified a code
		log.Printf("Code match from %s: %s", m.Author.Username, code)

		// Record the code usage and get the counts
		dailyCount, userCombo, comboEvent := s.comboTracker.RecordCode(m.Author.ID, code)
		log.Printf("Code %s used by %s. Daily count: %d, User combo: %d", code, m.Author.Username, dailyCount, userCombo)

		if gifURL, found := s.gifList.GetGif(code); found {
			log.Printf("Sending GIF for code %s: %s", code, gifURL)

			// Respond with the gif
			_, err := session.ChannelMessageSend(m.ChannelID, gifURL)
			if err != nil {
				log.Printf("Error sending GIF response: %v", err)
			}

			// Check if there is a combo event and send the combo message and GIF
			if comboEvent != nil {
				_, err = session.ChannelMessageSend(m.ChannelID, fmt.Sprintf("%s %s", comboEvent.Message, comboEvent.GifURL))
				if err != nil {
					log.Printf("Error sending combo message: %v", err)
				}
			}
		} else {
			log.Printf("No GIF found for code: %s", code)
		}
	}
	// Removed logging for no code found - too verbose
}

// handleListCommand processes commands for managing the gif list
func (s *Server) handleListCommand(session *discordgo.Session, m *discordgo.MessageCreate) {
	parts := strings.Fields(m.Content)
	if len(parts) < 2 {
		log.Println("Showing list command help")
		// Display help message
		session.ChannelMessageSend(m.ChannelID, "Available commands:\n"+
			"!list show - Display all available codes\n"+
			"!list show [code] - Show all GIFs for a specific code\n"+
			"!list add [code] [url] - Add a new GIF\n"+
			"!list remove [code] - Remove all GIFs for a code\n"+
			"!list remove [code] [url] - Remove a specific GIF\n"+
			"!list help - Show detailed help")
		return
	}

	switch parts[1] {
	case "show":
		if len(parts) >= 3 {
			// Show all GIFs for a specific code
			code := strings.ToLower(parts[2])
			log.Printf("Showing GIFs for code: %s", code)

			urls, found := s.gifList.GetAllGifsForCode(code)
			if !found || len(urls) == 0 {
				session.ChannelMessageSend(m.ChannelID, fmt.Sprintf("No GIFs found for code: %s", code))
				return
			}

			message := fmt.Sprintf("**GIFs for code `%s` (%d):**\n", code, len(urls))
			for i, url := range urls {
				message += fmt.Sprintf("%d. %s\n", i+1, url)
			}

			session.ChannelMessageSend(m.ChannelID, message)
		} else {
			// Show all codes with counts
			log.Println("Processing list show command")
			codeDetails := s.gifList.ListCodesWithCounts()

			if len(codeDetails) == 0 {
				log.Println("No codes available to show")
				session.ChannelMessageSend(m.ChannelID, "No codes available yet.")
				return
			}

			message := fmt.Sprintf("**Available codes (%d):**\n", len(codeDetails))
			for _, detail := range codeDetails {
				message += fmt.Sprintf("`%s` (%d GIFs)\n", detail.Code, detail.GifCount)
			}

			session.ChannelMessageSend(m.ChannelID, message)
		}

	case "add":
		if len(parts) < 4 {
			log.Println("Invalid add command format")
			session.ChannelMessageSend(m.ChannelID, "Usage: !list add [code] [url]")
			return
		}

		code := strings.ToLower(parts[2])
		url := parts[3]
		log.Printf("Adding GIF for code %s: %s", code, url)

		if err := s.gifList.AddGif(code, url); err != nil {
			log.Printf("Error adding GIF: %v", err)
			session.ChannelMessageSend(m.ChannelID, "Error: "+err.Error())
			return
		}

		urls, _ := s.gifList.GetAllGifsForCode(code)
		log.Printf("Successfully added GIF for code: %s (now has %d GIFs)", code, len(urls))
		session.ChannelMessageSend(m.ChannelID,
			fmt.Sprintf("Added GIF for code: `%s` → %s (now has %d GIFs)", code, url, len(urls)))

	case "remove":
		if len(parts) < 3 {
			log.Println("Invalid remove command format")
			session.ChannelMessageSend(m.ChannelID, "Usage: !list remove [code] or !list remove [code] [url]")
			return
		}

		code := strings.ToLower(parts[2])

		// Check if we're removing a specific URL
		if len(parts) >= 4 {
			url := parts[3]
			log.Printf("Attempting to remove specific URL for code %s: %s", code, url)

			if s.gifList.RemoveGif(code, url) {
				log.Printf("Successfully removed URL for code: %s", code)
				session.ChannelMessageSend(m.ChannelID,
					fmt.Sprintf("Removed GIF from code: `%s`", code))
			} else {
				log.Printf("Failed to remove URL for code: %s", code)
				session.ChannelMessageSend(m.ChannelID,
					fmt.Sprintf("URL not found for code: %s", code))
			}
		} else {
			// Remove all GIFs for the code
			log.Printf("Attempting to remove all GIFs for code: %s", code)

			if s.gifList.RemoveCode(code) {
				log.Printf("Successfully removed code: %s", code)
				session.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Removed code: `%s`", code))
			} else {
				log.Printf("Failed to remove non-existent code: %s", code)
				session.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Code not found: %s", code))
			}
		}

	case "help":
		log.Println("Showing detailed help")
		helpMsg := "**The List Bot Commands:**\n" +
			"`!list show` - Display all available codes with GIF counts\n" +
			"`!list show [code]` - Show all GIFs for a specific code\n" +
			"`!list add [code] [url]` - Add a GIF URL to a code\n" +
			"`!list remove [code]` - Remove all GIFs for a code\n" +
			"`!list remove [code] [url]` - Remove a specific GIF URL from a code\n" +
			"`!counts` - Display the daily code counts\n\n" + // Added combo command to help
			"**Usage:**\n" +
			"Type a code at the start of your message to trigger a random GIF\n" +
			"Example: `gg` or `ty everyone`"
		session.ChannelMessageSend(m.ChannelID, helpMsg)

	default:
		log.Printf("Unknown list subcommand: %s", parts[1])
		session.ChannelMessageSend(m.ChannelID, "Unknown command. Use `!list help` for help.")
	}
}

// handleComboCommand displays the daily counts
func (s *Server) handleComboCommand(session *discordgo.Session, m *discordgo.MessageCreate) {
	dailyCounts := s.comboTracker.GetDailyCounts()
	lifetimeCounts := s.comboTracker.GetLifetimeCounts()

	if len(dailyCounts) == 0 && len(lifetimeCounts) == 0 {
		session.ChannelMessageSend(m.ChannelID, "No codes have been used yet.")
		return
	}

	message := "**Daily Code Counts:**\n"
	for code, count := range dailyCounts {
		message += fmt.Sprintf("`%s`: %d\n", code, count)
	}

	message += "\n**Lifetime Code Counts:**\n"
	for code, count := range lifetimeCounts {
		message += fmt.Sprintf("`%s`: %d\n", code, count)
	}

	session.ChannelMessageSend(m.ChannelID, message)
}
