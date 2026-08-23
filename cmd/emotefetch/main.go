// Command emotefetch vendors the language's keyword emotes into a
// static directory the playground page can serve locally, so rendering
// does not depend on 7tv being reachable at runtime.
//
// For each keyword it resolves an emote id in priority order:
//  1. hand-pinned ids (curated picks that beat whatever the channel has)
//  2. towdan's channel emote set on 7tv
//  3. a global 7tv exact-name search (v4 GraphQL endpoint)
//
// It downloads each emote as 2x.webp, skips files already present, and
// writes manifest.json mapping names to filenames.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/ohnishat/peepoScript/cmd/evaluator"
	"github.com/ohnishat/peepoScript/cmd/token"
)

const (
	towdanTwitchID = "76020462"
	gqlEndpoint    = "https://7tv.io/v4/gql"
	channelAPI     = "https://7tv.io/v3/users/twitch/" + towdanTwitchID
)

// pinned holds curated emote ids that override discovery.
var pinned = map[string]string{
	"peepoChat": "01GEZX9N900003T6Y8KR7ZKS00",
}

type file struct {
	Name string `json:"name"`
}

type host struct {
	URL   string `json:"url"`
	Files []file `json:"files"`
}

type emoteData struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Host host   `json:"host"`
}

type channelEmote struct {
	Name string `json:"name"`
	Data struct {
		Host host `json:"host"`
	} `json:"data"`
}

type channelResponse struct {
	EmoteSet struct {
		Emotes []channelEmote `json:"emotes"`
	} `json:"emote_set"`
}

func main() {
	out := flag.String("out", "web/emotes", "directory to write emotes into")
	flag.Parse()

	if err := run(*out); err != nil {
		log.Fatal(err)
	}
}

func run(out string) error {
	if err := os.MkdirAll(out, 0o755); err != nil {
		return err
	}

	client := &http.Client{}
	channel, err := fetchChannel(client)
	if err != nil {
		log.Printf("channel set unavailable (%v), falling back to search only", err)
		channel = map[string]channelEmote{}
	}

	manifest := map[string]string{}
	var missing []string
	symbols := append(token.Keywords(), evaluator.Builtins()...)
	for _, name := range symbols {
		emote, ok := resolve(name, channel)
		if !ok {
			missing = append(missing, name)
			continue
		}
		file := name + ".webp"
		path := filepath.Join(out, file)
		if _, err := os.Stat(path); err == nil {
			log.Printf("%s: already vendored", name)
		} else if err := download(client, emote, path); err != nil {
			log.Printf("%s: download failed: %v", name, err)
			missing = append(missing, name)
			continue
		}
		manifest[name] = file
	}
	for _, name := range missing {
		delete(manifest, name)
	}

	data, err := json.MarshalIndent(map[string]any{"emotes": manifest}, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(out, "manifest.json"), append(data, '\n'), 0o644); err != nil {
		return err
	}

	log.Printf("vendored %d/%d symbols into %s", len(manifest), len(symbols), out)
	if len(missing) > 0 {
		log.Printf("no emote found for: %v", missing)
	}
	return nil
}

func resolve(name string, channel map[string]channelEmote) (host, bool) {
	if id, ok := pinned[name]; ok {
		return host{URL: "//cdn.7tv.app/emote/" + id}, true
	}
	if emote, ok := channel[name]; ok {
		return emote.Data.Host, true
	}
	id, ok := searchEmoteID(name)
	if !ok {
		return host{}, false
	}
	return host{URL: "//cdn.7tv.app/emote/" + id}, true
}

func fetchChannel(client *http.Client) (map[string]channelEmote, error) {
	res, err := client.Get(channelAPI)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("7tv returned %s", res.Status)
	}
	var data channelResponse
	if err := json.NewDecoder(res.Body).Decode(&data); err != nil {
		return nil, err
	}
	found := make(map[string]channelEmote, len(data.EmoteSet.Emotes))
	for _, emote := range data.EmoteSet.Emotes {
		found[emote.Name] = emote
	}
	return found, nil
}

type gqlResponse struct {
	Data struct {
		Search struct {
			All struct {
				Emotes struct {
					Items []struct {
						ID          string `json:"id"`
						DefaultName string `json:"defaultName"`
					} `json:"items"`
				} `json:"emotes"`
			} `json:"all"`
		} `json:"search"`
	} `json:"data"`
}

// searchEmoteID asks the global 7tv search for an exact-name match.
func searchEmoteID(name string) (string, bool) {
	gql := fmt.Sprintf(`query{search{all(query:%q,perPage:1){emotes{items{id defaultName}}}}}`, name)
	body, _ := json.Marshal(map[string]string{"query": gql})
	res, err := http.Post(gqlEndpoint, "application/json", strings.NewReader(string(body)))
	if err != nil || res.StatusCode != http.StatusOK {
		return "", false
	}
	defer res.Body.Close()

	var data gqlResponse
	if err := json.NewDecoder(res.Body).Decode(&data); err != nil {
		return "", false
	}
	items := data.Data.Search.All.Emotes.Items
	if len(items) == 0 || items[0].DefaultName != name {
		return "", false
	}
	return items[0].ID, true
}

func download(client *http.Client, emote host, path string) error {
	url := "https:" + emote.URL + "/2x.webp"
	res, err := client.Get(url)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("%s returned %s", url, res.Status)
	}
	data, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644) //nolint:gosec // trusted CDN content
}
