package screenscraper

// DownloadMediaParams parameters for media download
type DownloadMediaParams struct {
	// Hash of existing local file (for deduplication)
	CRC  string
	MD5  string
	SHA1 string

	// Required
	SystemID    string
	GameID      string
	Media       string // media identifier like "box-2D(us)", "wheel-hd(eu)", etc.
	MediaFormat string // optional: jpg, png, etc.

	// Output options
	MaxWidth     string
	MaxHeight    string
	OutputFormat string // "png" or "jpg"
}

// DownloadGameMedia downloads game image media
func (c *Client) DownloadGameMedia(params DownloadMediaParams) ([]byte, error) {
	p := map[string]string{
		"crc":          params.CRC,
		"md5":          params.MD5,
		"sha1":         params.SHA1,
		"systemeid":    params.SystemID,
		"jeuid":        params.GameID,
		"media":        params.Media,
		"mediaformat":  params.MediaFormat,
		"maxwidth":     params.MaxWidth,
		"maxheight":    params.MaxHeight,
		"outputformat": params.OutputFormat,
	}
	return c.get("mediaJeu.php", p)
}

// DownloadSystemMedia downloads system image media
func (c *Client) DownloadSystemMedia(params DownloadMediaParams) ([]byte, error) {
	p := map[string]string{
		"crc":          params.CRC,
		"md5":          params.MD5,
		"sha1":         params.SHA1,
		"systemeid":    params.SystemID,
		"media":        params.Media,
		"mediaformat":  params.MediaFormat,
		"maxwidth":     params.MaxWidth,
		"maxheight":    params.MaxHeight,
		"outputformat": params.OutputFormat,
	}
	return c.get("mediaSysteme.php", p)
}

// DownloadGroupMediaParams parameters for group media download
type DownloadGroupMediaParams struct {
	// Hash of existing local file (for deduplication)
	CRC  string
	MD5  string
	SHA1 string

	// Required
	GroupID     string // Numeric identifier of the group (genre, mode, famille, theme, style)
	Media       string // media identifier like "logo-monochrome", "picto-couleur", etc.
	MediaFormat string // optional: jpg, png, etc.

	// Output options
	MaxWidth     string
	MaxHeight    string
	OutputFormat string // "png" or "jpg"
}

// DownloadGroupMedia downloads group image media (genres, modes, families, themes, styles)
func (c *Client) DownloadGroupMedia(params DownloadGroupMediaParams) ([]byte, error) {
	p := map[string]string{
		"crc":          params.CRC,
		"md5":          params.MD5,
		"sha1":         params.SHA1,
		"groupid":      params.GroupID,
		"media":        params.Media,
		"mediaformat":  params.MediaFormat,
		"maxwidth":     params.MaxWidth,
		"maxheight":    params.MaxHeight,
		"outputformat": params.OutputFormat,
	}
	return c.get("mediaGroup.php", p)
}

// DownloadCompanyMediaParams parameters for company media download
type DownloadCompanyMediaParams struct {
	// Hash of existing local file (for deduplication)
	CRC  string
	MD5  string
	SHA1 string

	// Required
	CompanyID   string // Numeric identifier of the company
	Media       string // media identifier like "logo-monochrome", "logo-couleur", etc.
	MediaFormat string // optional: jpg, png, etc.

	// Output options
	MaxWidth     string
	MaxHeight    string
	OutputFormat string // "png" or "jpg"
}

// DownloadCompanyMedia downloads company image media (publishers, developers)
func (c *Client) DownloadCompanyMedia(params DownloadCompanyMediaParams) ([]byte, error) {
	p := map[string]string{
		"crc":          params.CRC,
		"md5":          params.MD5,
		"sha1":         params.SHA1,
		"companyid":    params.CompanyID,
		"media":        params.Media,
		"mediaformat":  params.MediaFormat,
		"maxwidth":     params.MaxWidth,
		"maxheight":    params.MaxHeight,
		"outputformat": params.OutputFormat,
	}
	return c.get("mediaCompagnie.php", p)
}
