package screenscraper

import (
	"testing"
)

func TestGetGameMediaList(t *testing.T) {
	client := testClient(t)

	resp, err := client.GetGameMediaList()
	if err != nil {
		t.Fatalf("GetGameMediaList() error = %v", err)
	}

	if resp.Header.Success != "true" {
		t.Errorf("Expected success=true, got %s", resp.Header.Success)
	}

	if len(resp.Response.Medias) == 0 {
		t.Error("Expected at least one media type")
	}

	// Check first media type
	for _, mediaType := range resp.Response.Medias {
		if mediaType.ID == 0 {
			t.Error("Expected media type ID to be set")
		}
		break
	}
}

func TestGetSystemMediaList(t *testing.T) {
	client := testClient(t)

	resp, err := client.GetSystemMediaList()
	if err != nil {
		t.Fatalf("GetSystemMediaList() error = %v", err)
	}

	if resp.Header.Success != "true" {
		t.Errorf("Expected success=true, got %s", resp.Header.Success)
	}

	if len(resp.Response.Medias) == 0 {
		t.Error("Expected at least one media type")
	}

	// Check first media type
	for _, mediaType := range resp.Response.Medias {
		if mediaType.ID == 0 {
			t.Error("Expected media type ID to be set")
		}
		break
	}
}
