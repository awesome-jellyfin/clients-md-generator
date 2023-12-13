package generator

import (
	"fmt"
	"io"
	"strings"
)

// Icon represents configuration for icons that can be used in markdown output.
type Icon struct {
	Light  string `yaml:"light"`
	Dark   string `yaml:"dark"`
	Single string `yaml:"single"`
	Text   string `yaml:"text"`
}

// Markdown generates the markdown string for an icon.
func (i *Icon) Markdown(url string) string {
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
			sb.WriteString((&Icon{Single: hoster.IconURL}).Markdown(hoster.URL))
		} else if hoster.Text != "" {
			sb.WriteString(fmt.Sprintf("[%s](%s)", hoster.Text, hoster.URL))
		} else {
			panic("invalid download. specify either icon, icon-url, or text")
		}
	}

	return strings.ReplaceAll(sb.String(), "\n", "")
}

func PrintClientTable(
	writer io.Writer,
	has string,
	identifierClientMap map[string][]*Client,
	config *ClientsConfig,
) error {
	if _, err := fmt.Fprintln(writer, "| Name | OSS | Free | Paid | Downloads |"); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(writer, "| ---- | --- | ---- | ---- | --------- |"); err != nil {
		return err
	}
	for _, client := range identifierClientMap[strings.ToLower(strings.TrimSpace(has))] {
		if err := PrintClientTableRow(writer, client, config); err != nil {
			return err
		}
		if _, err := fmt.Fprintln(writer); err != nil {
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

	// Set client details
	name := client.Name
	oss := Select(client.OpenSourceURL != "", GoodTrue, BadFalse)
	free := Select(DerefDef(client.Price.Free, false), GoodTrue, BadFalse)
	paid := Select(DerefDef(client.Price.Paid, false), BadTrue, GoodFalse)

	// Append badges
	if Deref(client.Official) {
		name += " ` " + BadgeOfficial + " `"
	}
	if Deref(client.Beta) {
		name += " ` " + BadgeBeta + " `"
	}
	for _, t := range client.Types {
		if t == "Music" {
			name += " ` " + ClientTypeMusic + " `"
		}
	}

	// Determine website URL
	websiteURL := Select(client.Website != "", client.Website, client.OpenSourceURL)

	downloadsMarkdown := processClientDownloads(client, config)

	_, err := fmt.Fprintf(writer, "| [%s](%s) | %s | %s | %s | %s |", name, websiteURL, oss, free, paid, downloadsMarkdown)
	return err
}

func CreateMarkdownDocument(writer io.Writer, config *ClientsConfig) error {
	// Process clients and create an identifier-client map
	// e.g. iOS: [Swiftfin, Infuse, ...]
	targetClientsMap := createIdentifierClientMap(config.Clients)

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

	return nil
}
