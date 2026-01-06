package screenscraper

import (
	"testing"
)

func TestGetRegionsList(t *testing.T) {
	client := testClient(t)

	resp, err := client.GetRegionsList()
	if err != nil {
		t.Fatalf("GetRegionsList() error = %v", err)
	}

	if resp.Header.Success != "true" {
		t.Errorf("Expected success=true, got %s", resp.Header.Success)
	}

	if len(resp.Response.Regions) == 0 {
		t.Error("Expected at least one region")
	}

	// Check first region
	for _, region := range resp.Response.Regions {
		if region.ID == 0 {
			t.Error("Expected region ID to be set")
		}
		break
	}
}
