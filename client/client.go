package screenscraper

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const BaseURL = "https://api.screenscraper.fr/api2"

// Client is a Screenscraper API client
type Client struct {
	HTTPClient  *http.Client
	DevID       string
	DevPassword string
	SoftName    string
	SSID        string
	SSPassword  string
}

// NewClient creates a new Screenscraper API client
func NewClient(devID, devPassword, softName, ssid, sspassword string) *Client {
	return &Client{
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		DevID:       devID,
		DevPassword: devPassword,
		SoftName:    softName,
		SSID:        ssid,
		SSPassword:  sspassword,
	}
}

// buildURL constructs the API URL with common parameters
func (c *Client) buildURL(endpoint string, params map[string]string) string {
	u := fmt.Sprintf("%s/%s", BaseURL, endpoint)
	v := url.Values{}

	// Add required credentials
	v.Set("devid", c.DevID)
	v.Set("devpassword", c.DevPassword)
	v.Set("softname", c.SoftName)
	v.Set("output", "json")

	// Add optional user credentials if provided
	if c.SSID != "" {
		v.Set("ssid", c.SSID)
	}
	if c.SSPassword != "" {
		v.Set("sspassword", c.SSPassword)
	}

	// Add additional parameters
	for key, val := range params {
		if val != "" {
			v.Set(key, val)
		}
	}

	return u + "?" + v.Encode()
}

// validateResponse checks the Header fields in a JSON response and returns an error if the request failed
func validateResponse(header Header) error {
	// Check if success field indicates failure
	if header.Success != "" && header.Success != "true" {
		// If we have an error message, use it; otherwise use a generic message
		if header.Error != "" {
			// Try to map the error message to a status code
			// Most API errors come through HTTP status codes, but some may come through Header.Error
			// Default to 400 for errors in Header.Error without a status code
			return newAPIError(400, header.Error)
		}
		return fmt.Errorf("API request failed: success=%s", header.Success)
	}

	// Check if error field is set (even if success is true, error might be set)
	if header.Error != "" {
		return newAPIError(400, header.Error)
	}

	return nil
}

// get performs a GET request and returns the raw response body
func (c *Client) get(endpoint string, params map[string]string) ([]byte, error) {
	apiURL := c.buildURL(endpoint, params)

	resp, err := c.HTTPClient.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check for HTTP errors
	if resp.StatusCode != http.StatusOK {
		// Try to extract error message from response body
		message := string(body)
		// Try to parse as JSON to extract error message if possible
		var jsonResp struct {
			Header struct {
				Error string `json:"error"`
			} `json:"header"`
		}
		if err := json.Unmarshal(body, &jsonResp); err == nil && jsonResp.Header.Error != "" {
			message = jsonResp.Header.Error
		}
		return nil, newAPIError(resp.StatusCode, message)
	}

	return body, nil
}
