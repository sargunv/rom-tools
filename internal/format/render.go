package format

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/sargunv/rom-tools/clients/screenscraper"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

// KVPair represents a key-value pair for rendering
type KVPair struct {
	Key   string
	Value string
}

// RenderTable renders a table with headers and rows
func RenderTable(headers []string, rows [][]string) string {
	if len(rows) == 0 {
		return ""
	}

	t := table.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(BorderStyle).
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == table.HeaderRow {
				return TableHeaderStyle
			}
			if row%2 == 0 {
				return TableEvenRowStyle
			}
			return TableOddRowStyle
		}).
		Headers(headers...).
		Rows(rows...)

	return t.Render()
}

// RenderKeyValue renders a list of key-value pairs
func RenderKeyValue(pairs []KVPair) string {
	if len(pairs) == 0 {
		return ""
	}

	var lines []string
	for _, pair := range pairs {
		if pair.Value == "" {
			continue
		}
		key := LabelStyle.Render(pair.Key + ":")
		lines = append(lines, fmt.Sprintf("%s %s", key, ValueStyle.Render(pair.Value)))
	}

	return strings.Join(lines, "\n")
}

// RenderID renders an ID in a dimmed style
func RenderID(id string) string {
	if id == "" || id == "0" {
		return ""
	}
	return DimStyle.Render("(" + id + ")")
}

// RenderSystemsList renders a list of systems
func RenderSystemsList(systems []screenscraper.System, lang string) string {
	if len(systems) == 0 {
		return "No systems found.\n"
	}

	// Sort by ID
	sort.Slice(systems, func(i, j int) bool {
		return systems[i].ID < systems[j].ID
	})

	headers := []string{"ID", "Name", "Company", "Type"}
	var rows [][]string

	for _, sys := range systems {
		name := GetLocalizedFromMap(lang, sys.Names)
		if name == "" {
			name = fmt.Sprintf("System %d", sys.ID)
		}

		rows = append(rows, []string{
			strconv.Itoa(sys.ID),
			name,
			sys.Company,
			sys.Type,
		})
	}

	return RenderTable(headers, rows) + "\n"
}

// RenderRegionsList renders a list of regions
func RenderRegionsList(regions map[string]screenscraper.Region, lang string) string {
	if len(regions) == 0 {
		return "No regions found.\n"
	}

	// Convert to slice and sort by ID
	type regionEntry struct {
		key    string
		region screenscraper.Region
	}
	var entries []regionEntry
	for k, r := range regions {
		entries = append(entries, regionEntry{k, r})
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].region.ID < entries[j].region.ID
	})

	headers := []string{"ID", "Short Name", "Name"}
	var rows [][]string

	for _, entry := range entries {
		r := entry.region
		name := GetLocalizedName(lang, r.NameDE, r.NameEN, r.NameES, r.NameFR, r.NameIT, r.NamePT)
		if name == "" {
			name = r.ShortName
		}

		rows = append(rows, []string{
			strconv.Itoa(r.ID),
			r.ShortName,
			name,
		})
	}

	return RenderTable(headers, rows) + "\n"
}

// RenderGenresList renders a list of genres
func RenderGenresList(genres map[string]screenscraper.Genre, lang string) string {
	if len(genres) == 0 {
		return "No genres found.\n"
	}

	// Convert to slice and sort by ID
	type genreEntry struct {
		key   string
		genre screenscraper.Genre
	}
	var entries []genreEntry
	for k, g := range genres {
		entries = append(entries, genreEntry{k, g})
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].genre.ID < entries[j].genre.ID
	})

	headers := []string{"ID", "Short Name", "Name"}
	var rows [][]string

	for _, entry := range entries {
		g := entry.genre
		name := GetLocalizedName(lang, g.NameDE, g.NameEN, g.NameES, g.NameFR, g.NameIT, g.NamePT)
		if name == "" {
			name = g.ShortName
		}

		rows = append(rows, []string{
			strconv.Itoa(g.ID),
			g.ShortName,
			name,
		})
	}

	return RenderTable(headers, rows) + "\n"
}

