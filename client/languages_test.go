package screenscraper

import (
	"testing"
)

func TestGetLanguagesList(t *testing.T) {
	client := testClient(t)

	resp, err := client.GetLanguagesList()
	if err != nil {
		t.Fatalf("GetLanguagesList() error = %v", err)
	}

	if resp.Header.Success != "true" {
		t.Errorf("Expected success=true, got %s", resp.Header.Success)
	}

	if len(resp.Response.Languages) == 0 {
		t.Error("Expected at least one language")
	}

	// Check first language
	for _, lang := range resp.Response.Languages {
		if lang.ID == 0 {
			t.Error("Expected language ID to be set")
		}
		break
	}
}
