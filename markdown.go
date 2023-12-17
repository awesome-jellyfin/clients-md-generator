package generator

import (
	"fmt"
	"io"
	"strings"
)

const (
	OfficialTypeKey = "Official"
	BetaTypeKey     = "Beta"
)

// Markdown generates the markdown string for an icon.
func (i *HosterIcon) Markdown(url string) string {
	if (i.Dark != "") != (i.Light != "") {
		panic("use 'single' if only a single icon URL is available")
	}
	if i.Dark != "" {
		// Use picture element for alternate dark/light icons.
		return strings.TrimSpace(fmt.Sprintf(`<a href="%s">`+
			`<picture>`+
			`<source media="(prefers-color-scheme: dark)" srcset="%s">`+
			`<source media="(prefers-color-scheme: light)" srcset="%s">`+
			`<img src="%s">`+
			`</picture>`+
			`</a>`, url, i.Dark, i.Light, i.Dark))
	}
	if i.Text != "" {
		// Use Markdown link with text if text is provided.
		return fmt.Sprintf("[%s](%s)", i.Text, url)
	}
	// Use default single image icon if no text is given.
	return fmt.Sprintf("[![img](%s)](%s)", i.Single, url)
}

// processClientDownloads generates markdown for client downloads.
func processClientDownloads(client *Client, config *ClientsConfig) string {
	var sb strings.Builder

	for _, hoster := range client.Downloads {
		if sb.Len() > 0 {
			sb.WriteString(" ")
		}

		if icon, ok := config.Icons[hoster.Icon]; ok && hoster.Icon != "" {
			sb.WriteString(icon.Markdown(hoster.URL))
		} else if hoster.IconURL != "" {
			sb.WriteString((&HosterIcon{Single: hoster.IconURL}).Markdown(hoster.URL))
		} else if hoster.Text != "" {
			sb.WriteString(fmt.Sprintf("[%s](%s)", hoster.Text, hoster.URL))
		} else {
			panic("invalid download. specify either icon, icon-url, or text")
		}
	}

	return strings.ReplaceAll(sb.String(), "\n", "")
}

func PrintTableHeader(writer io.Writer) error {
	if _, err := fmt.Fprintln(writer, "| Name | OSS | Free | Paid | Downloads |"); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(writer, "| ---- | --- | ---- | ---- | --------- |"); err != nil {
		return err
	}
	return nil
}

func PrintClientTable(
	writer io.Writer,
	has string,
	identifierClientMap map[string][]*Client,
	config *ClientsConfig,
) error {
	if err := PrintTableHeader(writer); err != nil {
		return err
	}
	for _, client := range identifierClientMap[strings.ToLower(strings.TrimSpace(has))] {
		if err := PrintClientTableRow(writer, client, config); err != nil {
			return err
		}
	}
	return nil
}

// PrintClientTableRow prints a single row of the client table.
func PrintClientTableRow(writer io.Writer, client *Client, config *ClientsConfig) error {
	if client.Official == nil && strings.HasPrefix(client.OpenSourceURL, JellyfinOrgURL) {
		client.Official = Ref(true) // Default to official if part of Jellyfin organization
	}
	if client.Price.Free == nil && client.OpenSourceURL != "" {
		client.Price.Free = Ref(true) // Default to free if open-source
	}

	name := client.Name
	oss := Select(client.OpenSourceURL != "", GoodTrue, BadFalse)
	free := Select(DerefDef(client.Price.Free, false), GoodTrue, BadFalse)
	paid := Select(DerefDef(client.Price.Paid, false), BadTrue, GoodFalse)
	websiteURL := Select(client.Website != "", client.Website, client.OpenSourceURL)
	downloadsMarkdown := processClientDownloads(client, config)

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

	for _, b := range badges {
		name += fmt.Sprintf(" ` %s `", b)
	}

	if _, err := fmt.Fprintf(
		writer,
		"| [%s](%s) | %s | %s | %s | %s |",
		name,
		websiteURL,
		oss,
		free,
		paid,
		downloadsMarkdown,
	); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(writer); err != nil {
		return err
	}
	return nil
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

	if _, err := fmt.Fprint(writer, "# By Environment\n"); err != nil {
		return err
	}

	// Generate and print the markdown content
	for _, target := range config.Targets {
		if _, err := fmt.Fprintf(writer, "## %s\n\n", target.Display); err != nil {
			return err
		}
		hasMultipleTargets := len(target.Has) > 1
		for _, meta := range target.Has {
			if hasMultipleTargets {
				if _, err := fmt.Fprintf(writer, "### %s\n\n", meta.Mapped); err != nil {
					return err
				}
			}
			if err := PrintClientTable(writer, meta.Name, targetClientsMap, config); err != nil {
				return err
			}
			if _, err := fmt.Fprintln(writer); err != nil {
				return err
			}
		}
	}

	// Generate Type legend / sections
	if len(config.Types) > 0 {
		printHeader := true
		for _, customType := range config.Types {
			if !customType.Section {
				continue
			}
			if printHeader {
				printHeader = false

				if _, err := fmt.Fprint(writer, "\n---\n\n"); err != nil {
					return err
				}
				if _, err := fmt.Fprint(writer, "# By Type\n"); err != nil {
					return err
				}
			}
			// find all clients with this type
			printTypeHeader := true
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
				if printTypeHeader {
					printTypeHeader = false

					if _, err := fmt.Fprintf(writer, "\n## %s\n\n", customType.StringWithBadge()); err != nil {
						return err
					}

					if err := PrintTableHeader(writer); err != nil {
						return err
					}
				}
				if err := PrintClientTableRow(writer, client, config); err != nil {
					return err
				}
			}
		}

		if _, err := fmt.Fprint(writer, "\n---\n\n"); err != nil {
			return err
		}
		for _, customType := range config.Types {
			if customType.Badge == "" {
				continue
			}
			if _, err := fmt.Fprintf(writer, "* %s: ` %s `\n", customType.String(), customType.Badge); err != nil {
				return err
			}
		}
	}

	return nil
}
