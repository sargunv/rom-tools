package format

import (
	screenscraper "github.com/sargunv/screenscraper-go/client"
)

// GetLocalizedName extracts the localized name from individual language fields.
// The API provides fields for: DE, EN, ES, FR, IT, PT.
// Tries the preferred language, falls back to English, then any available.
func GetLocalizedName(lang string, nameDE, nameEN, nameES, nameFR, nameIT, namePT string) string {
	// Map of language codes to their corresponding fields
	langMap := map[string]string{
		"de": nameDE,
		"en": nameEN,
		"es": nameES,
		"fr": nameFR,
		"it": nameIT,
		"pt": namePT,
	}

	// 1. Try preferred language
	if name, ok := langMap[lang]; ok && name != "" {
		return name
	}

	// 2. Fall back to English
	if nameEN != "" {
		return nameEN
	}

	// 3. Fall back to any available language
	if nameFR != "" {
		return nameFR
	}
	if nameDE != "" {
		return nameDE
	}
	if nameES != "" {
		return nameES
	}
	if nameIT != "" {
		return nameIT
	}
	if namePT != "" {
		return namePT
	}

	return ""
}

// GetLocalizedFromMap extracts localized name from a map[string]string.
// The map keys are language codes (e.g., "en", "fr", "de").
// Tries the preferred language, falls back to English, then any available.
func GetLocalizedFromMap(lang string, names map[string]string) string {
	if names == nil {
		return ""
	}

	// 1. Try preferred language
	if name, ok := names[lang]; ok && name != "" {
		return name
	}

	// 2. Fall back to English
	if name, ok := names["en"]; ok && name != "" {
		return name
	}

	// 3. Fall back to any available language
	for _, name := range names {
		if name != "" {
			return name
		}
	}

	return ""
}

// GetLocalizedFromSlice extracts localized name from a slice of LocalizedName.
// Tries the preferred language, falls back to English, then any available.
func GetLocalizedFromSlice(lang string, names []screenscraper.LocalizedName) string {
	if len(names) == 0 {
		return ""
	}

	// 1. Try preferred language
	for _, n := range names {
		if n.Language == lang && n.Text != "" {
			return n.Text
		}
	}

	// 2. Fall back to English
	for _, n := range names {
		if n.Language == "en" && n.Text != "" {
			return n.Text
		}
	}

	// 3. Fall back to any available language
	for _, n := range names {
		if n.Text != "" {
			return n.Text
		}
	}

	return ""
}

// GetNameFromNameEntries extracts a name from NameEntry slice.
// Returns the first available name, optionally filtered by region.
func GetNameFromNameEntries(entries []screenscraper.NameEntry, preferredRegion string) string {
	if len(entries) == 0 {
		return ""
	}

	// If preferred region specified, try to find it
	if preferredRegion != "" {
		for _, entry := range entries {
			if entry.Region == preferredRegion && entry.Text != "" {
				return entry.Text
			}
		}
	}

	// Return first non-empty entry
	for _, entry := range entries {
		if entry.Text != "" {
			return entry.Text
		}
	}

	return ""
}
