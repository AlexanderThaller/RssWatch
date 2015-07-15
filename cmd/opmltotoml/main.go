package main

import (
	"flag"
	"fmt"
	"log"
	"strings"

	"github.com/gilliek/go-opml/opml"
)

var (
	buildVersion string
	buildTime    string

	flagSrcFile = flag.String("src", "feedlist.opml",
		"The path to the source file.")
)

func init() {
	flag.Parse()
}

func main() {
	doc, err := opml.NewOPMLFromFile(*flagSrcFile)
	if err != nil {
		log.Fatal(err)
	}

	printTitles(doc.Body.Outlines, "")
}

func printTitles(outlines []opml.Outline, folder string) {
	for _, outline := range outlines {
		switch outline.Type {
		case "atom":
			fmt.Println("[[Feeds]]")
			fmt.Println(`  Url = "` + outline.XMLURL + `"`)
			fmt.Println(`  Filters = [".*"]`)
			fmt.Println(`  Folder = "` + folder + `"`)
			fmt.Println("")

		case "rss":
			fmt.Println("[[Feeds]]")
			fmt.Println(`  Url = "` + outline.HTMLURL + `"`)
			fmt.Println(`  Filters = [".*"]`)
			fmt.Println(`  Folder = "` + folder + `"`)
			fmt.Println("")

		case "folder":
			f := strings.TrimPrefix(folder+"."+outline.Title, ".")
			printTitles(outline.Outlines, f)
		}
	}
}
