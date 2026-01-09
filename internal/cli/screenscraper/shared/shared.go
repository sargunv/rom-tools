package shared

import (
	"github.com/sargunv/rom-tools/clients/screenscraper"
)

// Shared state for screenscraper CLI commands
var (
	JsonOutput bool
	Locale     string
	Client     *screenscraper.Client
)
