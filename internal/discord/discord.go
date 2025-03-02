package discord

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
)

// manage out discord connection and oauth token infomation
type Discord struct {
	ClientID     string
	ClientSecret string
	RedirectURI  string
	Token        string
}

// create a new discord connection
func NewDiscord() *Discord {
	return &Discord{
		ClientID:     os.Getenv("DISCORD_CLIENT_ID"),
		ClientSecret: os.Getenv("DISCORD_CLIENT_SECRET"),
		RedirectURI:  os.Getenv("DISCORD_REDIRECT_URI"),
	}
}

// GenerateOAuthURL generates the URL for OAuth2 authorization
func (d *Discord) GenerateOAuthURL() string {
	return fmt.Sprintf("https://discord.com/api/oauth2/authorize?client_id=%s&redirect_uri=%s&response_type=code&scope=identify",
		d.ClientID, url.QueryEscape(d.RedirectURI))
}

// ExchangeCodeForToken exchanges the authorization code for an access token
func (d *Discord) ExchangeCodeForToken(code string) error {
	data := url.Values{}
	data.Set("client_id", d.ClientID)
	data.Set("client_secret", d.ClientSecret)
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("redirect_uri", d.RedirectURI)

	req, err := http.NewRequest("POST", "https://discord.com/api/oauth2/token", strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return err
	}

	if token, ok := result["access_token"].(string); ok {
		d.Token = token
		return nil
	}

	return fmt.Errorf("failed to get access token")
}

// token infomation stored in .env file
func (d *Discord) GetToken() string {
	return d.Token
}

// handle input from discord
func (d *Discord) HandleInput(input string) (string, bool) {
	// Check if the input matches the expected pattern for a 2-character code
	input = strings.TrimSpace(input)
	if len(input) == 2 {
		// In a real implementation, look up the code in a database or mapping
		// For now, return a dummy response
		return fmt.Sprintf("Responding to code: %s", input), true
	}
	return "", false
}
