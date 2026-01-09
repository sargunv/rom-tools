package screenscraper

import (
	"testing"
)

func TestGetPlayerCountsList(t *testing.T) {
	client := testClient(t)

	resp, err := client.GetPlayerCountsList()
	if err != nil {
		t.Fatalf("GetPlayerCountsList() error = %v", err)
	}

	if resp.Header.Success != "true" {
		t.Errorf("Expected success=true, got %s", resp.Header.Success)
	}

	if len(resp.Response.PlayerCounts) == 0 {
		t.Error("Expected at least one player count")
	}

	// Check first player count
	for _, count := range resp.Response.PlayerCounts {
		if count.ID == 0 {
			t.Error("Expected player count ID to be set")
		}
		if count.Name == "" {
			t.Error("Expected player count Name to be set")
		}
		break
	}
}