// RenderLanguagesList renders a list of languages
func RenderLanguagesList(languages map[string]screenscraper.Language, lang string) string {
	if len(languages) == 0 {
		return "No languages found.\n"
	}

	// Convert to slice and sort by ID
	type langEntry struct {
		key      string
		language screenscraper.Language
	}
	var entries []langEntry
	for k, l := range languages {
		entries = append(entries, langEntry{k, l})
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].language.ID < entries[j].language.ID
	})

	headers := []string{"ID", "Short Name", "Name"}
	var rows [][]string

	for _, entry := range entries {
		l := entry.language
		name := GetLocalizedName(lang, l.NameDE, l.NameEN, l.NameES, l.NameFR, l.NameIT, l.NamePT)
		if name == "" {
			name = l.ShortName
		}

		rows = append(rows, []string{
			strconv.Itoa(l.ID),
			l.ShortName,
			name,
		})
	}

	return RenderTable(headers, rows) + "\n"
}

// RenderFamiliesList renders a list of families
func RenderFamiliesList(families map[string]screenscraper.Family, lang string) string {
	if len(families) == 0 {
		return "No families found.\n"
	}

	// Convert to slice and sort by ID
	type familyEntry struct {
		key    string
		family screenscraper.Family
	}
	var entries []familyEntry
	for k, f := range families {
		entries = append(entries, familyEntry{k, f})
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].family.ID < entries[j].family.ID
	})

	headers := []string{"ID", "Name"}
	var rows [][]string

	for _, entry := range entries {
		f := entry.family
		rows = append(rows, []string{
			strconv.Itoa(f.ID),
			f.Name,
		})
	}

	return RenderTable(headers, rows) + "\n"
}

// RenderClassificationsList renders a list of classifications
func RenderClassificationsList(classifications map[string]screenscraper.Classification, lang string) string {
	if len(classifications) == 0 {
		return "No classifications found.\n"
	}

	// Convert to slice and sort by ID
	type classEntry struct {
		key            string
		classification screenscraper.Classification
	}
	var entries []classEntry
	for k, c := range classifications {
		entries = append(entries, classEntry{k, c})
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].classification.ID < entries[j].classification.ID
	})

	headers := []string{"ID", "Short Name", "Name"}
	var rows [][]string

	for _, entry := range entries {
		c := entry.classification
		name := GetLocalizedName(lang, c.NameDE, c.NameEN, c.NameES, c.NameFR, c.NameIT, c.NamePT)
		if name == "" {
			name = c.ShortName
		}

		rows = append(rows, []string{
			strconv.Itoa(c.ID),
			c.ShortName,
			name,
		})
	}

	return RenderTable(headers, rows) + "\n"
}

// boolCheck returns a checkmark if the value is "1", otherwise empty string
func boolCheck(value string) string {
	if value == "1" {
		return "✓"
	}
	return ""
}

// RenderMediaTypesList renders a list of media types
func RenderMediaTypesList(medias map[string]screenscraper.MediaType, lang string) string {
	if len(medias) == 0 {
		return "No media types found.\n"
	}

	// Convert to slice and sort by ID
	type mediaEntry struct {
		key   string
		media screenscraper.MediaType
	}
	var entries []mediaEntry
	for k, m := range medias {
		entries = append(entries, mediaEntry{k, m})
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].media.ID < entries[j].media.ID
	})

	headers := []string{"ID", "Short Name", "Name", "Format", "Region", "Support #", "Version"}
	var rows [][]string

	for _, entry := range entries {
		m := entry.media
		rows = append(rows, []string{
			strconv.Itoa(m.ID),
			m.ShortName,
			m.Name,
			m.FileFormat,
			boolCheck(m.MultiRegions),
			boolCheck(m.MultiSupports),
			boolCheck(m.MultiVersions),
		})
	}

	return RenderTable(headers, rows) + "\n"
}

// RenderROMTypesList renders a list of ROM types
func RenderROMTypesList(romTypes []string, lang string) string {
	if len(romTypes) == 0 {
		return "No ROM types found.\n"
	}

	sort.Strings(romTypes)

	headers := []string{"ROM Type"}
	var rows [][]string

	for _, rt := range romTypes {
		rows = append(rows, []string{rt})
	}

	return RenderTable(headers, rows) + "\n"
}

// RenderPlayerCountsList renders a list of player counts
func RenderPlayerCountsList(playerCounts map[string]screenscraper.PlayerCount, lang string) string {
	if len(playerCounts) == 0 {
		return "No player counts found.\n"
	}

	// Convert to slice and sort by ID
	type pcEntry struct {
		key         string
		playerCount screenscraper.PlayerCount
	}
	var entries []pcEntry
	for k, pc := range playerCounts {
		entries = append(entries, pcEntry{k, pc})
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].playerCount.ID < entries[j].playerCount.ID
	})

	headers := []string{"ID", "Name"}
	var rows [][]string

	for _, entry := range entries {
		pc := entry.playerCount
		rows = append(rows, []string{
			strconv.Itoa(pc.ID),
			pc.Name,
		})
	}

	return RenderTable(headers, rows) + "\n"
}

