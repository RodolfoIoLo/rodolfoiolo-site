package handler

import (
	"encoding/json"
	"net/http"
)

type OperationResponse struct {
	Status      int
	ContentType string
	Body        any
	Headers     map[string]string
}

func JSON(status int, data any) *OperationResponse {
	return &OperationResponse{Status: status, ContentType: "application/json", Body: data}
}

func Success(status int, data any) *OperationResponse {
	return JSON(status, map[string]any{"code": 0, "message": "success", "data": data})
}

func Failure(status, code int, message string) *OperationResponse {
	return JSON(status, map[string]any{"code": code, "message": message, "data": nil})
}

func (r *OperationResponse) visit(writer http.ResponseWriter) error {
	contentType := r.ContentType
	if contentType == "" {
		contentType = "application/json"
	}
	writer.Header().Set("Content-Type", contentType)
	for name, value := range r.Headers {
		writer.Header().Set(name, value)
	}
	status := r.Status
	if status == 0 {
		status = http.StatusOK
	}
	writer.WriteHeader(status)
	if r.Body == nil {
		return nil
	}
	if bytes, ok := r.Body.([]byte); ok {
		_, err := writer.Write(bytes)
		return err
	}
	return json.NewEncoder(writer).Encode(r.Body)
}
