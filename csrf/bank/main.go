package main

import (
    "crypto/rand"
    "encoding/hex"
    "fmt"
    "net/http"
    "os"
    "sync"
)

var (
    mu sync.Mutex

    passwords = map[string]string{
        "victim":   "123",
        "attacker": "123",
    }

    balances = map[string]int{
        "victim":   1500,
        "attacker": 0,
    }

    history = map[string][]string{
        "victim":   {"История операций:"},
        "attacker": {"История операций:"},
    }

    sessions = map[string]string{}
)

func newSession() string {
    buf := make([]byte, 16)
    rand.Read(buf)
    return hex.EncodeToString(buf)
}

func currentUser(r *http.Request) (string, bool) {
    c, err := r.Cookie("sessionid")
    if err != nil {
        return "", false
    }
    mu.Lock()
    user := sessions[c.Value]
    mu.Unlock()
    if user == "" {
        return "", false
    }
    return user, true
}

func loginHandler(w http.ResponseWriter, r *http.Request) {

    // ========== GET — показать страницу логина ==========
    if r.Method == http.MethodGet {
        if _, ok := currentUser(r); ok {
            http.Redirect(w, r, "/bank", http.StatusFound)
            return
        }

        page, _ := os.ReadFile("static/login.html")

        // если есть ?error=1 — вставим маленький JS для показа красной плашки
        if r.URL.Query().Get("error") == "1" {
            page = append(page, []byte(`
                <script>
                    setTimeout(() => {
                        alert("Неверный логин или пароль");
                    }, 100);
                </script>
            `)...)
        }

        w.Header().Set("Content-Type", "text/html; charset=utf-8")
        w.Write(page)
        return
    }

    // ========== POST — логин ==========
    if r.Method == http.MethodPost {
        user := r.FormValue("user")
        pass := r.FormValue("pass")

        mu.Lock()
        correctPass := passwords[user]
        mu.Unlock()

        if correctPass == "" || pass != correctPass {
            http.Redirect(w, r, "/login?error=1", http.StatusSeeOther)
            return
        }

        sid := newSession()
        mu.Lock()
        sessions[sid] = user
        mu.Unlock()

        http.SetCookie(w, &http.Cookie{
            Name:  "sessionid",
            Value: sid,
            Path:  "/",
        })

        http.Redirect(w, r, "/bank", http.StatusFound)
    }
}

func bankPage(w http.ResponseWriter, r *http.Request) {
    _, ok := currentUser(r)
    if !ok {
        http.Redirect(w, r, "/login", http.StatusFound)
        return
    }

    page, _ := os.ReadFile("static/bank.html")
    w.Header().Set("Content-Type", "text/html; charset=utf-8")
    w.Write(page)
}

func balanceAPI(w http.ResponseWriter, r *http.Request) {
    user, ok := currentUser(r)
    if !ok {
        http.Error(w, "Not logged in", http.StatusUnauthorized)
        return
    }

    mu.Lock()
    bal := balances[user]
    mu.Unlock()

    w.Write([]byte(fmt.Sprintf("Баланс (%s): %d ₽", user, bal)))
}

func historyAPI(w http.ResponseWriter, r *http.Request) {
    user, ok := currentUser(r)
    if !ok {
        http.Error(w, "Not logged in", http.StatusUnauthorized)
        return
    }

    mu.Lock()
    h := history[user]
    mu.Unlock()

    out := ""
    for _, x := range h {
        out += x + "\n"
    }
    w.Write([]byte(out))
}

func transferAPI(w http.ResponseWriter, r *http.Request) {
    user, ok := currentUser(r)
    if !ok {
        http.Error(w, "Not logged in", http.StatusUnauthorized)
        return
    }

    to := r.URL.Query().Get("to")
    sum := 1000

    mu.Lock()
    defer mu.Unlock()

    if balances[user] < sum {
        msg := "Ошибка: недостаточно средств"
        history[user] = append(history[user], msg)
        w.Write([]byte(msg))
        return
    }

    if _, exists := balances[to]; !exists {
        msg := "Ошибка: получатель не существует"
        history[user] = append(history[user], msg)
        w.Write([]byte(msg))
        return
    }

    balances[user] -= sum
    balances[to] += sum

    msg := fmt.Sprintf("Переведено %d ₽ → %s", sum, to)
    history[user] = append(history[user], msg)
    history[to] = append(history[to], fmt.Sprintf("Получено %d ₽ от %s", sum, user))

    fmt.Println("Операция:", msg)
    w.Write([]byte(msg))
}

func main() {
    http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        http.Redirect(w, r, "/login", http.StatusFound)
    })

    http.HandleFunc("/login", loginHandler)
    http.HandleFunc("/bank", bankPage)

    http.HandleFunc("/balance", balanceAPI)
    http.HandleFunc("/history", historyAPI)
    http.HandleFunc("/transfer", transferAPI)

    fmt.Println("Bank running at http://localhost:8000 ...")
    http.ListenAndServe(":8000", nil)
}
