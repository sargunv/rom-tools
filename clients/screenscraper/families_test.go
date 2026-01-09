package screenscraper

import (
	"testing"
)

func TestGetFamiliesList(t *testing.T) {
	client := testClient(t)

	resp, err := client.GetFamiliesList()
	if err != nil {
		t.Fatalf("GetFamiliesList() error = %v", err)
	}

	if resp.Header.Success != "true" {
		t.Errorf("Expected success=true, got %s", resp.Header.Success)
	}

	if len(resp.Response.Families) == 0 {
		t.Error("Expected at least one family")
	}

	// Check first family
	for _, family := range resp.Response.Families {
		if family.ID == 0 {
			t.Error("Expected family ID to be set")
		}
		if family.Name == "" {
			t.Error("Expected family Name to be set")
		}
		break
	}
}
