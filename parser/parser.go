package parser

import (
	"golang.org/x/net/idna"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"regexp"
)

var (
	regexEmoticons = regexp.MustCompile("\\((\\w{1,15})\\)")
	regexMentions  = regexp.MustCompile("@(\\w+)")
	regexLinks     = regexp.MustCompile(`http[s]?://(([^/:\.[:space:]]+(\.[^/:\.[:space:]]+)*)|([0-9](\.[0-9]{3})))(:[0-9]+)?((/[^?#[:space:]]+)(\?[^#[:space:]]+)?(\#.+)?)?`)
	regexTitle     = regexp.MustCompile(`<title((.|\n)*?)>((.|\n)*?)</title>`)
)

type MessageInfo struct {
	Mentions  []string `json:"mentions"`
	Emoticons []string `json:"emoticons"`
	Links     []Link   `json:"links"`
}

type Link struct {
	URL   string `json:"url"`
	Title string `json:"title"`
}

type Parser struct {
	httpClient *http.Client
	log        *log.Logger
}

// NewParser creates a new parser
func NewParser(client *http.Client, logger *log.Logger) *Parser {
	return &Parser{
		httpClient: client,
		log:        logger,
	}
}

// Parse returns message info that contains info about emoticons, mentions and links
func (p *Parser) Parse(text string) (mInfo MessageInfo) {
	mInfo.Mentions = p.parseMentions(text)
	mInfo.Emoticons = p.parseEmoticons(text)
	mInfo.Links = p.parseLinks(text)

	return mInfo
}

// parseEmoticons parse text for mentions and remove duplicates
func (p *Parser) parseMentions(text string) []string {
	matches := regexMentions.FindAllStringSubmatch(text, -1)
	uniqueMentions := make(map[string]struct{})

	for _, match := range matches {
		uniqueMentions[match[1]] = struct{}{}
	}

	mentions := make([]string, len(uniqueMentions))

	index := 0
	for emoticon := range uniqueMentions {
		mentions[index] = emoticon
		index++
	}

	return mentions
}

// parseEmoticons parse text for emoticons and remove duplicates
func (p *Parser) parseEmoticons(text string) []string {
	matches := regexEmoticons.FindAllStringSubmatch(text, -1)
	uniqueEmoticons := make(map[string]struct{})

	for _, match := range matches {
		uniqueEmoticons[match[1]] = struct{}{}
	}

	emoticons := make([]string, len(uniqueEmoticons))

	index := 0
	for emoticon := range uniqueEmoticons {
		emoticons[index] = emoticon
		index++
	}

	return emoticons
}

// parseLinks search for links in specified text and retrieves their titles.
// Titles retrieving are handled concurrently.
// If the link is broken a work "Website" will be printed instead of title.
func (p *Parser) parseLinks(text string) []Link {
	matches := regexLinks.FindAllString(text, -1)
	if len(matches) <= 0 {
		return []Link{}
	}
	links := make([]Link, len(matches))
	linkReceiver := make(chan Link)

	for _, link := range matches {
		go func(link string) {
			l := Link{Title: "Website", URL: link}
			defer func() { linkReceiver <- l }()
			sanitizedLink, err := sanitizeUrl(link)

			if err != nil {
				p.log.Printf("Unable to sanitize %s. Error: %s", link, err)
				return
			}

			resp, err := p.httpClient.Get(sanitizedLink)
			if err != nil {
				p.log.Printf("Request to %s failed. Error: %s", link, err)
				return
			}
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				p.log.Printf("Unable to read response body for URL %s. Error: %s", link, err)
				return
			}
			defer resp.Body.Close()

			subMatches := regexTitle.FindAllSubmatch(body, -1)
			if len(subMatches) <= 0 {
				return
			}

			l.Title = string(subMatches[0][3])
		}(link)
	}

	for index := range matches {
		links[index] = <-linkReceiver
	}

	return links
}

// sanitizeUrl converts non-ASCII symbols with punycode algorithm so all domain names can be handled correctly.
func sanitizeUrl(link string) (string, error) {
	//this check is a performance optimization. We rarely get URLs that contain non-ASCII symbols.
	if isASCII(link) {
		return link, nil
	}

	u, err := url.Parse(link)
	if err != nil {
		return "", err
	}
	u.Host, err = idna.ToASCII(u.Host)

	if err != nil {
		return "", err
	}

	return u.String(), nil
}

func isASCII(s string) bool {
	for _, c := range s {
		if c > 127 {
			return false
		}
	}
	return true
}
