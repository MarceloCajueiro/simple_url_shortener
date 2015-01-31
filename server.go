package main

import (
  "fmt"
  "log"
  "net/http"
  "strings"
  "time"
  "encoding/json"

  "github.com/marcelocajueiro/url_shortener/urls"
)

var (
  port    int
  urlBase string
  stats   chan   string
)

func init() {
  port    = 8888
  urlBase = fmt.Sprintf("http://localhost:%d", port)
}

func main() {
  stats = make (chan string)
  defer close (stats)
  go newStatistic(stats)

  http.HandleFunc("/api/shorten", Shortener)
  http.HandleFunc("/r/", Redirector)
  http.HandleFunc("/api/stats/", StatsViewer)

  log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
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
    "Link": fmt.Sprintf("<%s/api/stats/%s>; rel=\"stats\"", urlBase, url.Id),
  })
}

func Redirector(w http.ResponseWriter, r *http.Request) {
  path := strings.Split(r.URL.Path, "/")
  id := path[len(path) -1]

  if url := urls.Search(id); url != nil {
    http.Redirect(w, r, url.Destiny, http.StatusMovedPermanently)
    stats <- id
  } else {
    http.NotFound(w, r)
  }
}

func StatsViewer(w http.ResponseWriter, r *http.Request) {
  path := strings.Split(r.URL.Path, "/")
  id := path[len(path) -1]

  if url := urls.Search(id); url != nil {
    json, err := json.Marshal(url.Stats())

    if err != nil {
      w.WriteHeader(http.StatusInternalServerError)
      return
    }

    respondWithJSON(w, string(json))
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
  return string (url)
}

func newStatistic(ids <-chan string) {
  for id := range ids {
    urls.RegisterClick(id)
    fmt.Printf("Click in %s.\n", id)
  }
}
