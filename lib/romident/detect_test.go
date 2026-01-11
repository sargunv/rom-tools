package romident

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sargunv/rom-tools/internal/testutil"
)

func TestDetect(t *testing.T) {
	detector := NewDetector()

	tests := []struct {
		name     string
		filename string
		want     Format
	}{
		{
			name:     "CHD file",
			filename: "empty.chd",
			want:     FormatCHD,
		},
		{
			name:     "ZIP file",
			filename: "gbtictac.gb.zip",
			want:     FormatZIP,
		},
		{
			name:     "XISO file",
			filename: "xromwell.xiso.iso",
			want:     FormatXISO,
		},
		{
			name:     "XBE file",
			filename: "xromwell/default.xbe",
			want:     FormatXBE,
		},
		{
			name:     "GBA file",
			filename: "AGB_Rogue.gba",
			want:     FormatGBA,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filePath := filepath.Join(testutil.ROMsPath(t), tt.filename)

			file, err := os.Open(filePath)
			if err != nil {
				t.Fatalf("Failed to open file: %v", err)
			}
			defer file.Close()

			stat, err := file.Stat()
			if err != nil {
				t.Fatalf("Failed to stat file: %v", err)
			}

			got, err := detector.Detect(file, stat.Size(), tt.filename)
			if err != nil {
				t.Fatalf("Detect() error = %v", err)
			}

			if got != tt.want {
				t.Errorf("Detect() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCandidatesByExtension(t *testing.T) {
	detector := NewDetector()

	tests := []struct {
		name     string
		filename string
		want     []Format
	}{
		{
			name:     "CHD extension",
			filename: "test.chd",
			want:     []Format{FormatCHD},
		},
		{
			name:     "ZIP extension",
			filename: "test.zip",
			want:     []Format{FormatZIP},
		},
		{
			name:     "XISO extension",
			filename: "test.xiso",
			want:     []Format{FormatXISO},
		},
		{
			name:     "XBE extension",
			filename: "default.xbe",
			want:     []Format{FormatXBE},
		},
		{
			name:     "GBA extension",
			filename: "test.gba",
			want:     []Format{FormatGBA},
		},
		{
			name:     "ISO extension (ambiguous)",
			filename: "test.iso",
			want:     []Format{FormatXISO, FormatISO9660},
		},
		{
			name:     "Unknown extension",
			filename: "test.unknown",
			want:     nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := detector.CandidatesByExtension(tt.filename)
			if len(got) != len(tt.want) {
				t.Errorf("CandidatesByExtension() = %v, want %v", got, tt.want)
				return
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("CandidatesByExtension() = %v, want %v", got, tt.want)
					return
				}
			}
		})
	}
}
