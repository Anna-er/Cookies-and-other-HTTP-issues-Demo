package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
)

var (
	comments []string
	mu       sync.Mutex
)

func main() {
	http.HandleFunc("/", formHandler)
	http.HandleFunc("/steal", stealHandler)

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	fmt.Println("XSS demo running on http://localhost:8081 …")
	http.ListenAndServe(":8081", nil)
}

func formHandler(w http.ResponseWriter, r *http.Request) {
	// Уязвимость: НЕ экранируем HTML
	if r.Method == http.MethodPost {
		r.ParseForm()
		mu.Lock()
		comments = append(comments, r.Form.Get("comment"))
		mu.Unlock()
	}

	// Ставим cookie без HttpOnly, чтобы XSS видел её
	http.SetCookie(w, &http.Cookie{
		Name:  "sessionid",
		Value: "XSS_SECRET_456",
		Path:  "/",
	})

	// Загружаем HTML-файл
	html, err := os.ReadFile("static/index.html")
	if err != nil {
		http.Error(w, "index.html not found", 500)
		return
	}
	page := string(html)

	// Формируем HTML блок комментариев (уязвимый!)
	mu.Lock()
	var list strings.Builder
	for _, c := range comments {
		list.WriteString("<p>" + c + "</p>")
	}
	mu.Unlock()

	// Заменяем {{COMMENTS}} в HTML
	page = strings.Replace(page, "{{COMMENTS}}", list.String(), 1)

	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, page)
}

func stealHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Stolen cookies:", r.URL.Query().Get("c"))
	fmt.Fprintln(w, "OK")
}
