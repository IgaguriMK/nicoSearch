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

	"github.com/IgaguriMK/nicoSearch/model"
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

	var limit int
	flag.IntVar(&limit, "l", 10000, "Limit")

	var thumbsDir string
	flag.StringVar(&thumbsDir, "thumbs", "", "Thumbnail save dir")

	var descriptionDir string
	flag.StringVar(&descriptionDir, "desc", "", "Split description into directory")

	flag.Parse()

	args := flag.Args()

	if len(args) == 0 {
		log.Fatal("Need search text")
	}

	title := args[0]

	if descriptionDir != "" {
		err := os.MkdirAll(descriptionDir, 0744)
		if err != nil {
			log.Fatal("Can't create directory: ", err)
		}
	}

	ch := callAll(title, mode, limit)
	thumbCh := saveThumbs(thumbsDir)

	for d := range ch {
		if descriptionDir != "" {
			err := d.SplitDescription(descriptionDir)
			if err != nil {
				log.Fatal("Failed save split description: ", err)
			}
		}

		thumbCh <- d

		enc := json.NewEncoder(os.Stdout)
		err := enc.Encode(d)
		if err != nil {
			log.Fatal("Encode error: ", err)
		}

	}

	close(thumbCh)

	wg.Wait()
}

func callAll(searchText, mode string, softLimt int) chan model.VideoData {
	ch := make(chan model.VideoData, 10)

	updatedAt, err := lastMod()
	if err != nil {
		log.Fatal("API error: ", err)
	}

	go func() {
		defer close(ch)

		offset := 0

		for {
			resp, err := callAPI(searchText, mode, offset)
			if err != nil {
				log.Fatal("API error: ", err)
			}

			for _, d := range resp.Data {
				d.UpdatedAt = updatedAt
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
	Data []model.VideoData `json:"data"`
	Meta struct {
		ID         string `json:"id"`
		Status     int64  `json:"status"`
		TotalCount int64  `json:"totalCount"`
	} `json:"meta"`
}

func saveThumbs(dir string) chan model.VideoData {
	ch := make(chan model.VideoData, 1024)

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

func lastMod() (string, error) {
	res, err := http.Get("http://api.search.nicovideo.jp/api/v2/snapshot/version")
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	var lm struct {
		LastModified string `json:"last_modified"`
	}
	err = json.NewDecoder(res.Body).Decode(&lm)
	if err != nil {
		return "", err
	}

	return lm.LastModified, nil
}
