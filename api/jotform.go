package jotform

import (
	"fmt"
	"net/http"
)

func Jotform(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "<h1>Hello from Happy Tails!</h1>")
}
