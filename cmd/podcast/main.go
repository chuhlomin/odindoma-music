// A small app to parse RSS feed for OdinDoma podcast and print out the tracks
package main

import (
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"
)

const feedURL = "https://cloud.mave.digital/36700"

type podcast struct {
	Channel struct {
		Items []struct {
			Title   string `xml:"title"`
			Summary string `xmlns:"itunes" xml:"summary"`
		} `xml:"item"`
	} `xml:"channel"`
}

func main() {
	if err := run(); err != nil {
		log.Fatalf("Error: %v", err)
	}
}

func run() error {
	resp, err := http.Get(feedURL)
	if err != nil {
		return fmt.Errorf("failed to get feed: %w", err)
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	var p podcast
	if err := xml.Unmarshal(b, &p); err != nil {
		return fmt.Errorf("failed to unmarshal xml: %w", err)
	}

	regIntroAutro := regexp.MustCompile(`(?m)^(Интро|Аутро) (подкаста|выпуска):\s+(.*)$`)
	regEpisode := regexp.MustCompile(`(?m)Разрыв танцполов.* (песн[я|ю|и]( группы)?|композиция|трека?|кавер) (.*?).?$`)

	for _, item := range p.Channel.Items {
		if regIntroAutro.MatchString(item.Summary) {
			matches := regIntroAutro.FindAllStringSubmatch(item.Summary, -1)

			for _, match := range matches {
				fmt.Printf("%s\n", strings.TrimLeft(match[3], " "))
			}
		}

		if regEpisode.MatchString(item.Summary) {
			matches := regEpisode.FindAllStringSubmatch(item.Summary, -1)

			for _, match := range matches {
				fmt.Printf("%s\n", match[3])
			}
		}

	}

	return nil
}
