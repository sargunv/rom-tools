package screenscraper

import (
	"fmt"
	"io"
	"strconv"
	"strings"
)

// SubmitRatingParams contains parameters for submitting a game rating
type SubmitRatingParams struct {
	// GameID is the numeric identifier of the game on ScreenScraper
	GameID string
	// Rating is the score out of 20 (must be between 1 and 20)
	Rating int
}

// SubmitRatingResponse contains the response from submitting a rating
type SubmitRatingResponse struct {
	// Message is the textual response from the API
	Message string
}

// SubmitRating submits a rating for a game (botNote.php endpoint)
// This requires user credentials (SSID and SSPassword) to be set on the client.
// The rating must be between 1 and 20.
func (c *Client) SubmitRating(params SubmitRatingParams) (*SubmitRatingResponse, error) {
	// Validate that user credentials are provided
	if c.SSID == "" || c.SSPassword == "" {
		return nil, fmt.Errorf("user credentials (SSID and SSPassword) are required to submit ratings")
	}

	// Validate rating range
	if params.Rating < 1 || params.Rating > 20 {
		return nil, fmt.Errorf("rating must be between 1 and 20, got %d", params.Rating)
	}

	// Validate game ID
	if params.GameID == "" {
		return nil, fmt.Errorf("game ID is required")
	}

	p := map[string]string{
		"gameid": params.GameID,
		"note":   strconv.Itoa(params.Rating),
	}

	apiURL := c.buildURL("botNote.php", p)

	resp, err := c.HTTPClient.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	message := strings.TrimSpace(string(body))

	// Check for HTTP errors
	if resp.StatusCode != 200 {
		return nil, newAPIError(resp.StatusCode, message)
	}

	return &SubmitRatingResponse{
		Message: message,
	}, nil
}
