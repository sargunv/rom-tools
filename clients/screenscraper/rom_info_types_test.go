package screenscraper

import (
	"testing"
)

func TestGetROMInfoTypesList(t *testing.T) {
	client := testClient(t)

	resp, err := client.GetROMInfoTypesList()
	if err != nil {
		t.Fatalf("GetROMInfoTypesList() error = %v", err)
	}

	if resp.Header.Success != "true" {
		t.Errorf("Expected success=true, got %s", resp.Header.Success)
	}

	if len(resp.Response.InfoTypes) == 0 {
		t.Error("Expected at least one ROM info type")
	}

	// Check first info type
	for _, infoType := range resp.Response.InfoTypes {
		if infoType.ID == 0 {
			t.Error("Expected ROM info type ID to be set")
		}
		if infoType.ShortName == "" {
			t.Error("Expected ROM info type ShortName to be set")
		}
		break
	}
}
