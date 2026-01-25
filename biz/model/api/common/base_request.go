package common

type DeleteRequest struct {
	Id int `json:"id"`
}

type PageRequest struct {
	PageNumber int    `json:"pageNumber"`
	PageSize   int    `json:"pageSize"`
	SortField  string `json:"sortField"`
	SortOrder  string `json:"sortOrder"`
}