// RenderGame renders detailed game information
func RenderGame(game screenscraper.Game, lang string) string {
	var parts []string

	// Title
	title := GetNameFromNameEntries(game.Names, "")
	if title == "" {
		title = game.Name
	}
	if title != "" {
		parts = append(parts, TitleStyle.Render(title))
	}

	var kvPairs []KVPair

	// System
	if game.System.Text != "" {
		kvPairs = append(kvPairs, KVPair{"System", game.System.Text})
	}

	// Publisher
	if game.Publisher.Text != "" {
		kvPairs = append(kvPairs, KVPair{"Publisher", game.Publisher.Text})
	}

	// Developer
	if game.Developer.Text != "" {
		kvPairs = append(kvPairs, KVPair{"Developer", game.Developer.Text})
	}

	// Players
	if game.Players.Text != "" {
		kvPairs = append(kvPairs, KVPair{"Players", game.Players.Text})
	}

	// Release dates
	if len(game.Dates) > 0 {
		var dateStrs []string
		for _, date := range game.Dates {
			if date.Text != "" {
				region := ""
				if date.Region != "" {
					region = " (" + date.Region + ")"
				}
				dateStrs = append(dateStrs, date.Text+region)
			}
		}
		if len(dateStrs) > 0 {
			kvPairs = append(kvPairs, KVPair{"Release", strings.Join(dateStrs, ", ")})
		}
	}

	// Add key-value pairs
	if len(kvPairs) > 0 {
		parts = append(parts, RenderKeyValue(kvPairs))
	}

	// Synopsis
	if len(game.Synopsis) > 0 {
		synopsis := GetLocalizedFromSlice(lang, game.Synopsis)
		if synopsis != "" {
			parts = append(parts, "")
			parts = append(parts, HeaderStyle.Render("Synopsis:"))
			parts = append(parts, "  "+synopsis)
		}
	}

	// Genres
	if len(game.Genres) > 0 {
		var genreNames []string
		for _, genre := range game.Genres {
			name := GetLocalizedFromSlice(lang, genre.Names)
			if name == "" {
				name = genre.ShortName
			}
			if name != "" {
				genreNames = append(genreNames, name)
			}
		}
		if len(genreNames) > 0 {
			parts = append(parts, "")
			parts = append(parts, HeaderStyle.Render("Genres:"))
			parts = append(parts, "  "+strings.Join(genreNames, ", "))
		}
	}

	// Media URLs - group by type, show only main types
	if len(game.Medias) > 0 {
		// Main media types to show
		mainTypes := map[string]bool{
			"sstitle":         true,
			"ss":              true,
			"video":           true,
			"wheel":           true,
			"box-2D":          true,
			"box-2D-side":     true,
			"box-2D-back":     true,
			"manuel":          true,
			"support-texture": true,
		}

		// Group media by type, preserving order
		type mediaEntry struct {
			region string
			url    string
		}
		mediaByType := make(map[string][]mediaEntry)
		var typeOrder []string
		for _, media := range game.Medias {
			if media.URL == "" || !mainTypes[media.Type] {
				continue
			}
			if _, seen := mediaByType[media.Type]; !seen {
				typeOrder = append(typeOrder, media.Type)
			}
			region := media.Region
			if region == "" {
				region = "link"
			}
			mediaByType[media.Type] = append(mediaByType[media.Type], mediaEntry{region, media.URL})
		}

		if len(typeOrder) > 0 {
			parts = append(parts, "")
			parts = append(parts, HeaderStyle.Render("Media:"))
			for _, mediaType := range typeOrder {
				entries := mediaByType[mediaType]
				var links []string
				for _, entry := range entries {
					links = append(links, Hyperlink(entry.url, entry.region))
				}
				parts = append(parts, "  "+LabelStyle.Render(mediaType)+": "+strings.Join(links, " "))
			}
		}
	}

	return strings.Join(parts, "\n") + "\n"
}

