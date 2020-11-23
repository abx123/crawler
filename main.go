package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/db"
	"github.com/PuerkitoBio/goquery"
	"github.com/aws/aws-lambda-go/lambda"
	"google.golang.org/api/option"
)

var client *db.Client

var (
	URL = make(map[string]string)
)

type chapter struct {
	Title   string `json:"title"`
	Text    string `json:"text"`
	Link    string `json:"link"`
	Chapter int64  `json:"chapter"`
	Novel   string `json:"novel"`
}

type Response struct {
	Message string `json:"message"`
	Ok      bool   `json:"ok"`
}

func Init() {
	ctx := context.Background()
	conf := &firebase.Config{
		DatabaseURL: "https://novel-fac48.firebaseio.com",
	}
	opt := option.WithCredentialsJSON([]byte(os.Getenv("FIREBASE")))

	app, err := firebase.NewApp(ctx, conf, opt)
	if err != nil {
		log.Fatalf("firebase.NewApp: %v", err)
	}
	client, err = app.Database(ctx)
	if err != nil {
		log.Fatalf("app.Firestore: %v", err)
	}

	URL["MGA"] = "https://novelfull.com/martial-god-asura.html"
	URL["MYBWH"] = "https://novelfull.com/my-youth-began-with-him.html"
	URL["ROTSSG"] = "https://novelfull.com/reincarnation-of-the-strongest-sword-god.html"
}

func main() {
	Init()
	lambda.Start(Handler)
}

func crawl(novel string) error {
	fmt.Println("Start crawling for " + novel)
	latestChapter, err := getLatestChapter(novel)
	if err != nil {
		return err
	}
	latestChapters, latestErr := crawlLatestChapters(latestChapter, novel)
	if latestErr != nil {
		return latestErr
	}
	chapters, getErr := getChapter(latestChapters)
	if getErr != nil {
		return getErr
	}
	saveErr := save(chapters, novel)
	if saveErr != nil {
		return saveErr
	}
	fmt.Println(fmt.Sprintf("Finished crawling for %s, added %d chapters", novel, len(latestChapters)))
	return nil
}

func Handler() (Response, error) {
	for key, _ := range URL {
		err := crawl(key)
		if err != nil {
			return Response{
				Message: err.Error(),
				Ok:      false,
			}, err
		}
	}
	return Response{
		Message: "ok",
		Ok:      true,
	}, nil
}

func crawlLatestChapters(latestChapter int64, novel string) ([]chapter, error) {

	var latestChapters []chapter
	re := regexp.MustCompile("[0-9]+")
	doc, err := goquery.NewDocument(URL[novel])
	if err != nil {
		return nil, err
	}
	doc.Find("body div .l-chapters a").Each(func(index int, item *goquery.Selection) {
		link, _ := item.Attr("href")
		chap := re.FindAllString(item.Text(), -1)[0]
		curChap, _ := strconv.ParseInt(chap, 10, 64)
		if curChap > latestChapter {
			latestChapters = append(latestChapters, chapter{
				Title:   item.Text(),
				Link:    link,
				Chapter: curChap,
				Novel:   novel,
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

func getChapter(chapters []chapter) ([]chapter, error) {
	var resp []chapter
	for _, chapter := range chapters {
		doc, err := goquery.NewDocument("https://novelfull.com" + chapter.Link)
		if err != nil {
			return resp, err
		}
		doc.Find("div#chapter-content").Each(func(index int, item *goquery.Selection) {
			chapter.Text = sanitize(chapter.Title, item.Text())
			resp = append(resp, chapter)
		})
	}
	return resp, nil
}

func save(chapters []chapter, novel string) error {
	sort.Slice(chapters, func(i, j int) bool {
		return chapters[i].Chapter < chapters[j].Chapter
	})
	for _, chapter := range chapters {
		ref := client.NewRef("novels/" + novel + "/" + strconv.FormatInt(chapter.Chapter, 10))
		if err := ref.Set(context.Background(), chapter); err != nil {
			return err
		}
		req, err := json.Marshal(map[string]string{
			"text": fmt.Sprintf("Added %s %s \n https://novelfull.com%s", novel, chapter.Title, chapter.Link),
		})
		if err != nil {
			return err
		}
		resp2, err := http.Post(os.Getenv("SLACK"), "application/json", bytes.NewBuffer(req))
		if err != nil {
			return err
		}
		defer resp2.Body.Close()
		fmt.Printf("Added %s Chapter %s \n", novel, chapter.Title)
	}
	return nil
}

func getLatestChapter(novel string) (int64, error) {
	q := client.NewRef("novels/" + novel).OrderByChild("chapter").LimitToLast(1)
	result, err := q.GetOrdered(context.Background())
	if err != nil || len(result) == 0 {
		return 0, err
	}
	latestChapter, _ := strconv.ParseInt(result[0].Key(), 10, 64)
	return latestChapter, nil
}
