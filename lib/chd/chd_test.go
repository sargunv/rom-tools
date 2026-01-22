package chd

import (
	"os"
	"testing"
)

func TestParseTrackMetadataEntry(t *testing.T) {
	tests := []struct {
		name    string
		data    string
		want    Track
		wantErr bool
	}{
		{
			name: "CHTR format",
			data: "TRACK:1 TYPE:MODE1_RAW SUBTYPE:NONE FRAMES:337350",
			want: Track{
				Number: 1,
				Type:   "MODE1_RAW",
				Frames: 337350,
			},
		},
		{
			name: "audio track",
			data: "TRACK:2 TYPE:AUDIO SUBTYPE:RW FRAMES:15000",
			want: Track{
				Number: 2,
				Type:   "AUDIO",
				Frames: 15000,
			},
		},
		{
			name: "CHT2 format with pregap",
			data: "TRACK:1 TYPE:MODE2_RAW SUBTYPE:RW_RAW FRAMES:300000 PREGAP:150 PGTYPE:MODE2 PGSUB:NONE POSTGAP:75",
			want: Track{
				Number: 1,
				Type:   "MODE2_RAW",
				Frames: 300000,
				Pregap: 150,
			},
		},
		{
			name: "CHGD format",
			data: "TRACK:3 TYPE:MODE1_RAW SUBTYPE:NONE FRAMES:450000 PAD:100 PREGAP:150 PGTYPE:MODE1 PGSUB:NONE POSTGAP:0",
			want: Track{
				Number: 3,
				Type:   "MODE1_RAW",
				Frames: 450000,
				Pregap: 150,
			},
		},
		{
			name:    "empty data",
			data:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseTrackMetadataEntry([]byte(tt.data))
			if (err != nil) != tt.wantErr {
				t.Errorf("parseTrackMetadataEntry() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}
			if got.Number != tt.want.Number {
				t.Errorf("Number = %v, want %v", got.Number, tt.want.Number)
			}
			if got.Type != tt.want.Type {
				t.Errorf("Type = %v, want %v", got.Type, tt.want.Type)
			}
			if got.Frames != tt.want.Frames {
				t.Errorf("Frames = %v, want %v", got.Frames, tt.want.Frames)
			}
			if got.Pregap != tt.want.Pregap {
				t.Errorf("Pregap = %v, want %v", got.Pregap, tt.want.Pregap)
			}
		})
	}
}

func TestNewReader(t *testing.T) {
	chdPath := "testdata/empty.chd"

	file, err := os.Open(chdPath)
	if err != nil {
		t.Fatalf("Failed to open CHD file: %v", err)
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		t.Fatalf("Failed to stat CHD file: %v", err)
	}

	reader, err := NewReader(file, stat.Size())
	if err != nil {
		t.Fatalf("NewReader() error = %v", err)
	}

	// Verify header is accessible
	header := reader.Header()
	if header == nil {
		t.Fatal("Header() returned nil")
	}
	if header.Version < 5 {
		t.Errorf("Expected version >= 5, got %d", header.Version)
	}

	if header.RawSHA1 != "f6348f85d8487e7aff1fa54e5987b172bce2a3a6" {
		t.Errorf("Expected raw SHA1 'f6348f85d8487e7aff1fa54e5987b172bce2a3a6', got '%s'", header.RawSHA1)
	}

	if header.SHA1 != "cdd8baa51e7b84bb11037fb3415d698d011fe40a" {
		t.Errorf("Expected compressed SHA1 'cdd8baa51e7b84bb11037fb3415d698d011fe40a', got '%s'", header.SHA1)
	}

	// Parent SHA1 should be empty for standalone CHD
	if header.ParentSHA1 != "" {
		t.Errorf("Expected empty parent SHA1, got '%s'", header.ParentSHA1)
	}

	// Verify size accessor
	if reader.Size() <= 0 {
		t.Errorf("Size() = %d, expected positive value", reader.Size())
	}

	// empty.chd has no track metadata (it's a simple test file)
	// Tracks should be nil or empty
	t.Logf("Tracks count: %d", len(reader.Tracks))
}

func TestTrackSize(t *testing.T) {
	track := &Track{Frames: 100}
	want := int64(100 * 2352) // rawSectorSize = 2352
	if got := track.Size(); got != want {
		t.Errorf("Track.Size() = %v, want %v", got, want)
	}
}
