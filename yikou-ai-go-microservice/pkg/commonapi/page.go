package common

type PageResponse[T any] struct {
	Records            []T  `json:"records"`
	PageNum            int  `json:"pageNum"`
	PageSize           int  `json:"pageSize"`
	TotalPage          int  `json:"totalPage"`
	TotalRow           int  `json:"totalRow"`
	OptimizeCountQuery bool `json:"optimizeCountQuery"`
}
