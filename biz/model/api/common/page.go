package common

type PageResponse[T any] struct {
	Records            []T  `json:"records"`
	PageNumber         int  `json:"pageNumber"`
	PageSize           int  `json:"pageSize"`
	TotalPage          int  `json:"totalPage"`
	TotalRow           int  `json:"totalRow"`
	OptimizeCountQuery bool `json:"optimizeCountQuery"`
}
