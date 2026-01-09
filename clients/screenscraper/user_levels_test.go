package screenscraper

import (
	"testing"
)

func TestGetUserLevelsList(t *testing.T) {
	client := testClient(t)

	resp, err := client.GetUserLevelsList()
	if err != nil {
		t.Fatalf("GetUserLevelsList() error = %v", err)
	}

	if resp.Header.Success != "true" {
		t.Errorf("Expected success=true, got %s", resp.Header.Success)
	}

	if len(resp.Response.UserLevels) == 0 {
		t.Error("Expected at least one user level")
	}

	// Check first user level
	for _, level := range resp.Response.UserLevels {
		if level.ID == 0 {
			t.Error("Expected user level ID to be set")
		}
		if level.NameFR == "" {
			t.Error("Expected user level NameFR to be set")
		}
		break
	}
}
