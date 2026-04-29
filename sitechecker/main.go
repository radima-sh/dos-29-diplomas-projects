package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"
)

type Site struct {
	URL            string
	Active         bool
	ResponseTimeMs int64
	LastChecked    time.Time
}

var db *sql.DB
var templates *template.Template

func init() {
	// Подключение к БД из переменных окружения
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("PGHOST"),
		os.Getenv("PGPORT"),
		os.Getenv("PGUSER"),
		os.Getenv("PGPASSWORD"),
		os.Getenv("PGDATABASE"),
	)
	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("DB connect error: %v", err)
	}
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)

	// Загрузка шаблонов
	templates = template.Must(template.ParseGlob("templates/*.html"))
}

// indexHandler: главная страница — список сайтов с статусом
func indexHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT url, active FROM sites ORDER BY url")
	if err != nil {
		http.Error(w, "DB query failed", http.StatusInternalServerError)
		log.Printf("DB query error: %v", err)
		return
	}
	defer rows.Close()

	var sites []Site
	for rows.Next() {
		var s Site
		if err := rows.Scan(&s.URL, &s.Active); err != nil {
			log.Printf("Row scan error: %v", err)
			continue
		}
		sites = append(sites, s)
	}

	// index.html ожидает struct{Sites, Title}
	data := struct {
		Sites []Site
		Title string
	}{
		Sites: sites,
		Title: "Site Checker",
	}

	if err := templates.ExecuteTemplate(w, "index.html", data); err != nil {
		http.Error(w, "Template render failed", http.StatusInternalServerError)
		log.Printf("Template error: %v", err)
		return
	}
}

// responseTimeHandler: страница с таблицей времени ответа
func responseTimeHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT url, response_time_ms FROM sites WHERE response_time_ms IS NOT NULL ORDER BY url")
	if err != nil {
		http.Error(w, "DB query failed", http.StatusInternalServerError)
		log.Printf("DB query error: %v", err)
		return
	}
	defer rows.Close()

	var sites []Site
	for rows.Next() {
		var s Site
		if err := rows.Scan(&s.URL, &s.ResponseTimeMs); err != nil {
			log.Printf("Row scan error: %v", err)
			continue
		}
		sites = append(sites, s) //nolint:staticcheck
	}

	// response_time.html ожидает {{range .}}, передаём просто []Site
	if err := templates.ExecuteTemplate(w, "response_time.html", sites); err != nil {
		http.Error(w, "Template render failed", http.StatusInternalServerError)
		log.Printf("Template error: %v", err)
		return
	}
}

// dashboardHandler: дашборд с индикаторами и названиями
func dashboardHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT url, active FROM sites ORDER BY url")
	if err != nil {
		http.Error(w, "DB query failed", http.StatusInternalServerError)
		log.Printf("DB query error: %v", err)
		return
	}
	defer rows.Close()

	var sites []Site
	for rows.Next() {
		var s Site
		if err := rows.Scan(&s.URL, &s.Active); err != nil {
			log.Printf("Row scan error: %v", err)
			continue
		}
		sites = append(sites, s)
	}

	// dashboard.html ожидает {{range .}}, передаём просто []Site
	if err := templates.ExecuteTemplate(w, "dashboard.html", sites); err != nil {
		http.Error(w, "Template render failed", http.StatusInternalServerError)
		log.Printf("Template error: %v", err)
		return
	}
}

func main() {
	// Регистрация всех трёх хендлеров
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/response-time", responseTimeHandler)
	http.HandleFunc("/dashboard", dashboardHandler)

	// Запуск сервера
	port := ":8080"
	log.Printf("Server starting on http://localhost%s", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
