package screenscraper

// Header contains common API response metadata
type Header struct {
	APIVersion       string `json:"APIversion"`
	CommandRequested string `json:"commandRequested"`
	DateTime         string `json:"dateTime"`
	Error            string `json:"error"`
	Success          string `json:"success"`
}

// ServerInfo contains server infrastructure information (included in most responses)
type ServerInfo struct {
	APIAccess             string `json:"apiacces"`
	CloseForLeecher       string `json:"closeforleecher"`
	CloseForNonMember     string `json:"closefornomember"`
	CPU1                  string `json:"cpu1"`
	CPU2                  string `json:"cpu2"`
	CPU3                  string `json:"cpu3"`
	CPU4                  string `json:"cpu4"`
	MaxThreadForMember    string `json:"maxthreadformember"`
	MaxThreadForNonMember string `json:"maxthreadfornonmember"`
	NbScrapeurs           string `json:"nbscrapeurs"`
	ThreadForMember       string `json:"threadformember"`
	ThreadForNonMember    string `json:"threadfornonmember"`
	ThreadsMin            string `json:"threadsmin"`
}

// UserInfo contains user quota and contribution information
type UserInfo struct {
	ID                  string `json:"id"`
	NumID               string `json:"numid"`
	Level               string `json:"niveau"`
	Contribution        string `json:"contribution"`
	UploadSystem        string `json:"uploadsysteme"`
	UploadInfos         string `json:"uploadinfos"`
	ROMAsso             string `json:"romasso"`
	UploadMedia         string `json:"uploadmedia"`
	PropositionOK       string `json:"propositionok"`
	PropositionKO       string `json:"propositionko"`
	QuotaRefu           string `json:"quotarefu"`
	MaxThreads          string `json:"maxthreads"`
	MaxDownloadSpeed    string `json:"maxdownloadspeed"`
	RequestsToday       string `json:"requeststoday"`
	RequestsKOToday     string `json:"requestskotoday"`
	MaxRequestsPerMin   string `json:"maxrequestspermin"`
	MaxRequestsPerDay   string `json:"maxrequestsperday"`
	MaxRequestsKOPerDay string `json:"maxrequestskoperday"`
	Visites             string `json:"visites"`
	LastVisitDate       string `json:"datedernierevisite"`
	FavoriteRegion      string `json:"favregion"`
}

// Media is a common media descriptor used across multiple endpoints
type Media struct {
	CRC     string `json:"crc,omitempty"`
	MD5     string `json:"md5,omitempty"`
	SHA1    string `json:"sha1,omitempty"`
	Format  string `json:"format,omitempty"`
	Parent  string `json:"parent,omitempty"`
	Region  string `json:"region,omitempty"`
	Size    string `json:"size,omitempty"`
	Support string `json:"support,omitempty"`
	Type    string `json:"type"`
	URL     string `json:"url,omitempty"`
}

// LocalizedName represents a name in a specific language
type LocalizedName struct {
	Language string `json:"langue"`
	Text     string `json:"text"`
}

// NameEntry represents a name entry with region and text
type NameEntry struct {
	Region string `json:"region"`
	Text   string `json:"text"`
}
