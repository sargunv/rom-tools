package screenscraper

import (
	"testing"
)

func TestGetGenresList(t *testing.T) {
	client := testClient(t)

	resp, err := client.GetGenresList()
	if err != nil {
		t.Fatalf("GetGenresList() error = %v", err)
	}

	if resp.Header.Success != "true" {
		t.Errorf("Expected success=true, got %s", resp.Header.Success)
	}

	if len(resp.Response.Genres) == 0 {
		t.Error("Expected at least one genre")
	}

	// Check first genre
	for _, genre := range resp.Response.Genres {
		if genre.ID == 0 {
			t.Error("Expected genre ID to be set")
		}
		break
	}
}
