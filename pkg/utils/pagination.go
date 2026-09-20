package utils

// Pagination dipakai ulang oleh seluruh endpoint yang menampilkan list.
type Pagination struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

const (
	defaultPage  = 1
	defaultLimit = 10
	maxLimit     = 100
)

// NewPagination membuat instance Pagination dengan nilai default jika page atau limit tidak valid.
func NewPagination(page, limit int) *Pagination {
	if page < 1 {
		page = defaultPage
	}
	if limit < 1 {
		limit = defaultLimit
	}
	if limit > maxLimit {
		limit = maxLimit
	}
	return &Pagination{Page: page, Limit: limit}
}

// Offset menghitung offset untuk query database
func (p *Pagination) Offset() int { return (p.Page - 1) * p.Limit }

//	SetTotal menghitung total halaman berdasarkan total data dan limit
func (p *Pagination) SetTotal(total int64) {
	p.Total = total
	if p.Limit > 0 {
		p.TotalPages = int((total + int64(p.Limit) - 1) / int64(p.Limit))
	}
}
