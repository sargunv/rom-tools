package screenscraper

import (
	"testing"
)

func TestGetSupportTypesList(t *testing.T) {
	client := testClient(t)

	resp, err := client.GetSupportTypesList()
	if err != nil {
		t.Fatalf("GetSupportTypesList() error = %v", err)
	}

	if resp.Header.Success != "true" {
		t.Errorf("Expected success=true, got %s", resp.Header.Success)
	}

	if len(resp.Response.SupportTypes) == 0 {
		t.Error("Expected at least one support type")
	}

	// Check first support type
	if len(resp.Response.SupportTypes) > 0 {
		st := resp.Response.SupportTypes[0]
		if st == "" {
			t.Error("Expected support type to be set")
		}
	}
}
