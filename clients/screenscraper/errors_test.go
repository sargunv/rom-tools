package screenscraper

import (
	"errors"
	"testing"
)

func TestErrorHandling_NotFound(t *testing.T) {
	client := testClient(t)

	// Test 404 Not Found - use a very large game ID that likely doesn't exist
	params := GameInfoParams{
		GameID: "999999999", // Non-existent game ID
	}

	resp, err := client.GetGameInfo(params)
	if err == nil {
		t.Fatal("Expected error for non-existent game, got nil")
	}

	// Verify it's a NotFound error
	if !IsNotFound(err) {
		t.Errorf("Expected NotFound error, got: %v", err)
	}

	// Verify error type
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("Expected APIError, got: %T", err)
	}

	if apiErr.StatusCode != 404 {
		t.Errorf("Expected status code 404, got %d", apiErr.StatusCode)
	}

	if apiErr.Type != ErrorTypeNotFound {
		t.Errorf("Expected ErrorTypeNotFound, got %v", apiErr.Type)
	}

	// Response should be nil on error
	if resp != nil {
		t.Error("Expected nil response on error")
	}
}

func TestErrorHandling_BadRequest_InvalidHash(t *testing.T) {
	client := testClient(t)

	// Test 400 Bad Request - invalid hash format (too short)
	params := GameInfoParams{
		CRC:      "INVALID", // Invalid CRC format (should be 8 hex chars)
		MD5:      "invalid", // Invalid MD5 format (should be 32 hex chars)
		SystemID: "1",
		ROMType:  "rom",
		ROMName:  "test.zip",
		ROMSize:  "1000",
	}

	resp, err := client.GetGameInfo(params)
	if err == nil {
		t.Fatal("Expected error for invalid hash format, got nil")
	}

	// Verify it's a BadRequest error
	if !IsBadRequest(err) {
		t.Errorf("Expected BadRequest error, got: %v", err)
	}

	// Verify error type
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("Expected APIError, got: %T", err)
	}

	if apiErr.StatusCode != 400 {
		t.Errorf("Expected status code 400, got %d", apiErr.StatusCode)
	}

	if apiErr.Type != ErrorTypeBadRequest {
		t.Errorf("Expected ErrorTypeBadRequest, got %v", apiErr.Type)
	}

	// Response should be nil on error
	if resp != nil {
		t.Error("Expected nil response on error")
	}
}

func TestErrorHandling_BadRequest_MissingFields(t *testing.T) {
	client := testClient(t)

	// Test 400 Bad Request - missing required fields (no system ID)
	params := GameInfoParams{
		ROMType: "rom",
		ROMName: "test.zip",
		// Missing SystemID, which is required
	}

	resp, err := client.GetGameInfo(params)
	if err == nil {
		t.Fatal("Expected error for missing required fields, got nil")
	}

	// Verify it's a BadRequest error
	if !IsBadRequest(err) {
		t.Errorf("Expected BadRequest error, got: %v", err)
	}

	// Verify error type
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("Expected APIError, got: %T", err)
	}

	if apiErr.StatusCode != 400 {
		t.Errorf("Expected status code 400, got %d", apiErr.StatusCode)
	}

	// Response should be nil on error
	if resp != nil {
		t.Error("Expected nil response on error")
	}
}
