package generator

import "fmt"

// Price indicates the cost of a client.
type Price struct {
	Free *bool `yaml:"free"`
	Paid *bool `yaml:"paid"`
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
	Downloads     Downloads `yaml:"downloads"`
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

// ClientType represents a client type, such as music or reader clients
type ClientType struct {
	Key     string `json:"key"`
	Badge   string `json:"badge"`
	Display string `json:"display"`
	Section bool   `json:"section"`
}

func (t ClientType) String() string {
	if t.Display != "" {
		return t.Display
	}
	return t.Key
}

func (t ClientType) StringWithBadge() string {
	if t.Badge == "" {
		return t.String()
	}
	return fmt.Sprintf("` %s ` %s", t.Badge, t.String())
}

type ClientTypes []*ClientType

// ClientsConfig holds the configuration for all clients.
type ClientsConfig struct {
	Clients []*Client      `yaml:"clients"`
	Targets []*TargetGroup `yaml:"targets"`
	Types   ClientTypes    `yaml:"types"`
}

func (t ClientTypes) FindType(key string) (*ClientType, bool) {
	for _, ct := range t {
		if ct.Key == key {
			return ct, true
		}
	}
	return nil, false
}
