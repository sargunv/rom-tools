package screenscraper

import (
	"testing"
)

func TestGetSystemsList(t *testing.T) {
	client := testClient(t)

	resp, err := client.GetSystemsList()
	if err != nil {
		t.Fatalf("GetSystemsList() error = %v", err)
	}

	if resp.Header.Success != "true" {
		t.Errorf("Expected success=true, got %s", resp.Header.Success)
	}

	if len(resp.Response.Systems) == 0 {
		t.Error("Expected at least one system")
	}

	if resp.Response.Systems[0].ID == 0 {
		t.Error("Expected system ID to be set")
	}
}
