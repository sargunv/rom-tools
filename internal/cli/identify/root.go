package identify

import (
	"cmp"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"

	"github.com/sargunv/rom-tools/internal/format"
	"github.com/sargunv/rom-tools/lib/core"
	romident "github.com/sargunv/rom-tools/lib/identify"

	"github.com/spf13/cobra"
)

var (
	jsonOutput  bool
	maxHashSize int64
)

var Cmd = &cobra.Command{
	Use:   "identify <file>...",
	Short: "Identify ROM files and extract metadata",
	Long: `Extract hashes and game identification data from ROM files.

Supports:
- Platform specific ROMs: identifies game information from the ROM header. Supported formats:
  - Famicom (NES): .nes
  - Super Famicom (SNES): .sfc, .smc
  - Nintendo 64: .z64, .v64, .n64
  - Nintendo GameCube / Wii: .gcm, .iso, .rvz, .wia
  - Nintendo Game Boy / Color: .gb, .gbc
  - Nintendo Game Boy Advance: .gba
  - Nintendo DS: .nds, .dsi, .ids
  - Nintendo 3DS: .3ds, .cci
  - Sega Master System / Game Gear: .sms, .gg
  - Sega Mega Drive (Genesis): .md, .gen, .smd, .32x
  - Sega CD: .bin, .chd
  - Sega Saturn: .bin, .chd
  - Sega Dreamcast: .bin, .chd
  - Sony PlayStation 1: .bin, .chd
  - Sony PlayStation 2: .iso, .bin, .chd
  - Sony PlayStation 3: .pkg
  - Sony PlayStation Portable: .iso, .chd
  - Sony PlayStation Vita: .pkg
  - Microsoft Xbox: .iso, .chd, .xbe
- .chd discs: extracts SHA1 hashes from header (no decompression needed)
- .zip archives: extracts CRC32 hashes from metadata (no decompression needed)
- All files: calculates SHA1, MD5, CRC32 for uncompressed files under --max-hash-size
- All folders: identifies files within`,
	Args: cobra.MinimumNArgs(1),
	RunE: runIdentify,
}

func init() {
	defaults := romident.DefaultOptions()

	Cmd.Flags().BoolVarP(&jsonOutput, "json", "j", false, "Output results as JSON Lines (one JSON object per line)")
	Cmd.Flags().Int64Var(&maxHashSize, "max-hash-size", defaults.MaxHashSize,
		"Max file size in bytes for hash calculation (-1 = no limit)")
}

func runIdentify(cmd *cobra.Command, args []string) error {
	opts := romident.Options{
		MaxHashSize: maxHashSize,
	}

	first := true

	for _, path := range args {
		result, err := romident.Identify(path, opts)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to identify %s: %v\n", path, err)
			continue
		}

		if jsonOutput {
			outputJSONLine(result)
		} else {
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
			return cmp.Compare(a.Name, b.Name)
		})

		for _, item := range items {
			fmt.Printf("  %s\n", item.Name)
			fmt.Printf("    Size: %s\n", formatSize(item.Size))

			if len(item.Hashes) > 0 {
				fmt.Println("    Hashes:")
				// Sort hash types for consistent output
				hashTypes := make([]core.HashType, 0, len(item.Hashes))
				for ht := range item.Hashes {
					hashTypes = append(hashTypes, ht)
				}
				slices.SortFunc(hashTypes, func(a, b core.HashType) int {
					return cmp.Compare(a, b)
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
