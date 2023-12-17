package generator

import (
	"gopkg.in/yaml.v3"
	"os"
	"strings"
)

const (
	// JellyfinOrgURL is the Jellyfin GitHub organization URL.
	JellyfinOrgURL = "https://github.com/jellyfin/"
	GoodTrue       = "✅"
	BadTrue        = "☑️"
	GoodFalse      = "❎"
	BadFalse       = "❌"
)

// LoadConfig reads and unmarshals the YAML config file.
func LoadConfig(filename string) (config *ClientsConfig, err error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(data, &config)
	return
}

// createIdentifierClientMap creates a map of identifiers to corresponding clients.
func createIdentifierClientMap(clients []*Client) map[string][]*Client {
	identifierClientMap := make(map[string][]*Client)

	for _, client := range clients {
		for _, targetStr := range client.Targets {
			targetStr = strings.TrimSpace(strings.ToLower(targetStr))
			identifierClientMap[targetStr] = append(identifierClientMap[targetStr], client)
		}
	}
	return identifierClientMap
}
