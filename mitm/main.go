package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
)

func generateSessionID() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	return hex.EncodeToString(b)
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("sessionid")

	if err != nil || cookie.Value == "" {
		newID := generateSessionID()

		http.SetCookie(w, &http.Cookie{
			Name:  "sessionid",
			Value: newID,
			Path:  "/",
		})

		fmt.Println("Assigned new sessionid:", newID)
	}

	http.ServeFile(w, r, "static/index.html")
}

func main() {
	http.HandleFunc("/", rootHandler)

	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	fmt.Println("MITM demo running on http://0.0.0.0:8080 …")
	http.ListenAndServe(":8080", nil)
}
