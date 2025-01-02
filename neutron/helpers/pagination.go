package helpers

type Pagination struct {
	Offset int `json:"offset"`
	Limit  int `json:"limit"`
	Page   int `json:"page"`
	Size   int `json:"size"`
}

func CalcPaginationByPage(page int, size int) *Pagination {
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 10
	}
	return &Pagination{
		Offset: (page - 1) * size,
		Limit:  size,
		Page:   page,
		Size:   size,
	}
}