// RenderGamesList renders a list of games (for search results)
func RenderGamesList(games []screenscraper.Game, lang string) string {
	if len(games) == 0 {
		return "No games found.\n"
	}

	headers := []string{"ID", "Name", "System", "Publisher"}
	var rows [][]string

	for _, game := range games {
		name := GetNameFromNameEntries(game.Names, "")
		if name == "" {
			name = game.Name
		}
		if name == "" {
			name = "Unknown"
		}

		rows = append(rows, []string{
			game.ID,
			name,
			game.System.Text,
			game.Publisher.Text,
		})
	}

	return RenderTable(headers, rows) + "\n"
}

// RenderUser renders user information
func RenderUser(user screenscraper.UserInfo, lang string) string {
	var parts []string

	// Account
	var accountPairs []KVPair
	if user.ID != "" {
		accountPairs = append(accountPairs, KVPair{"User", user.ID + " " + RenderID(user.NumID)})
	}
	if user.Level != "" {
		accountPairs = append(accountPairs, KVPair{"Level", user.Level})
	}
	if user.Contribution != "" {
		contribLevel := user.Contribution
		if contribLevel == "2" {
			contribLevel += " (1 Additional Thread)"
		} else if contribLevel != "0" && contribLevel != "1" {
			contribLevel += " (5 Additional Threads)"
		}
		accountPairs = append(accountPairs, KVPair{"Contribution", contribLevel})
	}
	if user.FavoriteRegion != "" {
		accountPairs = append(accountPairs, KVPair{"Favorite Region", user.FavoriteRegion})
	}
	if len(accountPairs) > 0 {
		parts = append(parts, HeaderStyle.Render("Account:"))
		parts = append(parts, RenderKeyValue(accountPairs))
	}

	// Limits
	var limitPairs []KVPair
	if user.MaxThreads != "" {
		limitPairs = append(limitPairs, KVPair{"Max Threads", user.MaxThreads})
	}
	if user.MaxDownloadSpeed != "" {
		limitPairs = append(limitPairs, KVPair{"Max Download Speed", user.MaxDownloadSpeed + " KB/s"})
	}
	if user.MaxRequestsPerMin != "" {
		limitPairs = append(limitPairs, KVPair{"Max Requests Per Minute", user.MaxRequestsPerMin})
	}
	if user.MaxRequestsPerDay != "" {
		limitPairs = append(limitPairs, KVPair{"Max Requests Per Day", user.MaxRequestsPerDay})
	}
	if user.MaxRequestsKOPerDay != "" {
		limitPairs = append(limitPairs, KVPair{"Max Failed Requests Per Day", user.MaxRequestsKOPerDay})
	}
	if len(limitPairs) > 0 {
		parts = append(parts, "")
		parts = append(parts, HeaderStyle.Render("Limits:"))
		parts = append(parts, RenderKeyValue(limitPairs))
	}

	// Usage
	var usagePairs []KVPair
	if user.RequestsToday != "" && user.MaxRequestsPerDay != "" {
		usagePairs = append(usagePairs, KVPair{
			"Requests Today",
			user.RequestsToday + " / " + user.MaxRequestsPerDay,
		})
	}
	if user.RequestsKOToday != "" && user.MaxRequestsKOPerDay != "" {
		usagePairs = append(usagePairs, KVPair{
			"Failed Requests Today",
			user.RequestsKOToday + " / " + user.MaxRequestsKOPerDay,
		})
	}
	if user.Visites != "" {
		usagePairs = append(usagePairs, KVPair{"Total Visits", user.Visites})
	}
	if user.LastVisitDate != "" {
		usagePairs = append(usagePairs, KVPair{"Last Visit", user.LastVisitDate})
	}
	if len(usagePairs) > 0 {
		parts = append(parts, "")
		parts = append(parts, HeaderStyle.Render("Usage:"))
		parts = append(parts, RenderKeyValue(usagePairs))
	}

	// Contributions
	var contribPairs []KVPair
	if user.UploadSystem != "" {
		contribPairs = append(contribPairs, KVPair{"System Media", user.UploadSystem})
	}
	if user.UploadInfos != "" {
		contribPairs = append(contribPairs, KVPair{"Text Info", user.UploadInfos})
	}
	if user.UploadMedia != "" {
		contribPairs = append(contribPairs, KVPair{"Game Media", user.UploadMedia})
	}
	if user.ROMAsso != "" {
		contribPairs = append(contribPairs, KVPair{"ROM Associations", user.ROMAsso})
	}
	if user.PropositionOK != "" {
		contribPairs = append(contribPairs, KVPair{"Proposals Accepted", user.PropositionOK})
	}
	if user.PropositionKO != "" {
		contribPairs = append(contribPairs, KVPair{"Proposals Rejected", user.PropositionKO})
	}
	if len(contribPairs) > 0 {
		parts = append(parts, "")
		parts = append(parts, HeaderStyle.Render("Contributions:"))
		parts = append(parts, RenderKeyValue(contribPairs))
	}

	return strings.Join(parts, "\n") + "\n"
}

