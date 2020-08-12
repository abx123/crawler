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
	opt := option.WithCredentialsJSON([]byte(`{
		"type": "service_account",
		"project_id": "novel-fac48",
		"private_key_id": "f75f7a985b9972e38834bb6771f3596d03c1c0ab",
		"private_key": "-----BEGIN PRIVATE KEY-----\nMIIEvgIBADANBgkqhkiG9w0BAQEFAASCBKgwggSkAgEAAoIBAQC84TbsXWp6i/p2\nMm2UYtiqE/8hbR8Xf4Lpk4B3IpeWTfXzlzjyRVG/Tso/wyjdNHG4VkNFQ0d94sIs\najHA0sCval2Q2u8k0owVVn5+RdVt50kX06mSa+LkWdbTXra3uZFPWncMb3FL+jTP\naMb7eq0UvJ9w1NX7hhLuC4xeS559GYbuDvXvfnVLwVUlmZfF//XNZhOBNZAokv1f\ntVnqJFXAQYc7hdjTixxUge/Ka2hVb32L2WWRVZfqYlnQTKWChU+jrvX2G2HR/itK\nFDlE7kFtNtyJ+qx20bSPgtUoMYtzr9t5U+Dsk89AKOTAX5EgAZKrFuMPMJs04w+m\n6mCZ2YsPAgMBAAECggEAI7cwhAx8DHU4pK4Pc34ney2x0jfIp9BYSGO4aI61fFn8\nlpWzUniSJys2ak00hnOax2Ekck3xEFhXID/qbYxMnD7wN2p2yw83JvfGjokU/SW9\neBBxobrd2hE04p4nzeD8nbU9CrBuC5Bh+RBWhAoj/WZXfeX5Gok1PicX4WLKMtxT\nZucVFP2GSFo4LvDcryrjQlj3jaJYZW/doV1EQlgzXgDD10Gud3nIs/1er42rnXyi\nCxzsduNh3RgiPK71Lx0gdyxgTbPOxcdUbcbpFyuZGa7sDD3SoiquL8VuwAyc6wOO\n97XYRiAhlmN+wLdxOvDsLCxdM2t95cyYCaDbepAibQKBgQD8fKVHpGi5oS/G9I+X\nKWaE+o+8WICIyVuLgdz4cN7JzqEcbueUHlwXm54yN1GS1Gdq755z5ZrZJIQ4/M9E\no/dcp/1qxHY72+J796bOFl51BcCkluqCCwsGsl3+5G3D+1tJcYMrb9FUBag0b5Vu\n7JEVoU/UCqfSnYb8H51p6Xuv3QKBgQC/ggBRGEPN6NdOJkaG2DSfXumDSBvfT5gP\nK3057sNOW/LoC9SIDGfK1XUoDluYGzHKSwaTROtjh05YFtHB4FB1wrZwqFTU81kl\nvWLAcdb9TiqA42WhBvMtxEeqU4Tglk4a3yJGqN2hpr9CAnjt8IFk9a/vXUPqKgzD\nu9XRgb1t2wKBgQCKORqqm+ERLqLfQmeRk4KibiFeNP045TMOrqtv/yqYRFyDGlwB\nBJXZ/sGeMBaiUVHEgyW1wQ8CrTENmalGpJT4zqa3WpJ3tqrIvw08aZaQbfPGpy/+\nvVjt85vtvNQypFqXXGM41mA8pVQuUJ/4N949fzAanzK85KxPPmeI4d9qqQKBgQCE\nN73+PzF49TvJIdXpfVX/fijcUamkqLBEMPNZTwYakJMJMDnA4Ee8m1kymY8VWhkr\nIFdez+NwKNenK8IQB82lMBSDfURsbcJrsvB+C1qyMghYSic9YK3+OBh+eQExibRN\nCyb//9BcreI4MbrKFBVR3epk6VBdWEDN1l5OMjPVpwKBgGc/R6ByKgJD2k1FUhR2\nNH5ygq4jJdZklZCfaekrOpabFan1m8e0dog9uU++UHLDK6uh0URg/20TZNHfNC8D\nPvsy/9pzazIks4aQPvzH9y0R+t3uERUgd1g8xD+Ccl07bgq+5IH9vZENVUM76bSB\nVYiy10FvU3n3Z7RcSMod90m8\n-----END PRIVATE KEY-----\n",
		"client_email": "firebase-adminsdk-cabau@novel-fac48.iam.gserviceaccount.com",
		"client_id": "101950008019101536286",
		"auth_uri": "https://accounts.google.com/o/oauth2/auth",
		"token_uri": "https://oauth2.googleapis.com/token",
		"auth_provider_x509_cert_url": "https://www.googleapis.com/oauth2/v1/certs",
		"client_x509_cert_url": "https://www.googleapis.com/robot/v1/metadata/x509/firebase-adminsdk-cabau%40novel-fac48.iam.gserviceaccount.com"
	  }
	  `))
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
			"text": fmt.Sprintf("Added %s %s \n", novel, chapter.Title),
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
