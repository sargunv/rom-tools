package screenscraper

import (
	"fmt"
	"os"

	"github.com/sargunv/rom-tools/clients/screenscraper"
	"github.com/sargunv/rom-tools/internal/cli/screenscraper/shared"

	"github.com/spf13/cobra"
)

var (
	dlSystemID  string
	dlGameID    string
	dlGroupID   string
	dlCompanyID string
	dlMedia     string
	dlOutput    string
	dlMaxWidth  string
	dlMaxHeight string
	dlFormat    string
)

var downloadCmd = &cobra.Command{
	Use:   "download",
	Short: "Download media files",
	Long:  "Download game or system media files (images, videos, etc.)",
}

var downloadGameMediaCmd = &cobra.Command{
	Use:   "game",
	Short: "Download game media",
	Long:  `Download game media (box art, logo, screenshot, etc.)`,
	Example: `  # Download game box art
  rom-tools screenscraper download game --system=1 --game-id=3 --media="box-2D(us)" --output=box.png

  # Download game wheel logo
  rom-tools screenscraper download game -s 1 -g 3 -m "wheel-hd(eu)" -o logo.png`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if dlSystemID == "" || dlGameID == "" || dlMedia == "" {
			return fmt.Errorf("--system, --game-id, and --media are required")
		}

		params := screenscraper.DownloadMediaParams{
			SystemID:     dlSystemID,
			GameID:       dlGameID,
			Media:        dlMedia,
			MaxWidth:     dlMaxWidth,
			MaxHeight:    dlMaxHeight,
			OutputFormat: dlFormat,
		}

		data, err := shared.Client.DownloadGameMedia(params)
		if err != nil {
			return err
		}

		// Check if it's a text response (NOMEDIA, CRCOK, etc.)
		if len(data) < 100 && data[0] != 0xFF && data[0] != 0x89 { // not binary
			fmt.Printf("Response: %s\n", string(data))
			return nil
		}

		// Write to file or stdout
		if dlOutput == "" || dlOutput == "-" {
			_, err = os.Stdout.Write(data)
			return err
		}

		if err := os.WriteFile(dlOutput, data, 0644); err != nil {
			return fmt.Errorf("failed to write file: %w", err)
		}

		fmt.Printf("Saved to %s (%d bytes)\n", dlOutput, len(data))
		return nil
	},
}

var downloadSystemMediaCmd = &cobra.Command{
	Use:   "system",
	Short: "Download system media",
	Long:  `Download system media (logo, wheel, photos, etc.)`,
	Example: `  # Download system wheel logo
  rom-tools screenscraper download system --system=1 --media="wheel(wor)" --output=system.png`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if dlSystemID == "" || dlMedia == "" {
			return fmt.Errorf("--system and --media are required")
		}

		params := screenscraper.DownloadMediaParams{
			SystemID:     dlSystemID,
			Media:        dlMedia,
			MaxWidth:     dlMaxWidth,
			MaxHeight:    dlMaxHeight,
			OutputFormat: dlFormat,
		}

		data, err := shared.Client.DownloadSystemMedia(params)
		if err != nil {
			return err
		}

		// Check if it's a text response (NOMEDIA, CRCOK, etc.)
		if len(data) < 100 && data[0] != 0xFF && data[0] != 0x89 { // not binary
			fmt.Printf("Response: %s\n", string(data))
			return nil
		}

		// Write to file or stdout
		if dlOutput == "" || dlOutput == "-" {
			_, err = os.Stdout.Write(data)
			return err
		}

		if err := os.WriteFile(dlOutput, data, 0644); err != nil {
			return fmt.Errorf("failed to write file: %w", err)
		}

		fmt.Printf("Saved to %s (%d bytes)\n", dlOutput, len(data))
		return nil
	},
}

var downloadGroupMediaCmd = &cobra.Command{
	Use:   "group",
	Short: "Download group media",
	Long:  `Download group media (genres, modes, families, themes, styles)`,
	Example: `  # Download genre logo
  rom-tools screenscraper download group --group-id=1 --media="logo-monochrome" --output=genre.png`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if dlGroupID == "" || dlMedia == "" {
			return fmt.Errorf("--group-id and --media are required")
		}

		params := screenscraper.DownloadGroupMediaParams{
			GroupID:      dlGroupID,
			Media:        dlMedia,
			MaxWidth:     dlMaxWidth,
			MaxHeight:    dlMaxHeight,
			OutputFormat: dlFormat,
		}

		data, err := shared.Client.DownloadGroupMedia(params)
		if err != nil {
			return err
		}

		// Check if it's a text response (NOMEDIA, CRCOK, etc.)
		if len(data) < 100 && data[0] != 0xFF && data[0] != 0x89 { // not binary
			fmt.Printf("Response: %s\n", string(data))
			return nil
		}

		// Write to file or stdout
		if dlOutput == "" || dlOutput == "-" {
			_, err = os.Stdout.Write(data)
			return err
		}

		if err := os.WriteFile(dlOutput, data, 0644); err != nil {
			return fmt.Errorf("failed to write file: %w", err)
		}

		fmt.Printf("Saved to %s (%d bytes)\n", dlOutput, len(data))
		return nil
	},
}