// RenderInfra renders infrastructure/server information
func RenderInfra(servers screenscraper.ServerInfo, lang string) string {
	var kvPairs []KVPair

	if servers.APIAccess != "" {
		kvPairs = append(kvPairs, KVPair{"API Access", servers.APIAccess})
	}
	if servers.NbScrapeurs != "" {
		kvPairs = append(kvPairs, KVPair{"Scrapers", servers.NbScrapeurs})
	}
	if servers.CPU1 != "" {
		cpus := []string{servers.CPU1}
		if servers.CPU2 != "" {
			cpus = append(cpus, servers.CPU2)
		}
		if servers.CPU3 != "" {
			cpus = append(cpus, servers.CPU3)
		}
		if servers.CPU4 != "" {
			cpus = append(cpus, servers.CPU4)
		}
		kvPairs = append(kvPairs, KVPair{"CPUs", strings.Join(cpus, ", ")})
	}
	if servers.ThreadForMember != "" {
		kvPairs = append(kvPairs, KVPair{"Threads (Member)", servers.ThreadForMember + " / " + servers.MaxThreadForMember})
	}
	if servers.ThreadForNonMember != "" {
		kvPairs = append(kvPairs, KVPair{"Threads (Non-Member)", servers.ThreadForNonMember + " / " + servers.MaxThreadForNonMember})
	}

	return RenderKeyValue(kvPairs) + "\n"
}

// RenderROMInfoTypesList renders a list of ROM info types
func RenderROMInfoTypesList(infoTypes map[string]screenscraper.ROMInfoType, lang string) string {
	if len(infoTypes) == 0 {
		return "No ROM info types found.\n"
	}

	// Convert to slice and sort by ID
	type infoEntry struct {
		key  string
		info screenscraper.ROMInfoType
	}
	var entries []infoEntry
	for k, info := range infoTypes {
		entries = append(entries, infoEntry{k, info})
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].info.ID < entries[j].info.ID
	})

	headers := []string{"ID", "Short Name", "Name", "Type", "Region", "Version", "Multi"}
	var rows [][]string

	for _, entry := range entries {
		info := entry.info
		rows = append(rows, []string{
			strconv.Itoa(info.ID),
			info.ShortName,
			info.Name,
			info.Type,
			boolCheck(info.MultiRegions),
			boolCheck(info.MultiVersions),
			boolCheck(info.MultiChoix),
		})
	}

	return RenderTable(headers, rows) + "\n"
}

// RenderGameInfoTypesList renders a list of game info types
func RenderGameInfoTypesList(infoTypes map[string]screenscraper.GameInfoType, lang string) string {
	if len(infoTypes) == 0 {
		return "No game info types found.\n"
	}

	// Convert to slice and sort by ID
	type infoEntry struct {
		key  string
		info screenscraper.GameInfoType
	}
	var entries []infoEntry
	for k, info := range infoTypes {
		entries = append(entries, infoEntry{k, info})
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].info.ID < entries[j].info.ID
	})

	headers := []string{"ID", "Short Name", "Name", "Type", "Region", "Version", "Multi"}
	var rows [][]string

	for _, entry := range entries {
		info := entry.info
		rows = append(rows, []string{
			strconv.Itoa(info.ID),
			info.ShortName,
			info.Name,
			info.Type,
			boolCheck(info.MultiRegions),
			boolCheck(info.MultiVersions),
			boolCheck(info.MultiChoix),
		})
	}

	return RenderTable(headers, rows) + "\n"
}

// RenderSupportTypesList renders a list of support types
func RenderSupportTypesList(supportTypes []string, lang string) string {
	if len(supportTypes) == 0 {
		return "No support types found.\n"
	}

	sort.Strings(supportTypes)

	headers := []string{"Support Type"}
	var rows [][]string

	for _, st := range supportTypes {
		rows = append(rows, []string{st})
	}

	return RenderTable(headers, rows) + "\n"
}

