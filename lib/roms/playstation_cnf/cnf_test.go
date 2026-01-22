package playstation_cnf

import "testing"

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
