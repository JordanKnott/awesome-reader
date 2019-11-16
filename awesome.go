package main

import (
	"encoding/json"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strings"
)

var GITHUB = regexp.MustCompile(`(?P<proto>https?:\/\/)?github.com\/(?P<owner>[^\/]*)\/(?P<repo>[^\/]*)\/?`)

func HasElement(s *goquery.Selection, tag string) bool {
	elementFound := false
	s.Each(func(index int, inner *goquery.Selection) {
		if goquery.NodeName(inner) == tag {
			elementFound = true
			return
		}
	})
	return elementFound
}

func getREADME(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("GET error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Status error: %v", resp.StatusCode)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("Read body: %v", err)
	}

	return string(data), nil
}

type GithubRepo struct {
	Owner string
	Repo  string
}
type Link struct {
	Name        string
	Description string
	GithubRepo  GithubRepo
	Categories  []string
}

func main() {
	res, err := http.Get("https://awesome-go.com")
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
	}

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	links := []Link{}

	// Find the review items
	doc.Find("div > ul li").Each(func(i int, s *goquery.Selection) {
		if s.Find("li").Length() == 0 {
			if goquery.NodeName(s) != "li" {
				return
			}
			linkFull := strings.TrimSpace(s.Text())
			if linkFull != "pq - Pure Go Postgres driver for database/sql." {
				return
			}
			linkName := strings.TrimSpace(s.Find("a").First().Text())

			linkParts := strings.Split(linkFull, "-")
			linkDescription := ""
			if len(linkParts) >= 2 {
				linkDescription = strings.TrimSpace(strings.Join(linkParts[1:], ""))
			}
			linkHref, _ := s.Find("a").First().Attr("href")
			if href, found := s.Find("a").Attr("href"); found {
				if strings.HasPrefix(href, "#") {
					return
				}
			}
			matches := GITHUB.FindStringSubmatch(linkHref)
			linkHrefOwner := ""
			linkHrefRepo := ""
			if len(matches) == 4 {
				linkHrefOwner = matches[2]
				linkHrefRepo = matches[3]
			}

			var prevNode *goquery.Selection
			categories := []string{}
			currentNode := s
			level := 0
			for goquery.NodeName(currentNode) != "div" {
				prevNode = currentNode
				currentNode = currentNode.Parent()
				if goquery.NodeName(currentNode) == "ul" {
					p := currentNode.Prev()
					if goquery.NodeName(p) == "p" {
						if !strings.HasPrefix(goquery.NodeName(p.Prev()), "h") {
							level += 1
							categories = append(categories, p.Text())
						}
					}
				} else if goquery.NodeName(currentNode) == "li" {
					html, _ := currentNode.Html()
					text := strings.Split(html, "\n")
					firstLine := strings.TrimSpace(text[0])
					if strings.HasPrefix(firstLine, "<p>") && strings.HasSuffix(firstLine, "</p>") {
						firstLine = strings.Replace(firstLine, "<p>", "", -1)
						firstLine = strings.Replace(firstLine, "</p>", "", -1)
					}
					level += 1
					categories = append(categories, firstLine)
				}
			}
			currentNode = prevNode.Prev()
			for goquery.NodeName(currentNode) != "h2" {
				if strings.HasPrefix(goquery.NodeName(currentNode), "h") {
					level += 1
					categories = append(categories, strings.TrimSpace(currentNode.Text()))
					currentNode = currentNode.Prev()
				} else {
					currentNode = currentNode.Prev()
				}
			}
			level += 1
			categories = append(categories, strings.TrimSpace(currentNode.Text()))
			for i := len(categories)/2 - 1; i >= 0; i-- {
				opp := len(categories) - 1 - i
				categories[i], categories[opp] = categories[opp], categories[i]
			}
			catSize := len(categories)
			if catSize >= 2 {
				if categories[catSize-1] == categories[catSize-2] {
					categories = categories[0 : catSize-1]
				}
			}
			links = append(links, Link{
				Name:        linkName,
				Description: linkDescription,
				GithubRepo: GithubRepo{
					Owner: linkHrefOwner,
					Repo:  linkHrefRepo,
				},
				Categories: categories,
			})
		}
	})

	b, err := json.Marshal(links)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(b))
}
