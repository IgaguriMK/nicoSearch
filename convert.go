package main

import (
	"encoding/json"
	"flag"
	"io"
	"log"
	"os"

	"github.com/IgaguriMK/nicoSearch/model"
)

func main() {
	var descriptionDir string
	flag.StringVar(&descriptionDir, "desc", "", "Split description into directory")

	flag.Parse()

	args := flag.Args()

	if len(args) == 0 {
		log.Fatal("Need file")
	}

	if descriptionDir == "" {
		log.Fatal("Need descriptionDir")
	}

	fileName := args[0]
	backup := fileName + ".bak"

	backupFile(fileName, backup)

	in, err := os.Open(backup)
	if err != nil {
		log.Fatal(err)
	}
	defer in.Close()

	out, err := os.Create(fileName)
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()

	dec := json.NewDecoder(in)
	enc := json.NewEncoder(out)

	for dec.More() {
		var v model.VideoData
		err = dec.Decode(&v)
		if err != nil {
			log.Fatal(err)
		}

		err = v.SplitDescription(descriptionDir)
		if err != nil {
			log.Fatal(err)
		}

		err = enc.Encode(v)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func backupFile(from, to string) {
	in, err := os.Open(from)
	if err != nil {
		log.Fatal(err)
	}
	defer in.Close()

	out, err := os.Create(to)
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		log.Fatal(err)
	}
}
