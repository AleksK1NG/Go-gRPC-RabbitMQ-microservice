package utils

import (
	"fmt"
	"math"
	"strconv"
)

const (
	defaultSize = 10
)

// Pagination query params
type PaginationQuery struct {
	Size    uint64 `json:"size,omitempty"`
	Page    uint64 `json:"page,omitempty"`
	OrderBy string `json:"orderBy,omitempty"`
}

// Set page size
func (q *PaginationQuery) SetStringSize(sizeQuery string) error {
	if sizeQuery == "" {
		q.Size = defaultSize
		return nil
	}
	n, err := strconv.Atoi(sizeQuery)
	if err != nil {
		return err
	}
	q.Size = uint64(n)

	return nil
}

// Set page number
func (q *PaginationQuery) SetStringPage(pageQuery string) error {
	if pageQuery == "" {
		q.Size = 0
		return nil
	}
	n, err := strconv.Atoi(pageQuery)
	if err != nil {
		return err
	}
	q.Page = uint64(n)

	return nil
}

// Set order by
func (q *PaginationQuery) SetOrderBy(orderByQuery string) {
	q.OrderBy = orderByQuery
}

// Get offset
func (q *PaginationQuery) GetOffset() uint64 {
	if q.Page == 0 {
		return 0
	}
	return (q.Page - 1) * q.Size
}

// Get limit
func (q *PaginationQuery) GetLimit() uint64 {
	return q.Size
}

// Get OrderBy
func (q *PaginationQuery) GetOrderBy() string {
	return q.OrderBy
}

// Get OrderBy
func (q *PaginationQuery) GetPage() uint64 {
	return q.Page
}

// Get OrderBy
func (q *PaginationQuery) GetSize() uint64 {
	return q.Size
}

//
func (q *PaginationQuery) GetQueryString() string {
	return fmt.Sprintf("page=%v&size=%v&orderBy=%s", q.GetPage(), q.GetSize(), q.GetOrderBy())
}

// Get total pages int
func GetTotalPages(totalCount uint64, pageSize uint64) uint64 {
	d := float64(totalCount) / float64(pageSize)
	return uint64(math.Ceil(d))
}

// Get has more
func GetHasMore(currentPage uint64, totalCount uint64, pageSize uint64) bool {
	return currentPage < totalCount/pageSize
}
