package generator

import (
	"io"
	"sort"
	"strings"
	"sync"
)

const (
	OfficialTypeKey = "Official"
	BetaTypeKey     = "Beta"
)

func PrintClientTable(
	clients []*Client,
	config *ClientsConfig,
) (MarkdownRenderer, error) {
	builder := NewTableBuilder(
		NewText("Name"),
		NewText("OSS"),
		NewText("Free"),
		NewText("Paid"),
		NewText("Downloads"),
	)

	for _, client := range clients {
		if client.Official == nil && strings.HasPrefix(client.OpenSourceURL, JellyfinOrgURL) {
			client.Official = Ref(true) // Default to official if part of Jellyfin organization
		}
		if client.Price.Free == nil && client.OpenSourceURL != "" {
			client.Price.Free = Ref(true) // Default to free if open-source
		}

		oss := Select[string](client.OpenSourceURL != "", GoodTrue, BadFalse)
		free := Select[string](DerefDef(client.Price.Free, false), GoodTrue, BadFalse)
		paid := Select[string](DerefDef(client.Price.Paid, false), BadTrue, GoodFalse)
		websiteURL := Select[string](client.Website != "", client.Website, client.OpenSourceURL)

		var badges []string
		if Deref(client.Official) {
			addTypeBadge(&badges, OfficialTypeKey, config)
		}
		if Deref(client.Beta) {
			addTypeBadge(&badges, BetaTypeKey, config)
		}
		for _, t := range client.Types {
			addTypeBadge(&badges, t, config)
		}

		nameWithBadges := NewHorizontal(NewText(client.Name))
		for _, b := range badges {
			nameWithBadges.Append(NewCode(b, CodePadded))
		}

		downloads := NewHorizontal()
		for _, download := range client.Downloads {
			downloads.Append(download.Render())
		}

		var nameElement MarkdownRenderer
		if websiteURL != "" {
			nameElement = &Link{
				Text: nameWithBadges,
				URL:  websiteURL,
			}
		} else {
			nameElement = nameWithBadges
		}

		builder.AddRow(
			nameElement,
			NewText(oss),
			NewText(free),
			NewText(paid),
			downloads,
		)
	}
	return builder.Build(), nil
}

func addTypeBadge(badges *[]string, key string, config *ClientsConfig) {
	// find beta type
	t, ok := config.Types.FindType(key)
	if !ok {
		panic("cannot find type with key: " + key)
	}
	if t.Badge != "" {
		*badges = append(*badges, t.Badge)
	}
}

func CreateMarkdownDocument(writer io.Writer, config *ClientsConfig) error {
	// Process clients and create an identifier-client map
	// e.g. iOS: [Swiftfin, Infuse, ...]
	targetClientsMap := createIdentifierClientMap(config.Clients)

	vertical := NewVertical()
	vertical.Append(NewHeading(1, NewText("By Environment")))

	// Generate and print the markdown content
	for _, target := range config.Targets {
		vertical.Append(NewHeading(2, NewText(target.Display)))

		hasMultipleTargets := len(target.Has) > 1
		for _, meta := range target.Has {
			if hasMultipleTargets {
				vertical.Append(NewHeading(3, NewText(meta.Mapped)))
			}

			clients := targetClientsMap[strings.ToLower(strings.TrimSpace(meta.Name))]
			SortClientsByName(clients)

			table, err := PrintClientTable(clients, config)
			if err != nil {
				return err
			}
			vertical.Append(table)
		}
	}

	// Generate Type legend / sections
	if len(config.Types) > 0 {
		var printHeader sync.Once
		for _, customType := range config.Types {
			if !customType.Section {
				continue
			}

			printHeader.Do(func() {
				vertical.Append(HorizontalDivider{})
				vertical.Append(NewHeading(1, NewText("By Type")))
			})

			var clients []*Client

			// find all clients with this type
			for _, client := range config.Clients {
				// check if client belongs to type
				belongs := false
				for _, clientType := range client.Types {
					if clientType == customType.Key {
						belongs = true
						break
					}
				}
				if !belongs {
					continue
				}

				clients = append(clients, client)
			}

			SortClientsByName(clients)
			table, err := PrintClientTable(clients, config)
			if err != nil {
				return err
			}

			vertical.Append(NewHeading(2, NewText(customType.StringWithBadge())))
			vertical.Append(table)
		}

		vertical.Append(HorizontalDivider{})

		badgeList := NewList(false)
		for _, customType := range config.Types {
			if customType.Badge == "" {
				continue
			}
			badgeList.Append(NewHorizontal(
				NewText(customType.String(), Bold),
				NewCode(customType.Badge, CodePadded),
			))
		}
		vertical.Append(badgeList)
	}

	_, err := writer.Write([]byte(vertical.Render()))
	return err
}

// SortClientsByName sorts the clients by name.
func SortClientsByName(clients []*Client) {
	sort.Slice(clients, func(i, j int) bool {
		return strings.ToLower(clients[i].Name) < strings.ToLower(clients[j].Name)
	})
}
