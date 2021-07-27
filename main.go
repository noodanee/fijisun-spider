package main

import (
	"encoding/csv"
	"encoding/json"
	"flag"
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

	useUrl := flag.String("url", "", "Input target url.")
	useOut := flag.String("out", "", "Out put the file.")
	useFormat := flag.String("format", "json", "File format.")
	flag.Parse()

	if *useUrl == "" {
		log.Fatalf("Mast be input url")
	}

	if *useOut == "" {
		log.Fatalf("Mast be input url")
	}

	if *useFormat != "json" && *useFormat != "csv" {
		log.Fatalf("The file format only support 'json' and 'csv'.")
	}

	c := colly.NewCollector(colly.Async(true))
	c.Limit(&colly.LimitRule{
		Parallelism: 5,
	})

	c.OnError(func(response *colly.Response, err error) {
		log.Println("ERROR", err.Error())
		//response.Request.Retry()
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
				log.Println("ERROR", err.Error())
				//response.Request.Retry()
			})

			cc.OnRequest(func(request *colly.Request) {
				log.Println("--> ", request.URL.String())
			})

			cc.OnHTML(".main-article-content", func(e *colly.HTMLElement) {
				date := e.ChildText("div.article-controls > div.left-side > div")
				content := e.ChildText("div.shortcode-content p")
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
			c.Visit(href)
		}
	})

	c.Visit(*useUrl)

	c.Wait()

	//enc := json.NewEncoder(os.Stdout)
	//enc.SetIndent("", "  ")
	//enc.Encode(articles)

	file, err := os.Create(*useOut)

	if err != nil {
		log.Fatalf("Failed creating file: %s", err)
	}
	defer file.Close()

	// Write UTF-8 BOM, support windows os show chinese
	file.WriteString("\xEF\xBB\xBF")

	if *useFormat == "json" {
		writer := json.NewEncoder(file)
		writer.SetIndent("", "  ")
		writer.Encode(articles)
		return
	}

	if *useFormat == "csv" {
		writer := csv.NewWriter(file)
		writer.Write([]string{"类别", "标题", "简介", "作者", "链接", "时间", "内容"})
		for _, article := range articles {
			writer.Write([]string{article.Tag, article.Title, article.Intro, article.Author, article.Url, article.Date, article.Content})
		}
		writer.Flush()
		return
	}

	log.Fatalf("Invalid filename.")

}
