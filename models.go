package generator

// Price indicates the cost of a client.
type Price struct {
	Free *bool `yaml:"free"`
	Paid *bool `yaml:"paid"`
}

// Hoster describes the hosting details for client downloads.
type Hoster struct {
	Icon    string `yaml:"icon"`
	IconURL string `yaml:"icon-url"`
	Text    string `yaml:"text"`
	URL     string `yaml:"url"`
}

// Client defines a client application for Jellyfin with its properties.
type Client struct {
	Name          string    `yaml:"name"`
	Targets       []string  `yaml:"targets"`
	Official      *bool     `yaml:"official"`
	Beta          *bool     `yaml:"beta"`
	Website       string    `yaml:"website"`
	OpenSourceURL string    `yaml:"oss"`
	Price         Price     `yaml:"price"`
	Downloads     []*Hoster `yaml:"downloads"`
	Types         []string  `yaml:"types"`
}

type Target struct {
	Name   string `json:"name,omitempty"`
	Mapped string `json:"mapped,omitempty"`
}

// TargetGroup defines a group of targets for the clients.
type TargetGroup struct {
	Key     string    `yaml:"key"`
	Display string    `yaml:"display"`
	Has     []*Target `yaml:"has"`
}

// ClientsConfig holds the configuration for all clients.
type ClientsConfig struct {
	Clients []*Client        `yaml:"clients"`
	Targets []*TargetGroup   `yaml:"targets"`
	Icons   map[string]*Icon `yaml:"icons"`
}
