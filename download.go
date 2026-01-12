package generator

import (
	"fmt"
	"net/url"

	"gopkg.in/yaml.v3"
)

var downloadFactories = map[string]func() Download{
	// simple renderers
	"icon": func() Download { return &IconDownload{} },
	"text": func() Download { return &TextDownload{} },
	// dynamic renderers
	"github":  func() Download { return &GitHubDownload{} },
	"flathub": func() Download { return &FlathubDownload{} },
	"docker":  func() Download { return &DockerDownload{} },
	// other renderers
	"shield":      func() Download { return &CustomShieldDownload{} },
	"gitlab":      func() Download { return &GitLabDownload{} },
	"demo":        func() Download { return &DemoDownload{} },
	"app-store":   func() Download { return &AppStoreDownload{} },
	"google-play": func() Download { return &GooglePlayDownload{} },
}

type Download interface {
	Render() MarkdownRenderer
}

type Downloads []Download

func (ds *Downloads) UnmarshalYAML(value *yaml.Node) error {
	var rawItems []map[string]any
	if err := value.Decode(&rawItems); err != nil {
		return err
	}

	var result []Download
	for _, raw := range rawItems {
		rawType := ""

		if val, ok := raw["type"]; ok {
			if rawType, ok = val.(string); !ok {
				return fmt.Errorf("missing or invalid 'type' in download item: %v", raw)
			}
		}

		factory, exists := downloadFactories[rawType]
		if !exists {
			return fmt.Errorf("unknown download type: %s", rawType)
		}

		// this is a hack to convert the map to YAML and back to get the correct type
		data, err := yaml.Marshal(raw)
		if err != nil {
			return err
		}

		d := factory()
		if err := yaml.Unmarshal(data, d); err != nil {
			return err
		}
		result = append(result, d)
	}

	*ds = result
	return nil
}

func isEmpty(val any) bool {
	switch t := val.(type) {
	case string:
		return t == ""
	case *string:
		return t == nil || *t == ""
	}
	panic(fmt.Sprintf("isEmpty: unsupported type: %T", val))
}

func preconditions(name string, values map[string]any) {
	for k, v := range values {
		if isEmpty(v) {
			panic(fmt.Sprintf("%s is required for %s download", k, name))
		}
	}
}

func first[T any](vals ...T) T {
	for _, v := range vals {
		if !isEmpty(v) {
			return v
		}
	}
	// return the last value if all are empty
	return vals[len(vals)-1]
}

// GitHubDownload returns a Markdown link to shield.io for the GitHub repository.
type GitHubDownload struct {
	Owner string
	Repo  string
	URL   string
	Label string // override the label on the shield
}

func (g *GitHubDownload) Render() MarkdownRenderer {
	preconditions("GitHub", map[string]any{
		"Owner": g.Owner,
		"Repo":  g.Repo,
	})

	// use the URL if provided, otherwise generate it
	u := first(g.URL, fmt.Sprintf("https://github.com/%s/%s/releases", g.Owner, g.Repo))
	label := "GitHub"
	if g.Label != "" {
		label = g.Label
	}

	return &Link{
		Text: &Image{
			AltText: NewText("github"),
			ImageURL: fmt.Sprintf(
				"https://img.shields.io/github/downloads/%s/%s/total?logo=github&label=%s",
				url.PathEscape(g.Owner), url.PathEscape(g.Repo), url.QueryEscape(label)),
		},
		URL: u,
	}
}

// IconDownload represents a download link with an icon.
type IconDownload struct {
	Icon string
	URL  string
}

func (i *IconDownload) Render() MarkdownRenderer {
	preconditions("Icon", map[string]any{
		"Icon": i.Icon,
		"URL":  i.URL,
	})
	return &Link{
		Text: &Image{
			AltText:  NewText(i.Icon),
			ImageURL: fmt.Sprintf("assets/clients/icons/%s.png", url.PathEscape(i.Icon)),
		},
		URL: i.URL,
	}
}

// TextDownload represents a download link with text.
type TextDownload struct {
	Text string
	URL  string
}

func (t *TextDownload) Render() MarkdownRenderer {
	preconditions("Text", map[string]any{
		"Text": t.Text,
		"URL":  t.URL,
	})
	return &Link{
		Text: NewText(t.Text),
		URL:  t.URL,
	}
}

