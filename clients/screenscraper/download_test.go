package screenscraper

import (
	"testing"
)

func TestDownloadGameMedia(t *testing.T) {
	client := testClient(t)

	params := DownloadMediaParams{
		SystemID: "1", // Mega Drive
		GameID:   "1", // Battletoads
		Media:    "sstitle(wor)",
	}

	data, err := client.DownloadGameMedia(params)
	if err != nil {
		t.Fatalf("DownloadGameMedia() error = %v", err)
	}

	if len(data) == 0 {
		t.Error("Expected non-empty media data")
	}

	// Check if it's a PNG (common format for game media)
	if len(data) < 8 {
		t.Error("Expected media data to be at least 8 bytes")
	}
}

func TestDownloadSystemMedia(t *testing.T) {
	client := testClient(t)

	params := DownloadMediaParams{
		SystemID: "1", // Mega Drive
		Media:    "logo-monochrome(wor)",
	}

	data, err := client.DownloadSystemMedia(params)
	if err != nil {
		t.Fatalf("DownloadSystemMedia() error = %v", err)
	}

	if len(data) == 0 {
		t.Error("Expected non-empty media data")
	}

	// Check if it's a PNG (common format for system media)
	if len(data) < 8 {
		t.Error("Expected media data to be at least 8 bytes")
	}
}

func TestDownloadGroupMedia(t *testing.T) {
	client := testClient(t)

	params := DownloadGroupMediaParams{
		GroupID: "10", // Genre ID
		Media:   "logo-monochrome",
	}

	data, err := client.DownloadGroupMedia(params)
	if err != nil {
		t.Fatalf("DownloadGroupMedia() error = %v", err)
	}

	if len(data) == 0 {
		t.Error("Expected non-empty media data")
	}

	if len(data) < 8 {
		t.Error("Expected media data to be at least 8 bytes")
	}
}

func TestDownloadCompanyMedia(t *testing.T) {
	client := testClient(t)

	params := DownloadCompanyMediaParams{
		CompanyID: "2", // Company ID
		Media:     "logo-monochrome",
	}

	data, err := client.DownloadCompanyMedia(params)
	if err != nil {
		t.Fatalf("DownloadCompanyMedia() error = %v", err)
	}

	if len(data) == 0 {
		t.Error("Expected non-empty media data")
	}

	if len(data) < 8 {
		t.Error("Expected media data to be at least 8 bytes")
	}
}
