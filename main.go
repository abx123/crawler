package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type chapter struct {
	Title   string `json:"title"`
	Text    string `json:"text"`
	Link    string `json:"link"`
	Chapter string `json:"chapter"`
}

type Request struct {
	ID    float64 `json="id"`
	Value syting  `json="value"`
}

type Response struct {
	Message string `json:"message"`
	Ok      bool   `json:"ok"`
}

func main() {
	// lambda.Start(Handler)
	Handler()
}

func Handler() (Response, error) {
	fmt.Println("Start crawling for MGA")
	existingChapters, err := getExistingChapters()
	if err != nil {
		return Response{
			Message: err.Error(),
			Ok:      false,
		}, err
	}
	var latest int64
	if len(existingChapters) > 0 {
		latest, _ = strconv.ParseInt(existingChapters[0].Chapter, 10, 64)
	}
	latestChapters, latestErr := getLatestChapters(latest)
	if latestErr != nil {
		return Response{
			Message: latestErr.Error(),
			Ok:      false,
		}, latestErr
	}
	getChapter(latestChapters)
	save(latestChapters)
	fmt.Println(fmt.Sprintf("Finished crawling for MGA, added %d chapters", len(latestChapters)))

	return Response{
		Message: fmt.Sprintf("Process Request ID %f", request.ID),
		Ok:      true,
	}, nil
}

func getLatestChapters(currentChapter int64) ([]chapter, error) {
	var latestChapters []chapter
	re := regexp.MustCompile("[0-9]+")
	doc, err := goquery.NewDocument("https://novelfull.com/martial-god-asura.html")
	if err != nil {
		return nil, err
	}
	doc.Find("body div .l-chapters a").Each(func(index int, item *goquery.Selection) {
		link, _ := item.Attr("href")
		chap := re.FindAllString(item.Text(), -1)[0]
		curChap, _ := strconv.ParseInt(chap, 10, 64)
		if curChap > currentChapter {
			latestChapters = append(latestChapters, chapter{
				Title:   item.Text(),
				Link:    link,
				Chapter: chap,
			})
		}
	})
	return latestChapters, nil
}

func sanitize(title string, input string) string {
	output := strings.ReplaceAll(input, "(adsbygoogle = window.adsbygoogle || []).push({});", "")
	output = strings.ReplaceAll(output, title, "")
	o := strings.Split(output, "If you find any errors")

	return o[0]
}

func getChapter(chapters []chapter) {
	for _, chapter := range chapters {
		doc, _ := goquery.NewDocument("https://novelfull.com" + chapter.Link)
		doc.Find("div#chapter-content").Each(func(index int, item *goquery.Selection) {
			chapter.Text = sanitize(chapter.Title, item.Text())
		})
	}
}

func save(chapters []chapter) {
	for _, chapter := range chapters {
		requestBody, _ := json.Marshal(map[string]string{
			"title":   chapter.Title,
			"text":    chapter.Text,
			"read":    "false",
			"chapter": chapter.Chapter,
		})
		resp, _ := http.Post("https://novel-fac48.firebaseio.com/novels/MGA.json", "application/json", bytes.NewBuffer(requestBody))
		defer resp.Body.Close()

		req, _ := json.Marshal(map[string]string{
			"text": fmt.Sprintf("Added MGA Chapter %s \n", chapter.Title),
		})
		resp2, _ := http.Post("https://hooks.slack.com/services/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", "application/json", bytes.NewBuffer(req))
		defer resp2.Body.Close()
		fmt.Printf("Added MGA Chapter %s \n", chapter.Title)

	}
}

func getExistingChapters() ([]chapter, error) {
	var test map[string]chapter
	var chapterList []chapter
	resp, err := http.Get("https://novel-fac48.firebaseio.com/novels/MGA.json")
	defer resp.Body.Close()
	if err != nil {
		return nil, err
	}

	b, readErr := ioutil.ReadAll(resp.Body)
	if readErr != nil {
		return nil, readErr
	}
	json.Unmarshal(b, &test)
	for _, chapter := range test {
		chapterList = append(chapterList, chapter)
	}
	sort.Slice(chapterList, func(i, j int) bool {
		chapi, _ := strconv.ParseInt(chapterList[i].Chapter, 10, 64)
		chapj, _ := strconv.ParseInt(chapterList[j].Chapter, 10, 64)
		return chapi > chapj
	})
	return chapterList, nil
}