// FlathubDownload represents a download link to Flathub.
type FlathubDownload struct {
	Package string
	URL     string
}

func (f *FlathubDownload) Render() MarkdownRenderer {
	preconditions("Flathub", map[string]any{
		"Package": f.Package,
	})

	// use the URL if provided, otherwise generate it
	u := first(f.URL, fmt.Sprintf("https://flathub.org/apps/%s", f.Package))

	return &Link{
		Text: &Image{
			AltText: NewText("flathub"),
			ImageURL: fmt.Sprintf(
				"https://img.shields.io/flathub/downloads/%s?logo=Flathub&label=Flathub",
				url.PathEscape(f.Package)),
		},
		URL: u,
	}
}

// DockerDownload represents a download link to Docker Hub.
type DockerDownload struct {
	User string
	Repo string
	URL  string
}

func (d *DockerDownload) Render() MarkdownRenderer {
	preconditions("Docker", map[string]any{
		"User": d.User,
		"Repo": d.Repo,
	})

	// use the URL if provided, otherwise generate it
	u := first(d.URL, fmt.Sprintf("https://hub.docker.com/r/%s/%s", d.User, d.Repo))

	return &Link{
		Text: &Image{
			AltText: NewText("docker"),
			ImageURL: fmt.Sprintf(
				"https://img.shields.io/docker/pulls/%s/%s?logo=docker&label=Docker",
				url.PathEscape(d.User), url.PathEscape(d.Repo)),
		},
		URL: u,
	}
}

// CustomShieldDownload represents a download link with a custom shield.
type CustomShieldDownload struct {
	Label   string
	Content *string
	Icon    string
	Color   string
	URL     string
}

func (c *CustomShieldDownload) Render() MarkdownRenderer {
	preconditions("CustomShield", map[string]any{
		"URL": c.URL,
	})

	color := first(c.Color, "grey")
	alt := first(c.Label, c.Icon, "alt")

	var content string
	if c.Content != nil {
		content = *c.Content
	} else if c.Icon != "" {
		content = c.Icon
	}

	return &Link{
		Text: &Image{
			AltText: NewText(alt),
			ImageURL: fmt.Sprintf(
				"https://img.shields.io/badge/%s-%s?logo=%s&label=%s",
				url.PathEscape(content), color, url.QueryEscape(c.Icon), url.QueryEscape(c.Label)),
		},
		URL: c.URL,
	}
}

// GitLabDownload is a download renderer for GitLab.
type GitLabDownload struct {
	Owner string
	Repo  string
	URL   string
}

func (g *GitLabDownload) Render() MarkdownRenderer {
	u := first(g.URL, fmt.Sprintf("https://gitlab.com/%s/%s",
		url.PathEscape(g.Owner), url.PathEscape(g.Repo)))
	cs := CustomShieldDownload{
		Icon: "GitLab",
		URL:  u,
	}
	return cs.Render()
}

// DemoDownload displays a Demo button
type DemoDownload struct {
	URL string
}

func (d *DemoDownload) Render() MarkdownRenderer {
	cs := CustomShieldDownload{
		Label:   "Demo",
		Content: &[]string{"Web"}[0],
		Color:   "blue",
		URL:     d.URL,
	}
	return cs.Render()
}

// AppStoreDownload represents a download link to the App Store.
type AppStoreDownload struct {
	ID  string
	URL string
}

func (a *AppStoreDownload) Render() MarkdownRenderer {
	u := first(a.URL, fmt.Sprintf("https://apps.apple.com/app/id%s", a.ID))
	cs := CustomShieldDownload{
		Icon: "App Store",
		URL:  u,
	}
	return cs.Render()
}

// GooglePlayDownload represents a download link to Google Play.
type GooglePlayDownload struct {
	ID  string
	URL string
}

func (g *GooglePlayDownload) Render() MarkdownRenderer {
	u := first(g.URL, fmt.Sprintf("https://play.google.com/store/apps/details?id=%s", g.ID))
	cs := CustomShieldDownload{
		Icon: "Google Play",
		URL:  u,
	}
	return cs.Render()
}

// FallbackDownload is a fallback download renderer.
type FallbackDownload struct{}

func (f *FallbackDownload) Render() MarkdownRenderer {
	return NewText("Unknown", Bold)
}
