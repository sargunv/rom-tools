package identify

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"

	"github.com/sargunv/rom-tools/clients/romident"
	"github.com/sargunv/rom-tools/internal/format"

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
  - Single files: calculates SHA1, MD5, CRC32
  - ZIP archives: extracts CRC32 from metadata (fast, no decompression)
  - CHD files: extracts SHA1 hashes from header (fast, no decompression)
  - Folders: identifies all files within

Format detection:
  - CHD: by magic bytes
  - Xbox XISO: by magic bytes at offset 0x10000
  - ISO9660: by magic bytes
  - ZIP: by magic bytes

Hash modes:
  - Default: uses fast methods where available, calculates for loose files
  - --fast: skips hash calculation entirely
  - --slow: calculates full hashes even when fast methods are available`,
	Args: cobra.MinimumNArgs(1),
	RunE: runIdentify,
}

func init() {
	Cmd.Flags().BoolVarP(&jsonOutput, "json", "j", false, "Output results as JSON Lines (one JSON object per line)")
	Cmd.Flags().BoolVar(&fastMode, "fast", false, "Skip hash calculation entirely")
	Cmd.Flags().BoolVar(&slowMode, "slow", false, "Calculate full hashes even for archives (requires decompression)")
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
		if rom.Ident.Region != "" {
			fmt.Printf("  Region: %s\n", rom.Ident.Region)
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
