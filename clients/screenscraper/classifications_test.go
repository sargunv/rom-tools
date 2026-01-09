package screenscraper

import (
	"testing"
)

func TestGetClassificationsList(t *testing.T) {
	client := testClient(t)

	resp, err := client.GetClassificationsList()
	if err != nil {
		t.Fatalf("GetClassificationsList() error = %v", err)
	}

	if resp.Header.Success != "true" {
		t.Errorf("Expected success=true, got %s", resp.Header.Success)
	}

	if len(resp.Response.Classifications) == 0 {
		t.Error("Expected at least one classification")
	}

	// Check first classification
	for _, classification := range resp.Response.Classifications {
		if classification.ID == 0 {
			t.Error("Expected classification ID to be set")
		}
		if classification.ShortName == "" {
			t.Error("Expected classification ShortName to be set")
		}
		break
	}
}
