package identify

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/sargunv/rom-tools/internal/format"
	"github.com/sargunv/rom-tools/lib/romident"

	"github.com/spf13/cobra"
)

var (
	jsonOutput bool
	fastMode   bool
	slowMode   bool
)

var Cmd = &cobra.Command{
	Use:   "identify <file>...",
	Short: "Identify ROM files and extract metadata",
	Long: `Extract hashes and game identification data from ROM files.

Supports:
- Platform specific ROMs: identifies game information from the ROM header, depending on the format. Supported ROM formats:
  - Nintendo Entertainment System: .nes
  - Super Nintendo Entertainment System: .sfc, .smc
  - Nintendo 64: .z64, .v64, .n64
  - Nintendo GameCube / Wii: .gcm, .iso, .rvz, .wia
  - Nintendo Game Boy / Color: .gb, .gbc
  - Nintendo Game Boy Advance: .gba
  - Nintendo DS: .nds, .dsi, .ids
  - Nintendo 3DS: TODO
  - Sega Master System: TODO
  - Sega Mega Drive / Genesis: .md, .gen, .smd
  - Sega Saturn: TODO
  - Sega Dreamcast: TODO
  - Sega Game Gear: TODO
  - Sony PlayStation 1: .bin/cue
  - Sony PlayStation 2: .iso, .bin/cue
  - Sony PlayStation Portable: TODO
  - Microsoft Xbox: .xiso, .xiso.iso, and .xbe
- .chd discs: extracts SHA1 hashes from header (fast, no decompression)
- .zip archives: extracts CRC32 from metadata (fast, no decompression). If in slow mode, also identifies files within the ZIP.
- all files: calculates SHA1, MD5, CRC32 (unless in fast mode).
- all folders: identifies files within.`,
	Args: cobra.MinimumNArgs(1),
	RunE: runIdentify,
}

func init() {
	Cmd.Flags().BoolVarP(&jsonOutput, "json", "j", false, "Output results as JSON Lines (one JSON object per line)")
	Cmd.Flags().BoolVar(&fastMode, "fast", false, "Skip hash calculation for large loose files, but calculates for small loose files (<65MiB).")
	Cmd.Flags().BoolVar(&slowMode, "slow", false, "Calculate full hashes and identify games inside archives (requires decompression).")
	Cmd.MarkFlagsMutuallyExclusive("fast", "slow")
}

func runIdentify(cmd *cobra.Command, args []string) error {
	opts := romident.Options{HashMode: romident.HashModeDefault}
	if fastMode {
		opts.HashMode = romident.HashModeFast
	} else if slowMode {
		opts.HashMode = romident.HashModeSlow
	}

	first := true

	for _, path := range args {
		rom, err := romident.IdentifyROM(path, opts)

		if jsonOutput {
			if err != nil {
				// For JSON output, include errors in the output
				outputJSONLine(&romident.ROM{Path: path}, err)
			} else {
				outputJSONLine(rom, nil)
			}
		} else {
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: failed to identify %s: %v\n", path, err)
				continue
			}
			if !first {
				fmt.Println()
			}
			outputTextSingle(rom)
			first = false
		}
	}

	return nil
}

// JSONResult wraps a ROM result with an optional error for JSON output.
type JSONResult struct {
	*romident.ROM
	Error string `json:"error,omitempty"`
}

func outputJSONLine(rom *romident.ROM, err error) {
	result := JSONResult{ROM: rom}
	if err != nil {
		result.Error = err.Error()
	}

	output, marshalErr := json.Marshal(result)
	if marshalErr != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to marshal JSON: %v\n", marshalErr)
		return
	}
	fmt.Println(string(output))
}

func outputTextSingle(rom *romident.ROM) {
	// Header
	baseName := filepath.Base(rom.Path)
	fmt.Println(format.HeaderStyle.Render(fmt.Sprintf("ROM (%s): %s", rom.Type, baseName)))

	// Files (sorted by path for consistent output)
	if len(rom.Files) > 0 {
		fmt.Println(format.HeaderStyle.Render("Files:"))

		// Sort file paths
		paths := make([]string, 0, len(rom.Files))
		for path := range rom.Files {
			paths = append(paths, path)
		}
		slices.Sort(paths)

		for _, path := range paths {
			f := rom.Files[path]
			prefix := "  "
			if f.IsPrimary {
				prefix = "* "
			}

			fmt.Printf("%s%s\n", prefix, path)
			fmt.Printf("    Size: %s\n", formatSize(f.Size))
			if f.Format != romident.FormatUnknown {
				fmt.Printf("    Format: %s\n", f.Format)
			}

			if len(f.Hashes) > 0 {
				fmt.Println("    Hashes:")
				for _, h := range f.Hashes {
					fmt.Printf("      %s: %s (%s)\n",
						format.LabelStyle.Render(string(h.Algorithm)),
						h.Value,
						h.Source)
				}
			}
		}
	}

	// Identification
	if rom.Ident != nil {
		fmt.Println(format.HeaderStyle.Render("Identification:"))
		fmt.Printf("  Platform: %s\n", rom.Ident.Platform)
		if rom.Ident.TitleID != "" {
			fmt.Printf("  Title ID: %s\n", rom.Ident.TitleID)
		}
		if rom.Ident.Title != "" {
			fmt.Printf("  Title: %s\n", rom.Ident.Title)
		}
		if len(rom.Ident.Regions) > 0 {
			regionStrs := make([]string, len(rom.Ident.Regions))
			for i, r := range rom.Ident.Regions {
				regionStrs[i] = string(r)
			}
			fmt.Printf("  Regions: %s\n", strings.Join(regionStrs, ", "))
		}
		if rom.Ident.MakerCode != "" {
			fmt.Printf("  Maker Code: %s\n", rom.Ident.MakerCode)
		}
		if rom.Ident.Version != nil {
			fmt.Printf("  Version: %d\n", *rom.Ident.Version)
		}
		if rom.Ident.DiscNumber != nil {
			fmt.Printf("  Disc Number: %d\n", *rom.Ident.DiscNumber)
		}
		if rom.Ident.Extra != nil {
			extraJSON, _ := json.Marshal(rom.Ident.Extra)
			fmt.Printf("  Extra: %s\n", string(extraJSON))
		}
	}
}

func formatSize(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.2f GiB", float64(bytes)/GB)
	case bytes >= MB:
		return fmt.Sprintf("%.2f MiB", float64(bytes)/MB)
	case bytes >= KB:
		return fmt.Sprintf("%.2f KiB", float64(bytes)/KB)
	default:
		return fmt.Sprintf("%d bytes", bytes)
	}
}
