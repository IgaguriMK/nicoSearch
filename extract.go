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
	flag.StringVar(&baseDayStr, "b", "2017-07-23T10:00:00+09:00", "Base date")

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

	fmt.Fprintln(outf, `ID	View	Comment	Mylist	PostAt	At	Day`)

	idMap := newSortMap()

	for dec.More() {
		var vd model.VideoData
		err := dec.Decode(&vd)
		if err != nil {
			log.Fatal("Failed parse json: ", err)
		}

		idMap.Add(vd.ContentID, vd.Title)

		upDay := toDay(vd.StartTime)
		getDay := toDay(vd.UpdatedAt)

		fmt.Fprintf(
			outf,
			"%s\t%d\t%d\t%d\t%f\t%f\t%f\n",
			vd.ContentID,
			vd.ViewCounter,
			vd.CommentCounter,
			vd.MylistCounter,
			upDay-baseDay,
			getDay-baseDay,
			getDay-upDay,
		)
	}

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
	t, err := time.Parse(time.RFC3339, str)
	if err != nil {
		log.Fatal("Time parse error: ", err)
	}

	return float64(t.Unix()) / (24 * 3600)
}
