package main

import (
	"fmt"
	"log"
	"os"

	"github.com/sargunv/rom-tools/internal/cli"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

// disableAutoGenTag recursively disables the auto-generated tag on all commands
func disableAutoGenTag(cmd *cobra.Command) {
	cmd.DisableAutoGenTag = true
	for _, c := range cmd.Commands() {
		disableAutoGenTag(c)
	}
}

func main() {
	// Create output directory if it doesn't exist
	docsDir := "./docs"
	if err := os.MkdirAll(docsDir, 0755); err != nil {
		log.Fatalf("Failed to create docs directory: %v", err)
	}

	fmt.Printf("Generating markdown docs to %s...\n", docsDir)

	// Get the root command without credential checks for doc generation
	cmd := cli.GetRootCommandForDocs()

	// Disable the auto-generated tag with date to avoid unnecessary changes
	// Need to disable it on all commands, not just the root
	disableAutoGenTag(cmd)

	// Generate markdown documentation
	err := doc.GenMarkdownTree(cmd, docsDir)
	if err != nil {
		log.Fatalf("Failed to generate documentation: %v", err)
	}

	fmt.Println("Documentation generated successfully!")
}
