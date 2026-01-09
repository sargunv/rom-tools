package screenscraper

import (
	"os"
	"testing"
)

// testClient creates a new client for testing using environment variables.
// It fails the test if required environment variables are not set.
func testClient(t *testing.T) *Client {
	t.Helper()

	devID := os.Getenv("SCREENSCRAPER_DEV_USER")
	devPassword := os.Getenv("SCREENSCRAPER_DEV_PASSWORD")
	ssID := os.Getenv("SCREENSCRAPER_ID")
	ssPassword := os.Getenv("SCREENSCRAPER_PASSWORD")

	if devID == "" || devPassword == "" {
		t.Fatalf("Required environment variables not set: SCREENSCRAPER_DEV_USER and SCREENSCRAPER_DEV_PASSWORD")
	}

	return NewClient(devID, devPassword, "github.com/sargunv/rom-tools/tests", ssID, ssPassword)
}
