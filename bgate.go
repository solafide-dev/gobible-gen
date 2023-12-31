package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/PuerkitoBio/goquery"
	"github.com/solafide-dev/gobible/bible"
)

type GatewayBibleBooks struct {
	Data [][]struct {
		Display     string `json:"display"`
		Osis        string `json:"osis"`
		Testament   string `json:"testament"`
		NumChapters int    `json:"num_chapters"`
		Chapters    []struct {
			Chapter int      `json:"chapter"`
			Type    string   `json:"type"`
			Content []string `json:"content"`
		} `json:"chapters"`
		Intro bool `json:"intro"`
	} `json:"data"`
}

type GatewayVersion struct {
	Name      string `json:"name"`
	Abbr      string `json:"abbr"`
	Language  string `json:"language"`
	URL       string `json:"url"`
	Copyright string `json:"copy"`
	About     string `json:"about"`
	Publisher string `json:"publisher"`
}

type GatewayVersions map[string]GatewayVersion

func (v *GatewayVersion) ExpandData() {
	log.Printf("Fetching additional data for %s", v.Name)
	url := "https://www.biblegateway.com" + v.URL + "?interface=print"
	doc := docFromUrl(url)

	// Find the review items
	doc.Find(".vinfo-content").Each(func(i int, s *goquery.Selection) {
		// remove the first h2 node
		s.Find("h2").First().Remove()
		html, err := s.Html()
		if err != nil {
			log.Fatal(err)
		}

		converter := md.NewConverter("", true, nil)
		markdown, err := converter.ConvertString(html)
		if err != nil {
			log.Fatal(err)
		}
		v.About = markdown
		//fmt.Println(markdown)
	})

	v.Publisher = doc.Find(".publisher a").First().Text()

	doc.Find(".copy-content").Each(func(i int, s *goquery.Selection) {
		// remove the first h2 node
		s.Find("h2").First().Remove()
		html, err := s.Html()
		if err != nil {
			log.Fatal(err)
		}

		converter := md.NewConverter("", true, nil)
		markdown, err := converter.ConvertString(html)
		if err != nil {
			log.Fatal(err)
		}
		v.Copyright = markdown
		//fmt.Println(markdown)
	})
}

func getVersions() GatewayVersions {
	fmt.Println("Getting available versions from Bible Gateway")

	versions := map[string]GatewayVersion{}
	doc := docFromUrl("https://www.biblegateway.com/versions/?interface=print")

	// Find the review items
	doc.Find(".info-row").Each(func(i int, s *goquery.Selection) {
		// For each item found, get the title

		text := s.Find("a").Text()
		if text == "" || !strings.Contains(text, "(") {
			return
		}

		language, _ := s.Attr("data-language")

		split := strings.Split(text, "(")
		name := strings.Trim(split[0], " ")
		split2 := strings.Split(split[1], ")")
		abbr := strings.Trim(split2[0], " ")

		url, ok := s.Find("a").Attr("href")
		if !ok {
			log.Fatal("Could not find href")
			return
		}

		versions[abbr] = GatewayVersion{
			Name:     name,
			Abbr:     abbr,
			URL:      url,
			Language: language,
		}
	})

	return versions
}

func getChapterVerses(version, book, chapter string) []bible.Verse {
	log.Println("Getting verses for " + book + " " + chapter)

	url := "https://www.biblegateway.com/passage/?search=" + book + " " + chapter + "&version=" + version + "&interface=print"
	doc := docFromUrl(url)

	verses := []bible.Verse{}

	// Get the main text
	doc.Find(".result-text-style-normal").Each(func(i int, s *goquery.Selection) {
		s.Find("h3").Remove()
		s.Find("sup").Remove() // TODO: footnotes!

		s.Find("p").Each(func(i int, s *goquery.Selection) {
			//html, _ := s.Html()
			//fmt.Println(html + "\n\n")
			s.Find("span").Each(func(i int, s *goquery.Selection) {
				// These are the verses.
				value, _ := s.Attr("class")
				search := "text " + book + "-" + chapter + "-"
				if strings.Contains(value, search) {
					s.Find(".chapternum").Remove()
					//fmt.Println(value, s.Text())
					num, _ := strconv.Atoi(strings.TrimPrefix(value, search))
					//fmt.Println(num)
					v := bible.Verse{
						Number: num,
						Text:   s.Text(),
						// Formatting:
						// Footnotes:
					}
					verses = append(verses, v)
				}
			})
		})

	})

	return verses
}

// Uses a public API at biblegateway.com to get the list of books in a bible translation
func getFromBibleGateway(t string) bible.Bible {

	versions := getVersions()
	if _, ok := versions[t]; !ok {
		log.Fatal("Could not find version " + t)
	}

	version := versions[t]
	version.ExpandData()

	log.Println("Getting books for " + t)

	url := "https://www.biblegateway.com/passage/bcv/?version=" + t

	resp, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	var books GatewayBibleBooks
	err = json.NewDecoder(resp.Body).Decode(&books)
	if err != nil {
		log.Fatal(err)
	}

	b := bible.Bible{
		Version: bible.Version{
			Name:      version.Name,
			Abbrev:    version.Abbr,
			Publisher: version.Publisher,
			Copyright: version.Copyright,
		},
		Extra: bible.Extra{
			About: version.About,
		},
		Books: []bible.Book{},
	}

	for i, book := range books.Data[0] {

		chap := make([]bible.Chapter, len(book.Chapters))

		for j, chapter := range book.Chapters {
			chap[j] = bible.Chapter{
				Name:   book.Display + " " + fmt.Sprint(chapter.Chapter),
				Number: chapter.Chapter,
				Verses: getChapterVerses(t, book.Osis, fmt.Sprint(chapter.Chapter)),
			}

			// If the bible has headings, store them as titles.
			if chapter.Type == "heading" {
				chap[j].Title = chapter.Content[0]
			}
		}

		b.Books = append(b.Books, bible.Book{
			Number:   i + 1,
			Name:     book.Display,
			Chapters: chap,
		})
	}

	return b
}

func docFromUrl(url string) *goquery.Document {
	res, err := http.Get(url)
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

	return doc
}
