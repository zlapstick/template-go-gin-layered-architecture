package helper

var (
	ErrInternalServerError = "terjadi kesalahan pada server"
)

type Meta struct {
	From       int `json:"from"`
	To         int `json:"to"`
	Page       int `json:"page"`
	PerPage    int `json:"per_page"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

type MessageResponse struct {
	Message string `json:"message"`
}

type DataResponse struct {
	Message string `json:"message"`
	Data    any    `json:"data"`
}

type PaginatedResponse struct {
	Message string `json:"message"`
	Data    any    `json:"data"`
	Meta    *Meta  `json:"meta"`
}

type ErrorResponse struct {
	Message string            `json:"message"`
	Errors  map[string]string `json:"errors"`
}

func NewMeta(page, perPage, total, count int) *Meta {
	from, to := 0, 0
	if total > 0 {
		from = (page-1)*perPage + 1
		to = from + count - 1
	}
	return &Meta{
		From:       from,
		To:         to,
		Page:       page,
		PerPage:    perPage,
		Total:      total,
		TotalPages: (total + perPage - 1) / perPage,
	}
}
