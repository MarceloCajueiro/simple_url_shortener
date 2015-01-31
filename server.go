package main

import (
  "fmt"
  "log"
  "net/http"
  "strings"
  "time"

  "github.com/marcelocajueiro/url_shortener/url"
)

var (
  port    int
  urlBase string
)

func init() {
  port    = 8888
  urlBase = fmt.Sprintf("http://localhost:%d", port)
}

func main() {
  http.HandleFunc("/api/shorten", Shortener)
  http.HandleFunc("/r/", Redirector)

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

  url, new, err := url.FindOrCreateNewUrl(extractUrl(r))

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
  respondWith(w, status, Headers{"Location": shortUrl})
}

func Redirector(w http.ResponseWriter, r *http.Request) {
  path := strings.Split(r.URL.Path, "/")
  id := path[len(path) -1]

  if url := url.Search(id); url != nil {
    http.Redirect(w, r, url.Destiny, http.StatusMovedPermanently)
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

func extractUrl(r *http.Request) string {
  url := make([]byte, r.ContentLength, r.ContentLength)
  r.Body.Read(url)
  return string (url)
}