var downloadCompanyMediaCmd = &cobra.Command{
	Use:   "company",
	Short: "Download company media",
	Long:  `Download company media (publishers, developers)`,
	Example: `  # Download company logo
  rom-tools screenscraper download company --company-id=3 --media="logo-monochrome" --output=company.png`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if dlCompanyID == "" || dlMedia == "" {
			return fmt.Errorf("--company-id and --media are required")
		}

		params := screenscraper.DownloadCompanyMediaParams{
			CompanyID:    dlCompanyID,
			Media:        dlMedia,
			MaxWidth:     dlMaxWidth,
			MaxHeight:    dlMaxHeight,
			OutputFormat: dlFormat,
		}

		data, err := shared.Client.DownloadCompanyMedia(params)
		if err != nil {
			return err
		}

		// Check if it's a text response (NOMEDIA, CRCOK, etc.)
		if len(data) < 100 && data[0] != 0xFF && data[0] != 0x89 { // not binary
			fmt.Printf("Response: %s\n", string(data))
			return nil
		}

		// Write to file or stdout
		if dlOutput == "" || dlOutput == "-" {
			_, err = os.Stdout.Write(data)
			return err
		}

		if err := os.WriteFile(dlOutput, data, 0644); err != nil {
			return fmt.Errorf("failed to write file: %w", err)
		}

		fmt.Printf("Saved to %s (%d bytes)\n", dlOutput, len(data))
		return nil
	},
}

func init() {
	// Game media flags
	downloadGameMediaCmd.Flags().StringVarP(&dlSystemID, "system", "s", "", "System ID (required)")
	downloadGameMediaCmd.Flags().StringVarP(&dlGameID, "game-id", "g", "", "Game ID (required)")
	downloadGameMediaCmd.Flags().StringVarP(&dlMedia, "media", "m", "", "Media identifier (required, e.g. 'box-2D(us)', 'wheel-hd(eu)')")
	downloadGameMediaCmd.Flags().StringVarP(&dlOutput, "output", "o", "", "Output file path (default: stdout)")
	downloadGameMediaCmd.Flags().StringVar(&dlMaxWidth, "max-width", "", "Maximum width in pixels")
	downloadGameMediaCmd.Flags().StringVar(&dlMaxHeight, "max-height", "", "Maximum height in pixels")
	downloadGameMediaCmd.Flags().StringVar(&dlFormat, "format", "", "Output format: png or jpg")

	// System media flags
	downloadSystemMediaCmd.Flags().StringVarP(&dlSystemID, "system", "s", "", "System ID (required)")
	downloadSystemMediaCmd.Flags().StringVarP(&dlMedia, "media", "m", "", "Media identifier (required)")
	downloadSystemMediaCmd.Flags().StringVarP(&dlOutput, "output", "o", "", "Output file path (default: stdout)")
	downloadSystemMediaCmd.Flags().StringVar(&dlMaxWidth, "max-width", "", "Maximum width in pixels")
	downloadSystemMediaCmd.Flags().StringVar(&dlMaxHeight, "max-height", "", "Maximum height in pixels")
	downloadSystemMediaCmd.Flags().StringVar(&dlFormat, "format", "", "Output format: png or jpg")

	// Group media flags
	downloadGroupMediaCmd.Flags().StringVarP(&dlGroupID, "group-id", "g", "", "Group ID (required)")
	downloadGroupMediaCmd.Flags().StringVarP(&dlMedia, "media", "m", "", "Media identifier (required, e.g. 'logo-monochrome')")
	downloadGroupMediaCmd.Flags().StringVarP(&dlOutput, "output", "o", "", "Output file path (default: stdout)")
	downloadGroupMediaCmd.Flags().StringVar(&dlMaxWidth, "max-width", "", "Maximum width in pixels")
	downloadGroupMediaCmd.Flags().StringVar(&dlMaxHeight, "max-height", "", "Maximum height in pixels")
	downloadGroupMediaCmd.Flags().StringVar(&dlFormat, "format", "", "Output format: png or jpg")

	// Company media flags
	downloadCompanyMediaCmd.Flags().StringVarP(&dlCompanyID, "company-id", "c", "", "Company ID (required)")
	downloadCompanyMediaCmd.Flags().StringVarP(&dlMedia, "media", "m", "", "Media identifier (required, e.g. 'logo-monochrome')")
	downloadCompanyMediaCmd.Flags().StringVarP(&dlOutput, "output", "o", "", "Output file path (default: stdout)")
	downloadCompanyMediaCmd.Flags().StringVar(&dlMaxWidth, "max-width", "", "Maximum width in pixels")
	downloadCompanyMediaCmd.Flags().StringVar(&dlMaxHeight, "max-height", "", "Maximum height in pixels")
	downloadCompanyMediaCmd.Flags().StringVar(&dlFormat, "format", "", "Output format: png or jpg")

	downloadCmd.AddCommand(downloadGameMediaCmd)
	downloadCmd.AddCommand(downloadSystemMediaCmd)
	downloadCmd.AddCommand(downloadGroupMediaCmd)
	downloadCmd.AddCommand(downloadCompanyMediaCmd)
	Cmd.AddCommand(downloadCmd)
}
