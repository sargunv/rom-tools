package identify

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"

	"github.com/sargunv/rom-tools/internal/format"
	romident "github.com/sargunv/rom-tools/lib/identify"

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
  - Famicom: .nes
  - Super Famicom: .sfc, .smc
  - Nintendo 64: .z64, .v64, .n64
  - Nintendo GameCube / Wii: .gcm, .iso, .rvz, .wia
  - Nintendo Game Boy / Color: .gb, .gbc
  - Nintendo Game Boy Advance: .gba
  - Nintendo DS: .nds, .dsi, .ids
  - Sega Master System: .sms
  - Sega Mega Drive / Genesis: .md, .gen, .smd
  - Sega Saturn: .bin
  - Sega Dreamcast: .bin
  - Sega Game Gear: .gg
  - Sony PlayStation 1: .bin
  - Sony PlayStation 2: .iso, .bin
  - Sony PlayStation Portable: .iso
  - Microsoft Xbox: .iso, .xbe
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
		result, err := romident.Identify(path, opts)

		if jsonOutput {
			if err != nil {
				// For JSON output, include errors in the output
				outputJSONLine(&romident.Result{Path: path, Error: err.Error()})
			} else {
				outputJSONLine(result)
			}
		} else {
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: failed to identify %s: %v\n", path, err)
				continue
			}
			if !first {
				fmt.Println()
			}
			outputText(result)
			first = false
		}
	}

	return nil
}

func outputJSONLine(result *romident.Result) {
	output, err := json.Marshal(result)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to marshal JSON: %v\n", err)
		return
	}
	fmt.Println(string(output))
}

func outputText(result *romident.Result) {
	baseName := filepath.Base(result.Path)

	// Determine type label
	typeLabel := "file"
	if len(result.Items) > 1 {
		typeLabel = "container"
	}

	fmt.Println(format.HeaderStyle.Render(fmt.Sprintf("ROM (%s): %s", typeLabel, baseName)))

	// Items (sorted by name for consistent output)
	if len(result.Items) > 0 {
		fmt.Println(format.HeaderStyle.Render("Items:"))

		// Sort by name
		items := make([]romident.Item, len(result.Items))
		copy(items, result.Items)
		slices.SortFunc(items, func(a, b romident.Item) int {
			if a.Name < b.Name {
				return -1
			}
			if a.Name > b.Name {
				return 1
			}
			return 0
		})

		for _, item := range items {
			fmt.Printf("  %s\n", item.Name)
			fmt.Printf("    Size: %s\n", formatSize(item.Size))
			if item.Format != "" && item.Format != romident.FormatUnknown {
				fmt.Printf("    Format: %s\n", item.Format)
			}

			if len(item.Hashes) > 0 {
				fmt.Println("    Hashes:")
				// Sort hash types for consistent output
				hashTypes := make([]romident.HashType, 0, len(item.Hashes))
				for ht := range item.Hashes {
					hashTypes = append(hashTypes, ht)
				}
				slices.SortFunc(hashTypes, func(a, b romident.HashType) int {
					if a < b {
						return -1
					}
					if a > b {
						return 1
					}
					return 0
				})
				for _, ht := range hashTypes {
					fmt.Printf("      %s: %s\n",
						format.LabelStyle.Render(string(ht)),
						item.Hashes[ht])
				}
			}

			if item.Game != nil {
				fmt.Println("    Game:")
				if item.Game.GamePlatform() != "" {
					fmt.Printf("      Platform: %s\n", item.Game.GamePlatform())
				}
				if item.Game.GameTitle() != "" {
					fmt.Printf("      Title: %s\n", item.Game.GameTitle())
				}
				if item.Game.GameSerial() != "" {
					fmt.Printf("      Serial: %s\n", item.Game.GameSerial())
				}
			}
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
