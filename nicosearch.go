package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	apiLimit      = 100
	callInterval  = 5 * time.Second
	thumbInterval = 100 * time.Millisecond
)

var wg = new(sync.WaitGroup)

func main() {
	var mode string
	flag.StringVar(&mode, "mode", "keyword", "Search mode [keyword, title, tag]")

	var thumbsDir string
	flag.StringVar(&thumbsDir, "thumbs", "", "Thumbnail save dir")

	var limit int
	flag.IntVar(&limit, "l", 10000, "Limit")

	flag.Parse()

	args := flag.Args()

	if len(args) == 0 {
		log.Fatal("Need search text")
	}

	title := args[0]

	ch := callAll(title, mode, limit)
	thumbCh := saveThumbs(thumbsDir)

	fmt.Print("[\n\t")
	cnt := 0
	for d := range ch {
		if cnt > 0 {
			fmt.Print(",\n\t")
		}

		thumbCh <- d

		enc := json.NewEncoder(os.Stdout)
		err := enc.Encode(d)
		if err != nil {
			log.Fatal("Encode error: ", err)
		}

		cnt++
	}
	fmt.Println("\n]")

	close(thumbCh)

	wg.Wait()
}

func callAll(searchText, mode string, softLimt int) chan videoData {
	ch := make(chan videoData, 10)

	go func() {
		defer close(ch)

		offset := 0

		for {
			resp, err := callAPI(searchText, mode, offset)
			if err != nil {
				log.Fatal("API error: ", err)
			}

			for _, d := range resp.Data {
				ch <- d
			}

			offset += apiLimit
			if offset >= softLimt {
				return
			}
			if resp.Meta.TotalCount < apiLimit {
				return
			}

			time.Sleep(callInterval)
		}
	}()

	return ch
}

func callAPI(searchText, mode string, offset int) (*resp, error) {
	service := "video"

	params := url.Values{}

	params.Add("q", searchText)

	switch mode {
	case "keyword":
		params.Add("targets", "title,description,tags")
	case "title":
		params.Add("targets", "title")
	case "tag":
		params.Add("targets", "tagsExact")
	}

	params.Add("fields", strings.Join(
		[]string{
			"contentId",
			"title",
			"description",
			"categoryTags",
			"tags",
			"viewCounter",
			"commentCounter",
			"mylistCounter",
			"startTime",
			"lengthSeconds",
			"thumbnailUrl",
		},
		",",
	))
	params.Add("_sort", "+startTime")
	params.Add("_limit", strconv.Itoa(apiLimit))
	params.Add("_offset", strconv.Itoa(offset))

	url := fmt.Sprintf("http://api.search.nicovideo.jp/api/v2/%s/contents/search?%s", service, params.Encode())

	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	resp := new(resp)
	err = json.NewDecoder(res.Body).Decode(resp)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

type resp struct {
	Data []videoData `json:"data"`
	Meta struct {
		ID         string `json:"id"`
		Status     int64  `json:"status"`
		TotalCount int64  `json:"totalCount"`
	} `json:"meta"`
}

type videoData struct {
	CategoryTags   string `json:"categoryTags"`
	CommentCounter int64  `json:"commentCounter"`
	ContentID      string `json:"contentId"`
	Description    string `json:"description"`
	LengthSeconds  int64  `json:"lengthSeconds"`
	MylistCounter  int64  `json:"mylistCounter"`
	StartTime      string `json:"startTime"`
	Tags           string `json:"tags"`
	ThumbnailURL   string `json:"thumbnailUrl"`
	Title          string `json:"title"`
	ViewCounter    int64  `json:"viewCounter"`
}

func saveThumbs(dir string) chan videoData {
	ch := make(chan videoData, 1024)

	if dir == "" {
		go func() {
			for {
				if _, ok := <-ch; !ok {
					return
				}
			}
		}()
		return ch
	}

	err := os.MkdirAll(dir, 0744)
	if err != nil {
		log.Fatal("Can't create directory: ", err)
	}

	wg.Add(1)

	go func() {
		for d := range ch {
			res, err := http.Get(d.ThumbnailURL)
			if err != nil {
				log.Printf("Thumbnail get error <%s>: %s", d.ThumbnailURL, err)
				time.Sleep(5 * time.Second)
				continue
			}

			f, err := os.Create(filepath.Join(dir, d.ContentID+".jpg"))
			if err != nil {
				log.Fatal("File error: ", err)
			}

			_, err = io.Copy(f, res.Body)
			if err != nil {
				log.Fatal("File error: ", err)
			}

			f.Close()
			res.Body.Close()
			time.Sleep(thumbInterval)
		}

		wg.Done()
	}()

	return ch
}
