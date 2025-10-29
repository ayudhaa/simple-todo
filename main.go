package main

import (
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"unicode"
)

type Todo struct {
	ID   int
	Text string
	Done bool
}

type PageData struct {
	Todos      []Todo
	Keyword    string
	Page       int
	TotalPages int
	Error      string
	Success    string
}

var (
	todos []Todo
	mu    sync.Mutex
)

const todosPerPage = 10

var funcMap = template.FuncMap{
	"add": func(a, b int) int { return a + b },
	"sub": func(a, b int) int { return a - b },
}

var tmpl = template.Must(template.New("index.html").Funcs(funcMap).ParseFiles("templates/index.html"))

func main() {
	http.HandleFunc("/", listTodos)
	http.HandleFunc("/add", addTodo)
	http.HandleFunc("/done", markDone)
	http.HandleFunc("/delete", deleteTodo)
	http.HandleFunc("/clear", clearTodos)

	log.Println("🚀 Server berjalan di http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func listTodos(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	keyword := strings.TrimSpace(r.URL.Query().Get("q"))
	pageStr := r.URL.Query().Get("page")
	page := 1
	if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
		page = p
	}

	var filtered []Todo
	if keyword != "" {
		for _, t := range todos {
			if strings.Contains(strings.ToLower(t.Text), strings.ToLower(keyword)) {
				filtered = append(filtered, t)
			}
		}
	} else {
		filtered = todos
	}

	total := len(filtered)
	start := (page - 1) * todosPerPage
	end := start + todosPerPage
	if start > total {
		start = total
	}
	if end > total {
		end = total
	}
	paged := filtered[start:end]

	totalPages := (total + todosPerPage - 1) / todosPerPage

	data := PageData{
		Todos:      paged,
		Keyword:    keyword,
		Page:       page,
		TotalPages: totalPages,
	}

	// Check for error/success messages in URL parameters
	if msg := r.URL.Query().Get("error"); msg != "" {
		data.Error = msg
	}
	if msg := r.URL.Query().Get("success"); msg != "" {
		data.Success = msg
	}

	tmpl.Execute(w, data)
}

func addTodo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	text := strings.TrimSpace(r.FormValue("todo"))

	// Validasi input kosong
	if text == "" {
		http.Redirect(w, r, "/?error="+encodeURLParam("Input tidak boleh kosong"), http.StatusSeeOther)
		return
	}

	// Validasi jika hanya berisi simbol/karakter spesial
	if isOnlySpecialChars(text) {
		http.Redirect(w, r, "/?error="+encodeURLParam("Input tidak boleh hanya berisi simbol atau karakter spesial"), http.StatusSeeOther)
		return
	}

	mu.Lock()
	defer mu.Unlock()
	todos = append(todos, Todo{ID: len(todos) + 1, Text: text})
	http.Redirect(w, r, "/?success="+encodeURLParam("Todo berhasil ditambahkan!"), http.StatusSeeOther)
}

func markDone(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, _ := strconv.Atoi(idStr)

	mu.Lock()
	defer mu.Unlock()
	for i := range todos {
		if todos[i].ID == id {
			todos[i].Done = true
			break
		}
	}
	http.Redirect(w, r, "/?success="+encodeURLParam("Todo ditandai sebagai selesai!"), http.StatusSeeOther)
}

func deleteTodo(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, _ := strconv.Atoi(idStr)

	mu.Lock()
	defer mu.Unlock()
	for i := range todos {
		if todos[i].ID == id {
			todos = append(todos[:i], todos[i+1:]...)
			break
		}
	}
	http.Redirect(w, r, "/?success="+encodeURLParam("Todo berhasil dihapus!"), http.StatusSeeOther)
}

func clearTodos(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()
	todos = []Todo{}
	http.Redirect(w, r, "/?success="+encodeURLParam("Semua todo berhasil dihapus!"), http.StatusSeeOther)
}

// Fungsi untuk mengecek apakah string hanya berisi karakter spesial
func isOnlySpecialChars(s string) bool {
	hasAlphanumeric := false
	for _, char := range s {
		if unicode.IsLetter(char) || unicode.IsDigit(char) || unicode.IsSpace(char) {
			hasAlphanumeric = true
			break
		}
	}
	return !hasAlphanumeric
}

// Fungsi untuk encode parameter URL
func encodeURLParam(s string) string {
	return strings.ReplaceAll(s, " ", "+")
}
