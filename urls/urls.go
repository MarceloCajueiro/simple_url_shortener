package urls

import (
  "math/rand"
  "net/url"
  "time"
)

const (
  size = 5
  symbols = "qwertyuiopasdfghjklzxcvbnmQWERTYUIOPASDFGHJKLZXCVBNM1234567890_-+"
)

func init() {
  rand.Seed(time.Now().UnixNano())
  ConfigRepository(NewMemoryRepository())
}

type Url struct {
  Id        string
  CreatedAt time.Time
  Destiny   string
}

type Repository interface {
  IdExist(id string) bool
  FindById(id string) *Url
  FindByUrl(url string) *Url
  Save(url Url) error
  RegisterClick(id string)
}

var repo Repository

func ConfigRepository(r Repository) {
  repo = r
}

func FindOrCreateNewUrl(destiny string) (u *Url, new bool, err error) {
  if u = repo.FindByUrl(destiny); u != nil {
    return u, false, nil
  }

  if _, err = url.ParseRequestURI(destiny); err != nil {
    return nil, false, err
  }

  url := Url{generateId(), time.Now(), destiny}
  repo.Save(url)
  return &url, true, nil
}

func generateId() string {
  newId := func () string {
    id := make([]byte, size, size)
    for i := range id {
      id[i] = symbols[rand.Intn(len(symbols))]
    }

    return string(id)
  }

  for {
    if id := newId(); !repo.IdExist(id) {
      return id
    }
  }
}

func Search(id string) *Url {
  return repo.FindById(id)
}

func RegisterClick(id string) {
  repo.RegisterClick(id)
}
