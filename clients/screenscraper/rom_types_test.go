package screenscraper

import (
	"testing"
)

func TestGetROMTypesList(t *testing.T) {
	client := testClient(t)

	resp, err := client.GetROMTypesList()
	if err != nil {
		t.Fatalf("GetROMTypesList() error = %v", err)
	}

	if resp.Header.Success != "true" {
		t.Errorf("Expected success=true, got %s", resp.Header.Success)
	}

	if len(resp.Response.ROMTypes) == 0 {
		t.Error("Expected at least one ROM type")
	}

	// Check first ROM type
	if len(resp.Response.ROMTypes) > 0 {
		rt := resp.Response.ROMTypes[0]
		if rt == "" {
			t.Error("Expected ROM type to be set")
		}
	}
}
