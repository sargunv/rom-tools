package playstation_cnf

import (
	"bytes"
	"testing"

	"github.com/sargunv/rom-tools/lib/core"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name         string
		content      string
		wantPlatform core.Platform
		wantBootPath string
		wantDiscID   string
		wantVersion  string
		wantVMode    VideoMode
	}{
		{
			name: "PS2 NTSC",
			content: `BOOT2 = cdrom0:\SLUS_123.45;1
VER = 1.00
VMODE = NTSC`,
			wantPlatform: core.PlatformPS2,
			wantBootPath: `cdrom0:\SLUS_123.45;1`,
			wantDiscID:   "SLUS_123.45",
			wantVersion:  "1.00",
			wantVMode:    VideoModeNTSC,
		},
		{
			name: "PS2 PAL",
			content: `BOOT2 = cdrom0:\SLES_543.21;1
VER = 2.01
VMODE = PAL`,
			wantPlatform: core.PlatformPS2,
			wantBootPath: `cdrom0:\SLES_543.21;1`,
			wantDiscID:   "SLES_543.21",
			wantVersion:  "2.01",
			wantVMode:    VideoModePAL,
		},
		{
			name: "PS1 standard",
			content: `BOOT = cdrom:\SCUS_943.00;1
TCB = 4
EVENT = 16`,
			wantPlatform: core.PlatformPS1,
			wantBootPath: `cdrom:\SCUS_943.00;1`,
			wantDiscID:   "SCUS_943.00",
			wantVersion:  "",
			wantVMode:    "",
		},
		{
			name:         "PS1 without backslash (Wipeout style)",
			content:      "BOOT = cdrom:SCUS_943.01",
			wantPlatform: core.PlatformPS1,
			wantBootPath: "cdrom:SCUS_943.01",
			wantDiscID:   "SCUS_943.01",
		},
		{
			name: "PS2 with BOOT and BOOT2 (BOOT2 takes precedence)",
			content: `BOOT = cdrom:\IGNORED.00;1
BOOT2 = cdrom0:\SLPM_123.45;1
VER = 1.00`,
			wantPlatform: core.PlatformPS2,
			wantBootPath: `cdrom0:\SLPM_123.45;1`,
			wantDiscID:   "SLPM_123.45",
			wantVersion:  "1.00",
		},
		{
			name:         "PS2 minimal (only BOOT2)",
			content:      `BOOT2 = cdrom0:\SCKA_200.01;1`,
			wantPlatform: core.PlatformPS2,
			wantBootPath: `cdrom0:\SCKA_200.01;1`,
			wantDiscID:   "SCKA_200.01",
		},
		{
			name: "extra whitespace",
			content: `  BOOT2   =   cdrom0:\SLUS_999.99;1
  VER   =   1.23  `,
			wantPlatform: core.PlatformPS2,
			wantBootPath: `cdrom0:\SLUS_999.99;1`,
			wantDiscID:   "SLUS_999.99",
			wantVersion:  "1.23",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := []byte(tt.content)
			info, err := Parse(bytes.NewReader(data), int64(len(data)))
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}
			if info.Platform != tt.wantPlatform {
				t.Errorf("Platform = %v, want %v", info.Platform, tt.wantPlatform)
			}
			if info.BootPath != tt.wantBootPath {
				t.Errorf("BootPath = %q, want %q", info.BootPath, tt.wantBootPath)
			}
			if info.DiscID != tt.wantDiscID {
				t.Errorf("DiscID = %q, want %q", info.DiscID, tt.wantDiscID)
			}
			if info.Version != tt.wantVersion {
				t.Errorf("Version = %q, want %q", info.Version, tt.wantVersion)
			}
			if info.VideoMode != tt.wantVMode {
				t.Errorf("VideoMode = %q, want %q", info.VideoMode, tt.wantVMode)
			}
		})
	}
}

func TestParse_Errors(t *testing.T) {
	tests := []struct {
		name    string
		content string
	}{
		{
			name:    "empty file",
			content: "",
		},
		{
			name:    "no boot path",
			content: "VER = 1.00\nVMODE = NTSC",
		},
		{
			name:    "malformed lines only",
			content: "this is not a valid cnf\nno equals signs here",
		},
		{
			name:    "wrong key names",
			content: "BOOTSTRAP = cdrom0:\\SLUS_123.45;1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := []byte(tt.content)
			_, err := Parse(bytes.NewReader(data), int64(len(data)))
			if err == nil {
				t.Error("Parse() expected error, got nil")
			}
		})
	}
}

func TestExtractDiscID(t *testing.T) {
	tests := []struct {
		name     string
		bootPath string
		want     string
	}{
		{
			name:     "standard PS2 with backslash and ;1",
			bootPath: `cdrom0:\SLUS_123.45;1`,
			want:     "SLUS_123.45",
		},
		{
			name:     "standard PS1 with backslash and ;1",
			bootPath: `cdrom:\SCUS_943.00;1`,
			want:     "SCUS_943.00",
		},
		{
			name:     "PS1 without backslash (Wipeout case)",
			bootPath: "cdrom:SCUS_943.01",
			want:     "SCUS_943.01",
		},
		{
			name:     "no ;1 suffix",
			bootPath: `cdrom0:\SLES_123.45`,
			want:     "SLES_123.45",
		},
		{
			name:     "forward slash separator",
			bootPath: "cdrom0:/SLUS_999.99;1",
			want:     "SLUS_999.99",
		},
		{
			name:     "mixed separators - backslash after colon",
			bootPath: `cdrom0:\path\SCPS_123.45;1`,
			want:     "SCPS_123.45",
		},
		{
			name:     "mixed separators - forward slash after colon",
			bootPath: "cdrom0:/path/SCPS_123.45;1",
			want:     "SCPS_123.45",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractDiscID(tt.bootPath)
			if got != tt.want {
				t.Errorf("extractDiscID(%q) = %q, want %q", tt.bootPath, got, tt.want)
			}
		})
	}
}
