package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/IgaguriMK/nicoSearch/model"
)

func main() {
	var thumbDir string
	flag.StringVar(&thumbDir, "thumb", "", "Thumb dir")

	flag.Parse()

	args := flag.Args()

	if thumbDir == "" {
		log.Fatal("Need thumbDir")
	}

	if len(args) < 1 {
		log.Fatal("Need input file")
	}

	inputName := args[0]

	inf, err := os.Open(inputName)
	if err != nil {
		log.Fatal("Can't open input: ", err)
	}
	defer inf.Close()

	dec := json.NewDecoder(inf)

	fmt.Println("<html>")
	fmt.Println("<head>")
	fmt.Printf("<title>%q</title>\n", inputName)
	fmt.Println("</head>")
	fmt.Println("<body>")

	var cnt int
	fmt.Println("<h2>0</h2>")

	for dec.More() {
		var vd model.VideoData
		err := dec.Decode(&vd)
		if err != nil {
			log.Fatal("Failed parse json: ", err)
		}

		fmt.Printf(
			`<a href="http://www.nicovideo.jp/watch/%s"><img src="./%s/%s.jpg"></a>`+"\n",
			vd.ContentID,
			thumbDir,
			vd.ContentID,
		)

		cnt++

		if cnt%5 == 0 {
			fmt.Println("<br>")
		}
		if cnt%50 == 0 {
			fmt.Println("<h2>", cnt, "</h2>")
		} else if cnt%10 == 0 {
			fmt.Println(cnt, "<br>")
		}
	}

	fmt.Println("</body>")
	fmt.Println("</html>")
}
