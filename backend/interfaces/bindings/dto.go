package bindings

import (
	"strings"

	"github.com/google/uuid"

	"vfinancy/backend/internal/domain/repositories"
)

// PaginationRequest is the page/pageSize envelope sent by the frontend.
type PaginationRequest struct {
	Page     int `json:"page"`
	PageSize int `json:"pageSize"`
}

// toPageRequest clamps the values into a repositories.PageRequest.
func (p PaginationRequest) toPageRequest() repositories.PageRequest {
	page := p.Page
	if page < 1 {
		page = 1
	}
	size := p.PageSize
	if size < 1 {
		size = 25
	}
	if size > 200 {
		size = 200
	}
	return repositories.PageRequest{Limit: size, Offset: (page - 1) * size}
}

// PageResult is the paginated response envelope sent to the frontend.
type PageResult struct {
	Items    interface{} `json:"items"`
	Total    int         `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"pageSize"`
}

// parseOptionalUUID parses a UUID string, returning nil for empty input.
func parseOptionalUUID(s string) (*uuid.UUID, error) {
	if strings.TrimSpace(s) == "" {
		return nil, nil
	}
	id, err := uuid.Parse(s)
	if err != nil {
		return nil, err
	}
	return &id, nil
}

// uuidPtrString formats a UUID pointer as a string, "" for nil.
func uuidPtrString(id *uuid.UUID) string {
	if id == nil {
		return ""
	}
	return id.String()
}
