package common

type PageResponse[T any] struct {
	Records            []T  `json:"records"`
	PageNumber         int  `json:"page_number"`
	PageSize           int  `json:"page_size"`
	TotalPage          int  `json:"total_page"`
	TotalRow           int  `json:"total_row"`
	OptimizeCountQuery bool `json:"optimize_count_query"`
}
