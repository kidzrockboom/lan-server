package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"text/template"

	"github.com/joho/godotenv"
)

type PageData struct {
	Files        []os.DirEntry
	CurrentVideo string
}

var DIR string

func readDir(w http.ResponseWriter, req *http.Request) {

	videos, _ := os.ReadDir(DIR)

	// get video from Url
	selectedVideo := req.URL.Query().Get("play")

	tmpl, _ := template.ParseFiles("web/templates/index.html.tmpl")

	data := PageData{
		Files:        videos,
		CurrentVideo: selectedVideo,
	}

	fmt.Println(data.CurrentVideo)

	tmpl.Execute(w, data)
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Printf("no .env file found: %v", err)
	}

	DIR = os.Getenv("DIR")
	if DIR == "" {
		log.Fatal("DIR must be set in .env or the environment")
	}

	addr := flag.String("addr", ":8080", "address to listen on (binds all interfaces, not just localhost)")
	flag.Parse()

	// http.FileServer handles Range requests out of the box, so video
	// seeking/scrubbing works without extra code. See NOTES.md step 1.
	http.HandleFunc("/", readDir)

	http.Handle("/stream/", http.StripPrefix("/stream/", http.FileServer(http.Dir(DIR))))

	log.Printf("serving on %s", *addr)
	log.Fatal(http.ListenAndServe(*addr, nil))
}
