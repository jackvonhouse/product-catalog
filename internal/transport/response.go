package transport

import (
	"encoding/json"
	"net/http"
)

func Response(
	w http.ResponseWriter,
	data interface{},
) {

	w.Header().Add("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(&data); err != nil {
		Error(
			w,
			http.StatusInternalServerError,
			http.StatusText(http.StatusInternalServerError),
		)
	}
}

func Error(
	w http.ResponseWriter,
	statusCode int,
	message string,
) {

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if message != "" {
		json.NewEncoder(w).Encode(
			map[string]string{
				"error": message,
			},
		)
	}
}
