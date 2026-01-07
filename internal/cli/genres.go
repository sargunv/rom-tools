package cli

import (
	"encoding/json"
	"fmt"
	"strings"

	screenscraper "sargunv/screenscraper-go/client"
	"sargunv/screenscraper-go/internal/format"

	"github.com/spf13/cobra"
)

var genresCmd = &cobra.Command{
	Use:   "genres",
	Short: "Get list of genres",
	Long:  "Retrieves the list of all game genres",
	RunE: func(cmd *cobra.Command, args []string) error {
		resp, err := client.GetGenresList()
		if err != nil {
			return err
		}

		if jsonOutput {
			formatted, err := json.MarshalIndent(resp.Response.Genres, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to format JSON: %w", err)
			}
			fmt.Println(string(formatted))
			return nil
		}

		lang := format.GetPreferredLanguage(locale)
		fmt.Print(format.RenderGenresList(resp.Response.Genres, lang))
		return nil
	},
}

var (
	genreDetailID   int
	genreDetailName string
)

var genreDetailCmd = &cobra.Command{
	Use:   "genre",
	Short: "Get detailed genre information",
	Long:  "Retrieves detailed information about a specific genre",
	Example: `  screenscraper detail genre --id=1
  screenscraper detail genre --name=action`,
	RunE: func(cmd *cobra.Command, args []string) error {
		resp, err := client.GetGenresList()
		if err != nil {
			return err
		}

		// Find genre by ID or name
		var found *screenscraper.Genre
		for k, genre := range resp.Response.Genres {
			if genreDetailID != 0 && genre.ID == genreDetailID {
				found = &genre
				break
			}
			if genreDetailName != "" && strings.EqualFold(genre.ShortName, genreDetailName) {
				found = &genre
				break
			}
			// Also check the key (which might be the short name)
			if genreDetailName != "" && strings.EqualFold(k, genreDetailName) {
				found = &genre
				break
			}
		}

		if found == nil {
			if genreDetailID != 0 {
				return fmt.Errorf("genre with ID %d not found", genreDetailID)
			}
			return fmt.Errorf("genre with name '%s' not found", genreDetailName)
		}

		if jsonOutput {
			formatted, err := json.MarshalIndent(found, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to format JSON: %w", err)
			}
			fmt.Println(string(formatted))
			return nil
		}

		lang := format.GetPreferredLanguage(locale)
		fmt.Print(format.RenderGenreDetail(*found, lang))
		return nil
	},
}

func init() {
	genreDetailCmd.Flags().IntVar(&genreDetailID, "id", 0, "Genre ID")
	genreDetailCmd.Flags().StringVar(&genreDetailName, "name", "", "Genre short name")
	genreDetailCmd.MarkFlagsOneRequired("id", "name")
	detailCmd.AddCommand(genreDetailCmd)
	listCmd.AddCommand(genresCmd)
}
