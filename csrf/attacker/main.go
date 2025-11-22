package main

import (
    "fmt"
    "net/http"
)

func main() {
    fs := http.FileServer(http.Dir("static"))
    http.Handle("/", fs)

    fmt.Println("Attacker running at http://localhost:9000 ...")
    http.ListenAndServe(":9000", nil)
}