// RenderUserLevelsList renders a list of user levels
func RenderUserLevelsList(userLevels map[string]screenscraper.UserLevel, lang string) string {
	if len(userLevels) == 0 {
		return "No user levels found.\n"
	}

	// Convert to slice and sort by ID
	type ulEntry struct {
		key       string
		userLevel screenscraper.UserLevel
	}
	var entries []ulEntry
	for k, ul := range userLevels {
		entries = append(entries, ulEntry{k, ul})
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].userLevel.ID < entries[j].userLevel.ID
	})

	headers := []string{"ID", "Name"}
	var rows [][]string

	for _, entry := range entries {
		ul := entry.userLevel
		name := ul.NameFR
		if name == "" {
			name = "Level " + strconv.Itoa(ul.ID)
		}
		rows = append(rows, []string{
			strconv.Itoa(ul.ID),
			name,
		})
	}

	return RenderTable(headers, rows) + "\n"
}

// RenderSystemDetail renders detailed system information
func RenderSystemDetail(system screenscraper.System, lang string) string {
	var parts []string

	// Title
	name := GetLocalizedFromMap(lang, system.Names)
	if name == "" {
		name = fmt.Sprintf("System %d", system.ID)
	}
	parts = append(parts, TitleStyle.Render(fmt.Sprintf("%s %s", name, RenderID(strconv.Itoa(system.ID)))))

	var kvPairs []KVPair

	// Basic info
	kvPairs = append(kvPairs, KVPair{"ID", strconv.Itoa(system.ID)})
	if system.ParentID != 0 {
		kvPairs = append(kvPairs, KVPair{"Parent ID", strconv.Itoa(system.ParentID)})
	}
	if system.Company != "" {
		kvPairs = append(kvPairs, KVPair{"Company", system.Company})
	}
	if system.Type != "" {
		kvPairs = append(kvPairs, KVPair{"Type", system.Type})
	}
	if system.StartDate != "" {
		kvPairs = append(kvPairs, KVPair{"Start Date", system.StartDate})
	}
	if system.EndDate != "" {
		kvPairs = append(kvPairs, KVPair{"End Date", system.EndDate})
	}
	if system.Extensions != "" {
		kvPairs = append(kvPairs, KVPair{"Extensions", system.Extensions})
	}
	if system.ROMType != "" {
		kvPairs = append(kvPairs, KVPair{"ROM Type", system.ROMType})
	}
	if system.SupportType != "" {
		kvPairs = append(kvPairs, KVPair{"Support Type", system.SupportType})
	}

	if len(kvPairs) > 0 {
		parts = append(parts, RenderKeyValue(kvPairs))
	}

	// Names section
	if len(system.Names) > 0 {
		parts = append(parts, "")
		parts = append(parts, HeaderStyle.Render("Names:"))
		var nameLines []string
		for key, value := range system.Names {
			if value != "" {
				nameLines = append(nameLines, fmt.Sprintf("  %s: %s", key, value))
			}
		}
		sort.Strings(nameLines)
		parts = append(parts, strings.Join(nameLines, "\n"))
	}

	// Media section
	if len(system.Medias) > 0 {
		// Main media types to show for systems
		mainTypes := map[string]bool{
			"photo":        true,
			"illustration": true,
			"controller":   true,
			"wheel":        true,
			"video":        true,
		}

		// Group media by type, preserving order
		type mediaEntry struct {
			region string
			url    string
		}
		mediaByType := make(map[string][]mediaEntry)
		var typeOrder []string
		for _, media := range system.Medias {
			if media.URL == "" || !mainTypes[media.Type] {
				continue
			}
			if _, seen := mediaByType[media.Type]; !seen {
				typeOrder = append(typeOrder, media.Type)
			}
			region := media.Region
			if region == "" {
				region = "link"
			}
			mediaByType[media.Type] = append(mediaByType[media.Type], mediaEntry{region, media.URL})
		}

		if len(typeOrder) > 0 {
			parts = append(parts, "")
			parts = append(parts, HeaderStyle.Render("Media:"))
			for _, mediaType := range typeOrder {
				entries := mediaByType[mediaType]
				var links []string
				for _, entry := range entries {
					links = append(links, Hyperlink(entry.url, entry.region))
				}
				parts = append(parts, "  "+LabelStyle.Render(mediaType)+": "+strings.Join(links, " "))
			}
		}
	}

	return strings.Join(parts, "\n") + "\n"
}
