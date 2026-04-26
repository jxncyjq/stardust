package databases

type Pageable interface {
	Skip() int
	Limit() int
	Sort() string
}

type pageable struct {
	skip  int
	limit int
	sort  string
}

func (p *pageable) Skip() int {
	return p.skip
}

func (p *pageable) Limit() int {
	return p.limit
}

func (p *pageable) Sort() string {
	return p.sort
}

func NewPageable(skip, limit int, sort string) Pageable {
	return &pageable{skip: skip, limit: limit, sort: sort}
}
