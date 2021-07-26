package main

import (
	"encoding/json"
	"github.com/gocolly/colly/v2"
	"log"
	"os"
)

type Article struct {
	Tag string `json:"tag"`
	Title string `json:"title"`
	Intro string `json:"intro"`
	Author string `json:"author"`
	Url string `json:"url"`
	Content string `json:"content"`
	Date string `json:"date"`
}

func main() {

	c := colly.NewCollector(colly.Async(true))
	c.Limit(&colly.LimitRule{
		Parallelism: 5,
	})

	c.OnError(func(response *colly.Response, err error) {
		log.Println(err.Error())
	})

	c.OnRequest(func(request *colly.Request) {
		log.Println("-> ", request.URL.String())
	})

	articles := []Article{}

	c.OnHTML(".main-content-left .main-no-split", func(e *colly.HTMLElement) {
		e.ForEach("div.article-big-block", func(_ int, el *colly.HTMLElement) {
			tag := el.ChildText("span.article-slug a")
			title := el.ChildText("div.article-header a")
			intro := el.ChildText("div.article-content p")
			author := el.ChildText(".article-author")
			url := el.ChildAttr("div.article-header a", "href")

			cc := c.Clone()

			cc.OnError(func(response *colly.Response, err error) {
				log.Println(err.Error())
			})

			cc.OnRequest(func(request *colly.Request) {
				log.Println("--> ", request.URL.String())
			})

			cc.OnHTML(".main-article-content", func(e *colly.HTMLElement) {
				date := e.ChildText("div.article-controls > div.left-side > div")
				content := e.ChildText("div.shortcode-content > p")
				articles = append(articles, Article{
					Tag: tag,
					Title: title,
					Intro: intro,
					Author: author,
					Url: url,
					Date: date,
					Content: content,
				})
			})

			cc.Visit(url)

			cc.Wait()
		})
	})

	c.OnHTML("div.page-pager a.next.page-numbers", func(e *colly.HTMLElement) {
		href := e.Attr("href")
		if href != "" {
			log.Println(href)
			//c.Visit(href)
		}
	})

	c.Visit("https://fijisun.com.fj/page/1/?s=Chinese+Tourism")

	c.Wait()

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.Encode(articles)
}
