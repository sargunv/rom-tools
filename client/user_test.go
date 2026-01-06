package screenscraper

import (
	"testing"
)

func TestGetUserInfo(t *testing.T) {
	client := testClient(t)

	resp, err := client.GetUserInfo()
	if err != nil {
		t.Fatalf("GetUserInfo() error = %v", err)
	}

	if resp.Header.Success != "true" {
		t.Errorf("Expected success=true, got %s", resp.Header.Success)
	}

	if resp.Response.SSUser.ID == "" {
		t.Error("Expected SSUser.ID to be set")
	}
}
