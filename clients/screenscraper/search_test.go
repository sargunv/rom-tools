package screenscraper

import (
	"testing"
)

func TestSearchGame(t *testing.T) {
	client := testClient(t)

	params := SearchGameParams{
		Query:    "Chrono Trigger",
		SystemID: "4", // SNES
	}

	resp, err := client.SearchGame(params)
	if err != nil {
		t.Fatalf("SearchGame() error = %v", err)
	}

	if resp.Header.Success != "true" {
		t.Errorf("Expected success=true, got %s", resp.Header.Success)
	}

	if len(resp.Response.Games) == 0 {
		t.Error("Expected at least one game")
	}

	if resp.Response.Games[0].ID == "" {
		t.Error("Expected game ID to be set")
	}
}
