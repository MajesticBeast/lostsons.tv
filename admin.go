package main

import (
	"html/template"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func adminRouter() chi.Router {
	r := chi.NewRouter()
	r.Get("/", handleAdminIndex)

	return r
}

// Admin Handlers
//
// --> /admin/index
func handleAdminIndex(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("./templates/admin/index.html")
	if err != nil {
		log.Fatal(err)
	}

	data := map[string]string{
		"Title":   "Admin Index",
		"Content": "This is the admin index page",
	}

	if err := t.Execute(w, data); err != nil {
		log.Fatal(err)
	}
}
