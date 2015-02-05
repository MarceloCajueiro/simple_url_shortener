package urls

type memoryRepository struct {
	urls   map[string]*Url
	clicks map[string]int
}

func NewMemoryRepository() *memoryRepository {
	return &memoryRepository{
		make(map[string]*Url),
		make(map[string]int),
	}
}

func (r *memoryRepository) IdExist(id string) bool {
	_, exist := r.urls[id]
	return exist
}

func (r *memoryRepository) FindById(id string) *Url {
	return r.urls[id]
}

func (r *memoryRepository) FindByUrl(url string) *Url {
	for _, u := range r.urls {
		if u.Destiny == url {
			return u
		}
	}

	return nil
}

func (r *memoryRepository) Save(url Url) error {
	r.urls[url.Id] = &url
	return nil
}

func (r *memoryRepository) RegisterClick(id string) {
	r.clicks[id] += 1
}

func (r *memoryRepository) FindClicks(id string) int {
	return r.clicks[id]
}
