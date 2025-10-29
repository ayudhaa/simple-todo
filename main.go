package main

import (
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

type Todo struct {
	ID   int
	Text string
	Done bool
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

	data := struct {
		Todos      []Todo
		Keyword    string
		Page       int
		TotalPages int
	}{
		Todos:      paged,
		Keyword:    keyword,
		Page:       page,
		TotalPages: totalPages,
	}

	tmpl.Execute(w, data)
}

func addTodo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	text := strings.TrimSpace(r.FormValue("todo"))
	if text == "" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	mu.Lock()
	defer mu.Unlock()
	todos = append(todos, Todo{ID: len(todos) + 1, Text: text})
	http.Redirect(w, r, "/", http.StatusSeeOther)
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
	http.Redirect(w, r, "/", http.StatusSeeOther)
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
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func clearTodos(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()
	todos = []Todo{}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
