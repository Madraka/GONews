package models

// PaginatedResponse represents a paginated response structure
type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	Page       int         `json:"page"`
	Limit      int         `json:"limit"`
	TotalItems int         `json:"totalItems"`
	TotalPages int         `json:"totalPages"`
	HasNext    bool        `json:"hasNext"`
	HasPrev    bool        `json:"hasPrev"`
}
