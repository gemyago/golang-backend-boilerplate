package healthcheckv1

import (
	"log"
	"net/http"
)

func handlePing(w http.ResponseWriter, r *http.Request) {
	if _, err := w.Write([]byte("PONG")); err != nil {
		log.Printf("failed to write response: %v", err.Error())
	}
}
