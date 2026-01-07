package cli

import (
	"encoding/json"
	"fmt"

	screenscraper "sargunv/screenscraper-go/client"
	"sargunv/screenscraper-go/internal/format"

	"github.com/spf13/cobra"
)

var (
	gameCRC      string
	gameMD5      string
	gameSHA1     string
	gameSize     string
	gameSystemID string
	gameROMType  string
	gameROMName  string
	gameID       string
	gameSerial   string
)

var gameCmd = &cobra.Command{
	Use:   "game",
	Short: "Get game information",
	Long: `Retrieves detailed game information including metadata and media URLs.

You can lookup by:
  1. ROM hash (CRC/MD5/SHA1) + size + system + name + type (recommended)
  2. Game ID (direct lookup)`,
	Example: `  # Lookup by ROM hash
  screenscraper game --crc=50ABC90A --size=749652 --system=1 --rom-type=rom --name="Sonic 2.zip"
  
  # Lookup by game ID
  screenscraper game --id=3`,
	RunE: func(cmd *cobra.Command, args []string) error {
		params := screenscraper.GameInfoParams{
			CRC:       gameCRC,
			MD5:       gameMD5,
			SHA1:      gameSHA1,
			ROMSize:   gameSize,
			SystemID:  gameSystemID,
			ROMType:   gameROMType,
			ROMName:   gameROMName,
			GameID:    gameID,
			SerialNum: gameSerial,
		}

		// Validate we have at least one lookup method
		hasHash := gameCRC != "" || gameMD5 != "" || gameSHA1 != ""
		hasGameID := gameID != ""

		if !hasHash && !hasGameID {
			return fmt.Errorf("must provide either ROM hash (--crc/--md5/--sha1) or --id")
		}

		if hasHash && gameSystemID == "" {
			return fmt.Errorf("--system is required when using ROM hash lookup")
		}

		resp, err := client.GetGameInfo(params)
		if err != nil {
			return err
		}

		if jsonOutput {
			formatted, err := json.MarshalIndent(resp.Response.Game, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to format JSON: %w", err)
			}
			fmt.Println(string(formatted))
			return nil
		}

		lang := format.GetPreferredLanguage(locale)
		fmt.Print(format.RenderGame(resp.Response.Game, lang))
		return nil
	},
}

func init() {
	gameCmd.Flags().StringVar(&gameCRC, "crc", "", "ROM CRC32 hash")
	gameCmd.Flags().StringVar(&gameMD5, "md5", "", "ROM MD5 hash")
	gameCmd.Flags().StringVar(&gameSHA1, "sha1", "", "ROM SHA1 hash")
	gameCmd.Flags().StringVar(&gameSize, "size", "", "ROM size in bytes")
	gameCmd.Flags().StringVarP(&gameSystemID, "system", "s", "", "System ID (required for hash lookup)")
	gameCmd.Flags().StringVar(&gameROMType, "rom-type", "rom", "ROM type: rom, iso, or folder")
	gameCmd.Flags().StringVarP(&gameROMName, "name", "n", "", "ROM filename")
	gameCmd.Flags().StringVar(&gameID, "id", "", "Game ID (alternative to hash lookup)")
	gameCmd.Flags().StringVar(&gameSerial, "serial", "", "Serial number (optional)")

	detailCmd.AddCommand(gameCmd)
}
