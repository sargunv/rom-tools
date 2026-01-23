package megadrive

import (
	"bytes"
	"testing"
)

func TestParseSegaCD(t *testing.T) {
	// Create a synthetic valid Sega CD header
	data := make([]byte, 0x200)

	// Disc identifier at 0x00
	copy(data[0x00:], "SEGADISCSYSTEM  ")
	// System Type at 0x100
	copy(data[0x100:], "SEGA MEGA DRIVE ")
	// Copyright at 0x110
	copy(data[0x110:], "(C)SEGA 1993.SEP")
	// Domestic Title at 0x120
	copy(data[0x120:], "SONIC THE HEDGEHOG CD")
	// Overseas Title at 0x150
	copy(data[0x150:], "SONIC CD")
	// Serial Number at 0x180
	copy(data[0x180:], "GM MK-4407-00")
	// Device Support at 0x190
	copy(data[0x190:], "J")
	// Region at 0x1F0
	copy(data[0x1F0:], "JUE")

	info, err := ParseSegaCD(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatalf("ParseSegaCD failed: %v", err)
	}

	// Verify all fields
	if info.DiscID != "SEGADISCSYSTEM" {
		t.Errorf("DiscID = %q, want %q", info.DiscID, "SEGADISCSYSTEM")
	}
	if info.DiscType != DiscTypeBootable {
		t.Errorf("DiscType = %d, want %d (bootable)", info.DiscType, DiscTypeBootable)
	}
	if info.SystemType != "SEGA MEGA DRIVE" {
		t.Errorf("SystemType = %q, want %q", info.SystemType, "SEGA MEGA DRIVE")
	}
	if info.Copyright != "(C)SEGA 1993.SEP" {
		t.Errorf("Copyright = %q, want %q", info.Copyright, "(C)SEGA 1993.SEP")
	}
	if info.DomesticTitle != "SONIC THE HEDGEHOG CD" {
		t.Errorf("DomesticTitle = %q, want %q", info.DomesticTitle, "SONIC THE HEDGEHOG CD")
	}
	if info.OverseasTitle != "SONIC CD" {
		t.Errorf("OverseasTitle = %q, want %q", info.OverseasTitle, "SONIC CD")
	}
	if info.SerialNumber != "GM MK-4407-00" {
		t.Errorf("SerialNumber = %q, want %q", info.SerialNumber, "GM MK-4407-00")
	}
	if len(info.Devices) != 1 || info.Devices[0] != DeviceJoypad3Button {
		t.Errorf("Devices = %v, want [J]", info.Devices)
	}
	expectedRegion := RegionDomestic60Hz | RegionOverseas60Hz | RegionOverseas50Hz
	if info.Region != expectedRegion {
		t.Errorf("Region = %d, want %d (JUE)", info.Region, expectedRegion)
	}
}

func TestParseSegaCD_BootDiscID(t *testing.T) {
	data := make([]byte, 0x200)
	copy(data[0x00:], "SEGABOOTDISC    ")
	copy(data[0x100:], "SEGA MEGA DRIVE ")

	info, err := ParseSegaCD(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatalf("ParseSegaCD failed: %v", err)
	}

	if info.DiscID != "SEGABOOTDISC" {
		t.Errorf("DiscID = %q, want %q", info.DiscID, "SEGABOOTDISC")
	}
	if info.DiscType != DiscTypeBootable {
		t.Errorf("DiscType = %d, want %d (bootable)", info.DiscType, DiscTypeBootable)
	}
}

func TestParseSegaCD_NonBootableDiscID(t *testing.T) {
	testCases := []struct {
		discID string
		want   string
	}{
		{"SEGADISC        ", "SEGADISC"},
		{"SEGADATADISC    ", "SEGADATADISC"},
	}

	for _, tc := range testCases {
		t.Run(tc.want, func(t *testing.T) {
			data := make([]byte, 0x200)
			copy(data[0x00:], tc.discID)
			copy(data[0x100:], "SEGA MEGA DRIVE ")

			info, err := ParseSegaCD(bytes.NewReader(data), int64(len(data)))
			if err != nil {
				t.Fatalf("ParseSegaCD failed: %v", err)
			}

			if info.DiscID != tc.want {
				t.Errorf("DiscID = %q, want %q", info.DiscID, tc.want)
			}
			if info.DiscType != DiscTypeNonBootable {
				t.Errorf("DiscType = %d, want %d (non-bootable)", info.DiscType, DiscTypeNonBootable)
			}
		})
	}
}

func TestParseSegaCD_InvalidDiscID(t *testing.T) {
	data := make([]byte, 0x200)
	copy(data[0x00:], "INVALID         ")

	_, err := ParseSegaCD(bytes.NewReader(data), int64(len(data)))
	if err == nil {
		t.Error("expected error for invalid disc ID, got nil")
	}
}

func TestParseSegaCD_TooSmall(t *testing.T) {
	data := make([]byte, 100)

	_, err := ParseSegaCD(bytes.NewReader(data), int64(len(data)))
	if err == nil {
		t.Error("expected error for too-small input, got nil")
	}
}

func TestParseSegaCD_HexRegionCode(t *testing.T) {
	// Test new-style hex region code (single digit bitfield)
	data := make([]byte, 0x200)
	copy(data[0x00:], "SEGADISCSYSTEM  ")
	copy(data[0x100:], "SEGA MEGA DRIVE ")
	// '5' = 0101 binary = Japan (bit 0) + USA (bit 2)
	data[0x1F0] = '5'

	info, err := ParseSegaCD(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatalf("ParseSegaCD failed: %v", err)
	}

	expectedRegion := RegionDomestic60Hz | RegionOverseas60Hz
	if info.Region != expectedRegion {
		t.Errorf("Region = %d, want %d (Japan+USA)", info.Region, expectedRegion)
	}
}

func TestParseSegaCD_AllDevices(t *testing.T) {
	data := make([]byte, 0x200)
	copy(data[0x00:], "SEGADISCSYSTEM  ")
	copy(data[0x100:], "SEGA MEGA DRIVE ")
	copy(data[0x190:], "J6KMTLPA40")

	info, err := ParseSegaCD(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatalf("ParseSegaCD failed: %v", err)
	}

	expectedDevices := []Device{
		DeviceJoypad3Button,
		DeviceJoypad6Button,
		DeviceKeyboard,
		DeviceMouse,
		DeviceTrackball,
		DeviceLightgun,
		DevicePaddle,
		DeviceActivator,
		DeviceTeamPlayer,
		DeviceMasterSystem,
	}

	if len(info.Devices) != len(expectedDevices) {
		t.Errorf("Devices count = %d, want %d", len(info.Devices), len(expectedDevices))
	}
	for i, d := range expectedDevices {
		if i < len(info.Devices) && info.Devices[i] != d {
			t.Errorf("Devices[%d] = %q, want %q", i, info.Devices[i], d)
		}
	}
}
