package successresponse

import (
	"encoding/json"
	"net/http"
)

type SuccessResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"` // Omit `data` if it's nil
}

func RespondWithSuccess(w http.ResponseWriter, code int, message string, data interface{}) { //  means it can hold any type of data (string, int, struct, array, map, etc.
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(SuccessResponse{Message: message, Data: data})
}
