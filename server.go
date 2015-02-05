package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/marcelocajueiro/simple_url_shortener/urls"
)

var (
	port    int
	urlBase string
	verbose bool
)

func init() {
	flag.IntVar(&port, "p", 8888, "to set a custom port")
	flag.BoolVar(&verbose, "v", false, "to print some log")
	urlBase = fmt.Sprintf("http://localhost:%d", port)

	flag.Parse()
}

func main() {
	stats := make(chan string)
	defer close(stats)
	go newStatistic(stats)

	http.HandleFunc("/api/shorten", Shortener)
	http.Handle("/r/", &Redirector{stats})
	http.HandleFunc("/api/stats/", StatsViewer)

	printLog("Starting server on port %d...", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}

type Redirector struct {
	stats chan string
}
type Headers map[string]string
type Url struct {
	Id        string
	CreatedAt time.Time
	Destiny   string
}

func Shortener(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		respondWith(w, http.StatusMethodNotAllowed, Headers{
			"Allow": "POST",
		})
		return
	}

	url, new, err := urls.FindOrCreateNewUrl(extractUrl(r))

	if err != nil {
		respondWith(w, http.StatusBadRequest, nil)
		return
	}

	var status int

	if new {
		status = http.StatusCreated
	} else {
		status = http.StatusOK
	}

	shortUrl := fmt.Sprintf("%s/r/%s", urlBase, url.Id)
	respondWith(w, status, Headers{
		"Location": shortUrl,
		"Link":     fmt.Sprintf("<%s/api/stats/%s>; rel=\"stats\"", urlBase, url.Id),
	})

	printLog("URL %s successfully shortened to %s", url.Destiny, shortUrl)
}

func (red *Redirector) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	findUrlAndExecute(w, r, func(url *urls.Url) {
		http.Redirect(w, r, url.Destiny, http.StatusMovedPermanently)
		red.stats <- url.Id
	})
}

func StatsViewer(w http.ResponseWriter, r *http.Request) {
	findUrlAndExecute(w, r, func(url *urls.Url) {
		json, err := json.Marshal(url.Stats())

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		respondWithJSON(w, string(json))
	})
}

func findUrlAndExecute(
	w http.ResponseWriter,
	r *http.Request,
	executor func(*urls.Url),
) {
	path := strings.Split(r.URL.Path, "/")
	id := path[len(path)-1]

	if url := urls.Search(id); url != nil {
		executor(url)
	} else {
		http.NotFound(w, r)
	}
}

func respondWith(w http.ResponseWriter, status int, headers Headers) {
	for k, v := range headers {
		w.Header().Set(k, v)
	}
	w.WriteHeader(status)
}

func respondWithJSON(w http.ResponseWriter, response string) {
	respondWith(w, http.StatusOK, Headers{
		"Content-Type": "application/json",
	})
	fmt.Fprintf(w, response)
}

func extractUrl(r *http.Request) string {
	url := make([]byte, r.ContentLength, r.ContentLength)
	r.Body.Read(url)
	return string(url)
}

func newStatistic(ids <-chan string) {
	for id := range ids {
		urls.RegisterClick(id)
		printLog("%s was clicked", id)
	}
}

func printLog(format string, values ...interface{}) {
	if verbose {
		log.Printf(fmt.Sprintf("%s\n", format), values...)
	}
}
