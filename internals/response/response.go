package response

import (
	"encoding/json"
	"net/http"
)

type ErrorResponse struct {
	Error Error `json:"error"`
}

type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func JSON(w http.ResponseWriter, status int, data any){
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	json.NewEncoder(w).Encode(data)
}

func ErrorJSON(w http.ResponseWriter, status int, code string, message string){
	JSON(w, status, ErrorResponse{Error: Error{Code: code, Message: message}})
}