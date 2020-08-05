package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/aws/aws-lambda-go/lambda"
)

type chapter struct {
	Title   string `json:"title"`
	Text    string `json:"text"`
	Link    string `json:"link"`
	Chapter string `json:"chapter"`
}

type Response struct {
	Message string `json:"message"`
	Ok      bool   `json:"ok"`
}

func main() {
	lambda.Start(Handler)
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
	getErr := getChapter(latestChapters)
	if getErr != nil {
		return Response{
			Message: getErr.Error(),
			Ok:      false,
		}, getErr
	}
	saveErr := save(latestChapters)
	if saveErr != nil {
		return Response{
			Message: saveErr.Error(),
			Ok:      false,
		}, saveErr
	}
	fmt.Println(fmt.Sprintf("Finished crawling for MGA, added %d chapters", len(latestChapters)))

	return Response{
		Message: "ok",
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

func getChapter(chapters []chapter) error {
	for _, chapter := range chapters {
		doc, err := goquery.NewDocument("https://novelfull.com" + chapter.Link)
		if err != nil {
			return err
		}
		doc.Find("div#chapter-content").Each(func(index int, item *goquery.Selection) {
			chapter.Text = sanitize(chapter.Title, item.Text())
		})
	}
	return nil
}

func save(chapters []chapter) error {
	for _, chapter := range chapters {
		requestBody, err := json.Marshal(map[string]string{
			"title":   chapter.Title,
			"text":    chapter.Text,
			"read":    "false",
			"chapter": chapter.Chapter,
		})
		if err != nil {
			return err
		}
		resp, err := http.Post("https://novel-fac48.firebaseio.com/novels/MGA.json", "application/json", bytes.NewBuffer(requestBody))
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		req, err := json.Marshal(map[string]string{
			"text": fmt.Sprintf("Added MGA Chapter %s \n", chapter.Title),
		})
		if err != nil {
			return err
		}
		resp2, err := http.Post(os.Getenv("SLACK"), "application/json", bytes.NewBuffer(req))
		if err != nil {
			return err
		}
		fmt.Println("resp2:", resp2)
		defer resp2.Body.Close()
		fmt.Printf("Added MGA Chapter %s \n", chapter.Title)
	}
	return nil
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