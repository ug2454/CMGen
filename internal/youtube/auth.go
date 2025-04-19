package youtube

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"time"

	"golang.org/x/oauth2"
)

// getTokenFromWeb uses Config to request a Token.
// It opens a browser for the user to authenticate and then
// returns the retrieved token.
func getTokenFromWeb(ctx context.Context, config *oauth2.Config) (*oauth2.Token, error) {
	// Generate the auth URL with offline access for refresh token
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)

	// Print instructions with the URL
	fmt.Println("\n=== YouTube Authentication Required ===")
	fmt.Println("To allow this application to update your YouTube videos, please:")
	fmt.Println("1. Open this URL in your browser:")
	fmt.Printf("   %s\n\n", authURL)
	fmt.Println("2. Log in with your Google account that has access to the YouTube channel")
	fmt.Println("3. Allow the requested permissions")
	fmt.Println("4. Copy the authorization code and paste it below")

	// Try to open the browser automatically
	browserOpened := openBrowser(authURL)
	if browserOpened {
		fmt.Println("\n✓ A browser window should have opened automatically.")
	} else {
		fmt.Println("\n⚠ Could not open browser automatically. Please copy and open the URL manually.")
	}

	// Add a delay to ensure messages are displayed before the prompt
	time.Sleep(500 * time.Millisecond)

	// Prompt for the code
	fmt.Print("\nEnter the authorization code: ")
	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		return nil, fmt.Errorf("unable to read authorization code: %v", err)
	}

	// Exchange the auth code for a token
	token, err := config.Exchange(ctx, authCode)
	if err != nil {
		// Provide detailed error messages for common issues
		if os.IsTimeout(err) {
			return nil, fmt.Errorf("authentication timed out: %v\nPlease try again and complete the authentication more quickly", err)
		}

		errMsg := err.Error()
		if contains(errMsg, "invalid_grant") {
			return nil, fmt.Errorf("invalid authorization code: %v\nPlease ensure you copied the entire code correctly", err)
		} else if contains(errMsg, "redirect_uri_mismatch") {
			return nil, fmt.Errorf("redirect URI mismatch: %v\nPlease ensure your OAuth credentials are correctly configured. Desktop application credentials are recommended", err)
		}

		return nil, fmt.Errorf("unable to exchange authorization code: %v", err)
	}

	fmt.Println("Authentication successful! You can now use YouTube integration.")
	return token, nil
}

// openBrowser tries to open the URL in the user's browser.
// Returns true if the browser was opened, false otherwise.
func openBrowser(url string) bool {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		return false
	}

	// Start the command without waiting for it to complete
	err := cmd.Start()
	if err != nil {
		return false
	}

	// Don't wait for the browser process
	// This avoids blocking the application

	return true
}

// contains checks if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[0:len(substr)] == substr
}
