package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	screenscraper "sargunv/screenscraper-go/client"

	"github.com/spf13/cobra"
)

var (
	proposeGameID     string
	proposeROMID      string
	proposeType       string
	proposeText       string
	proposeRegion     string
	proposeLanguage   string
	proposeVersion    string
	proposeSource     string
	proposeFile       string
	proposeURL        string
	proposeSupportNum string
)

var proposeCmd = &cobra.Command{
	Use:   "propose",
	Short: "Submit proposals to ScreenScraper",
	Long: `Submit info or media proposals to contribute to the ScreenScraper database.

This command requires user credentials (SCREENSCRAPER_ID and SCREENSCRAPER_PASSWORD).
Your proposals will be associated with your ScreenScraper account and reviewed by moderators.`,
}

var proposeInfoCmd = &cobra.Command{
	Use:   "info",
	Short: "Submit a text info proposal",
	Long: `Submit a text info proposal for a game or ROM.

Game info types (--game-id): name, editeur, developpeur, players, score,
rating, genres, datessortie, rotation, resolution, modes, familles, numero,
styles, themes, description

ROM info types (--rom-id): developpeur, editeur, datessortie, players,
regions, langues, clonetype, hacktype, friendly, serial, description`,
	Example: `  # Add a game name for US region
  screenscraper propose info --game-id=123 --type=name --text="Super Mario Bros." --region=us

  # Add a synopsis in English
  screenscraper propose info --game-id=123 --type=description --text="A classic platformer..." --language=en

  # Add a ROM serial number
  screenscraper propose info --rom-id=456 --type=serial --text="SLUS-01234"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Validate that exactly one of game-id or rom-id is provided
		if (proposeGameID == "" && proposeROMID == "") || (proposeGameID != "" && proposeROMID != "") {
			return fmt.Errorf("exactly one of --game-id or --rom-id must be specified")
		}

		params := screenscraper.SubmitInfoProposalParams{
			GameID:   proposeGameID,
			ROMID:    proposeROMID,
			InfoType: proposeType,
			Text:     proposeText,
			Region:   proposeRegion,
			Language: proposeLanguage,
			Version:  proposeVersion,
			Source:   proposeSource,
		}

		resp, err := client.SubmitInfoProposal(params)
		if err != nil {
			return err
		}

		if jsonOutput {
			formatted, err := json.MarshalIndent(resp, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to format JSON: %w", err)
			}
			fmt.Println(string(formatted))
			return nil
		}

		fmt.Println(resp.Message)
		return nil
	},
}

var proposeMediaCmd = &cobra.Command{
	Use:   "media",
	Short: "Submit a media proposal",
	Long: `Submit a media proposal for a game or ROM.

Media types: sstitle, ss, fanart, video, overlay, steamgrid, wheel, wheel-hd,
marquee, screenmarquee, box-2D, box-2D-side, box-2D-back, box-texture, manuel,
flyer, maps, figurine, support-texture, box-scan, support-scan, bezel-4-3,
bezel-4-3-v, bezel-4-3-cocktail, bezel-16-9, bezel-16-9-v, bezel-16-9-cocktail,
wheel-tarcisios, videotable, videotable4k, themehs, themehb

You can provide the media either as a file (--file) or URL (--url).
Use --file=- to read from stdin.`,
	Example: `  # Upload box art from file
  screenscraper propose media --game-id=123 --type=box-2D --file=box_us.png --region=us

  # Submit media from URL
  screenscraper propose media --game-id=123 --type=wheel --url="https://example.com/logo.png" --region=eu

  # Upload screenshot from stdin
  cat screenshot.jpg | screenscraper propose media --game-id=123 --type=ss --file=- --region=us`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Validate that exactly one of game-id or rom-id is provided
		if (proposeGameID == "" && proposeROMID == "") || (proposeGameID != "" && proposeROMID != "") {
			return fmt.Errorf("exactly one of --game-id or --rom-id must be specified")
		}

		// Validate that exactly one of file or url is provided
		if (proposeFile == "" && proposeURL == "") || (proposeFile != "" && proposeURL != "") {
			return fmt.Errorf("exactly one of --file or --url must be specified")
		}

		params := screenscraper.SubmitMediaProposalParams{
			GameID:        proposeGameID,
			ROMID:         proposeROMID,
			MediaType:     proposeType,
			MediaFileURL:  proposeURL,
			Region:        proposeRegion,
			SupportNumber: proposeSupportNum,
			Version:       proposeVersion,
			Source:        proposeSource,
		}

		// Handle file input if provided
		if proposeFile != "" {
			var file *os.File
			var err error

			if proposeFile == "-" {
				// Read from stdin
				file = os.Stdin
				params.MediaFileName = "stdin"
			} else {
				// Read from file
				file, err = os.Open(proposeFile)
				if err != nil {
					return fmt.Errorf("failed to open file: %w", err)
				}
				defer file.Close()
				params.MediaFileName = filepath.Base(proposeFile)
			}

			params.MediaFile = file
		}

		resp, err := client.SubmitMediaProposal(params)
		if err != nil {
			return err
		}

		if jsonOutput {
			formatted, err := json.MarshalIndent(resp, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to format JSON: %w", err)
			}
			fmt.Println(string(formatted))
			return nil
		}

		fmt.Println(resp.Message)
		return nil
	},
}

func init() {
	// Info subcommand flags
	proposeInfoCmd.Flags().StringVar(&proposeGameID, "game-id", "", "Game ID to submit info for")
	proposeInfoCmd.Flags().StringVar(&proposeROMID, "rom-id", "", "ROM ID to submit info for")
	proposeInfoCmd.Flags().StringVarP(&proposeType, "type", "t", "", "Info type (e.g. name, editeur, description)")
	proposeInfoCmd.Flags().StringVar(&proposeText, "text", "", "The text content")
	proposeInfoCmd.Flags().StringVarP(&proposeRegion, "region", "r", "", "Region short name (required for name, datessortie)")
	proposeInfoCmd.Flags().StringVarP(&proposeLanguage, "language", "l", "", "Language short name (required for description)")
	proposeInfoCmd.Flags().StringVarP(&proposeVersion, "version", "v", "", "Version (optional)")
	proposeInfoCmd.Flags().StringVarP(&proposeSource, "source", "s", "", "Source URL or info (optional)")

	proposeInfoCmd.MarkFlagRequired("type")
	proposeInfoCmd.MarkFlagRequired("text")

	// Media subcommand flags
	proposeMediaCmd.Flags().StringVar(&proposeGameID, "game-id", "", "Game ID to submit media for")
	proposeMediaCmd.Flags().StringVar(&proposeROMID, "rom-id", "", "ROM ID to submit media for")
	proposeMediaCmd.Flags().StringVarP(&proposeType, "type", "t", "", "Media type (e.g. ss, box-2D, wheel)")
	proposeMediaCmd.Flags().StringVarP(&proposeFile, "file", "f", "", "File path to upload (use '-' for stdin)")
	proposeMediaCmd.Flags().StringVarP(&proposeURL, "url", "u", "", "URL of media to download")
	proposeMediaCmd.Flags().StringVarP(&proposeRegion, "region", "r", "", "Region (required for ss, sstitle, wheel, box-*, bezel-*, etc.)")
	proposeMediaCmd.Flags().StringVarP(&proposeSupportNum, "support-num", "n", "", "Support number 0-10 (required for box-*, flyer, support-*)")
	proposeMediaCmd.Flags().StringVarP(&proposeVersion, "version", "v", "", "Version (for maps, box-scan, support-scan)")
	proposeMediaCmd.Flags().StringVarP(&proposeSource, "source", "s", "", "Source URL or info (optional)")

	proposeMediaCmd.MarkFlagRequired("type")

	proposeCmd.AddCommand(proposeInfoCmd)
	proposeCmd.AddCommand(proposeMediaCmd)
	rootCmd.AddCommand(proposeCmd)
}
