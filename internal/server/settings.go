package server

// AppSettings holds runtime configuration managed via the UI.
type AppSettings struct {
	SkipLinkHosts []string `json:"skip_link_hosts"`
}

// DefaultSkipLinkHosts are platforms known to block automated requests.
var DefaultSkipLinkHosts = []string{
	"linkedin.com",
	"www.linkedin.com",
	"twitter.com",
	"www.twitter.com",
	"x.com",
	"www.x.com",
	"instagram.com",
	"www.instagram.com",
	"facebook.com",
	"www.facebook.com",
	"tiktok.com",
	"www.tiktok.com",
}
