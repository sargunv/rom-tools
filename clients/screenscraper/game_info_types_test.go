package screenscraper

import (
	"testing"
)

func TestGetGameInfoTypesList(t *testing.T) {
	client := testClient(t)

	resp, err := client.GetGameInfoTypesList()
	if err != nil {
		t.Fatalf("GetGameInfoTypesList() error = %v", err)
	}

	if resp.Header.Success != "true" {
		t.Errorf("Expected success=true, got %s", resp.Header.Success)
	}

	if len(resp.Response.InfoTypes) == 0 {
		t.Error("Expected at least one game info type")
	}

	// Check first info type
	for _, infoType := range resp.Response.InfoTypes {
		if infoType.ID == 0 {
			t.Error("Expected game info type ID to be set")
		}
		if infoType.ShortName == "" {
			t.Error("Expected game info type ShortName to be set")
		}
		break
	}
}
