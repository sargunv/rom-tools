package screenscraper

import (
	"testing"
)

func TestGetGameInfo(t *testing.T) {
	client := testClient(t)

	params := GameInfoParams{
		GameID: "1", // Battletoads
	}

	resp, err := client.GetGameInfo(params)
	if err != nil {
		t.Fatalf("GetGameInfo() error = %v", err)
	}

	if resp.Header.Success != "true" {
		t.Errorf("Expected success=true, got %s", resp.Header.Success)
	}

	if resp.Response.Game.ID == "" {
		t.Error("Expected game ID to be set")
	}

	if resp.Response.Game.System.ID == "" {
		t.Error("Expected system ID to be set")
	}
}
