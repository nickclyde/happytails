package jotform

import (
	"fmt"
	"io"
	"net/http"
)

func Jotform(w http.ResponseWriter, r *http.Request) {
	// Copy the request body to the response
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Unable to read request body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	// Write the contents of the request body to the response
	fmt.Fprintf(w, "Request body:\n\n%s", body)
}
