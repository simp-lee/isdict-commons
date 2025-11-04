package model

// APIResponse is the unified response structure for all API endpoints
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
	Error   *ErrorInfo  `json:"error"`
	Meta    *MetaInfo   `json:"meta,omitempty"`
}

// ErrorInfo represents error details
type ErrorInfo struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

// MetaInfo represents pagination or statistical information
type MetaInfo struct {
	Page      *int     `json:"page,omitempty"`
	PageSize  *int     `json:"page_size,omitempty"`
	Total     *int64   `json:"total,omitempty"`
	Requested *int     `json:"requested,omitempty"`
	Found     *int     `json:"found,omitempty"`
	NotFound  []string `json:"not_found,omitempty"`
	Limit     *int     `json:"limit,omitempty"`
	Offset    *int     `json:"offset,omitempty"`
}

// NewSuccessResponse creates a successful API response
func NewSuccessResponse(data interface{}) *APIResponse {
	return &APIResponse{
		Success: true,
		Data:    data,
		Error:   nil,
		Meta:    nil,
	}
}

// NewSuccessResponseWithMeta creates a successful API response with metadata
func NewSuccessResponseWithMeta(data interface{}, meta *MetaInfo) *APIResponse {
	return &APIResponse{
		Success: true,
		Data:    data,
		Error:   nil,
		Meta:    meta,
	}
}

// NewErrorResponse creates an error API response
func NewErrorResponse(code, message string, details interface{}) *APIResponse {
	return &APIResponse{
		Success: false,
		Data:    nil,
		Error: &ErrorInfo{
			Code:    code,
			Message: message,
			Details: details,
		},
		Meta: nil,
	}
}
