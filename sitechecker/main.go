// rebuild: 2026-04-28-force
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

 // Шаблон response_time.html использует {{range .}}, поэтому передаём просто слайс
        if err := templates.ExecuteTemplate(w, "response_time.html", sites); err != nil {
                http.Error(w, "Template render failed", http.StatusInternalServerError)
                log.Printf("Template error: %v", err)
                return
        }
}

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
		sites = append(sites, s)
	}

}

func dashboardHandler(w http.ResponseWriter, r *http.Request) {
	// Запрашиваем сайты из БД (как в indexHandler)
	rows, err := db.Query("SELECT url, active FROM sites ORDER BY url")
	if err != nil {
		http.Error(w, "DB query failed", http.StatusInternalServerError)
		log.Printf("DB query error: %v", err)
		return
	}
	defer rows.Close()

	// Собираем результаты в слайс
	var sites []Site
	for rows.Next() {
		var s Site
		if err := rows.Scan(&s.URL, &s.Active); err != nil {
			log.Printf("Row scan error: %v", err)
			continue
		}
		sites = append(sites, s)
	}

	// Передаём слайс в шаблон (шаблон ожидает {{range .}})
	if err := templates.ExecuteTemplate(w, "dashboard.html", sites); err != nil {
		http.Error(w, "Template render failed", http.StatusInternalServerError)
		log.Printf("Template error: %v", err)
		return
	}
}

func main() {
	// Регистрация хендлеров
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

// redeploy Вт 28 апр 2026 18:26:51 MSK
