package models

type Direction string

const (
	Next     Direction = "next"
	Previous Direction = "previous"
)

type PaginationQuery struct {
	Cursor    *string   `json:"cursor,omitempty"`
	PageSize  int       `json:"pageSize"`
	Direction Direction `json:"direction"`
	Search    *string   `json:"search,omitempty"`
}

type PaginatedResponse struct {
	Data       interface{}        `json:"data"`
	Pagination PaginationResponse `json:"pagination"`
}

type PaginationResponse struct {
	NextCursor     *string `json:"nextCursor,omitempty"`
	PreviousCursor *string `json:"previousCursor,omitempty"`
	HasNext        bool    `json:"hasNext"`
	HasPrevious    bool    `json:"hasPrevious"`
}
