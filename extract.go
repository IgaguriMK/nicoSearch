package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/IgaguriMK/nicoSearch/model"
)

func main() {
	var baseDayStr string
	flag.StringVar(&baseDayStr, "b", "2007-03-06T00:33:00+09:00", "Base date")

	flag.Parse()

	baseDay := toDay(baseDayStr)

	args := flag.Args()

	if len(args) < 1 {
		log.Fatal("Need input file")
	}

	inputName := args[0]
	basename := strings.TrimSuffix(inputName, ".json")

	inf, err := os.Open(inputName)
	if err != nil {
		log.Fatal("Can't open input: ", err)
	}
	defer inf.Close()

	dec := json.NewDecoder(inf)

	outf, err := os.Create(basename + ".txt")
	if err != nil {
		log.Fatal("Can't create file: ", err)
	}
	defer outf.Close()

	fs := new(fileds)

	fs.Add("ID", "%s", func(v model.VideoData) interface{} { return v.ContentID })
	fs.Add("View", "%d", func(v model.VideoData) interface{} { return v.ViewCounter })
	fs.Add("Comment", "%d", func(v model.VideoData) interface{} { return v.CommentCounter })
	fs.Add("Mylist", "%d", func(v model.VideoData) interface{} { return v.MylistCounter })
	fs.Add("Dur", "%f", func(v model.VideoData) interface{} { return toDay(v.UpdatedAt) - toDay(v.StartTime) })
	fs.Add("Post", "%f", func(v model.VideoData) interface{} { return toDay(v.StartTime) - baseDay })
	fs.Add("Update", "%f", func(v model.VideoData) interface{} { return toDay(v.UpdatedAt) })
	fs.Add("TagNum", "%d", func(v model.VideoData) interface{} { return len(strings.Split(v.Tags, " ")) })
	fs.Add("Length", "%d", func(v model.VideoData) interface{} { return v.LengthSeconds })

	fmt.Fprintln(outf, fs.Header())

	idMap := newSortMap()

	oldest := toTime("2099-12-31T23:59:59+09:00")

	for dec.More() {
		var vd model.VideoData
		err := dec.Decode(&vd)
		if err != nil {
			log.Fatal("Failed parse json: ", err)
		}

		idMap.Add(vd.ContentID, vd.Title)

		fmt.Fprintln(outf, fs.Line(vd))

		t := toTime(vd.StartTime)
		if t.Before(oldest) {
			oldest = t
		}
	}

	fmt.Println(oldest.Format(time.RFC3339))

	idf, err := os.Create(basename + "-title.txt")
	if err != nil {
		log.Fatal("Can't create file: ", err)
	}
	defer idf.Close()

	fmt.Fprintln(idf, `ID	Title`)
	for _, kv := range idMap.List() {
		fmt.Fprintf(idf, "%s\t%q\n", kv.Key, kv.Val)
	}
}

type sortMap struct {
	keys []string
	vals map[string]string
}

func newSortMap() *sortMap {
	return &sortMap{
		keys: make([]string, 0),
		vals: make(map[string]string),
	}
}

func (sm *sortMap) Add(key, val string) {
	if _, found := sm.vals[key]; !found {
		sm.keys = append(sm.keys, key)
		sm.vals[key] = val
	}
}

type keyVal struct {
	Key string
	Val string
}

func (sm *sortMap) List() []keyVal {
	sort.Strings(sm.keys)

	res := make([]keyVal, 0, len(sm.keys))

	for _, k := range sm.keys {
		res = append(res, keyVal{k, sm.vals[k]})
	}

	return res
}

func toDay(str string) float64 {
	t := toTime(str)
	return float64(t.Unix()) / (24 * 3600)
}

func toTime(str string) time.Time {
	t, err := time.Parse(time.RFC3339, str)
	if err != nil {
		log.Fatal("Time parse error: ", err)
	}
	return t
}

type field struct {
	Name   string
	Format string
	GetVal func(v model.VideoData) interface{}
}

type fileds struct {
	fs []field
}

func (f *fileds) Add(name, format string, getVal func(v model.VideoData) interface{}) {
	f.fs = append(
		f.fs,
		field{
			Name:   name,
			Format: format,
			GetVal: getVal,
		},
	)
}

func (f *fileds) Header() string {
	res := make([]string, 0, len(f.fs))

	for _, e := range f.fs {
		res = append(res, e.Name)
	}

	return strings.Join(res, "\t")
}

func (f *fileds) Line(v model.VideoData) string {
	res := make([]string, 0, len(f.fs))

	for _, e := range f.fs {
		res = append(res, fmt.Sprintf(e.Format, e.GetVal(v)))
	}

	return strings.Join(res, "\t")
}
