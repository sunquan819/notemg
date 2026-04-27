package model

type APIError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *APIError   `json:"error,omitempty"`
}

func OK(data interface{}) APIResponse {
	return APIResponse{Success: true, Data: data}
}

func Err(code int, msg string) APIResponse {
	return APIResponse{Success: false, Error: &APIError{Code: code, Message: msg}}
}
