package screenscraper

import (
	"testing"
)

func TestGetInfraInfo(t *testing.T) {
	client := testClient(t)

	resp, err := client.GetInfraInfo()
	if err != nil {
		t.Fatalf("GetInfraInfo() error = %v", err)
	}

	if resp.Header.Success != "true" {
		t.Errorf("Expected success=true, got %s", resp.Header.Success)
	}

	if resp.Response.Servers.APIAccess == "" {
		t.Error("Expected Servers.APIAccess to be set")
	}
}
