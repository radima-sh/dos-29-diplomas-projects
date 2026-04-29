package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"
	"os"

	_ "github.com/lib/pq"
)

type Site struct {
	ID int
	URL string
	Active bool
	LastCheck time.Time
	ResponseTimeMs int64
	}

var (
	db *sql.DB
	templates *template.Template
)

func initDB() {
	//connStr := "postgres://postgres:password@localhost:5432/sitesdb?sslmode=disable"
	connStr := fmt.Sprintf(
    	"postgresql://%s:%s@%s:%s/%s?sslmode=%s",
    	os.Getenv("PGUSER"),
    	os.Getenv("PGPASSWORD"),
    	os.Getenv("PGHOST"),
    	os.Getenv("PGPORT"),
    	os.Getenv("PGDATABASE"),
    	os.Getenv("PGSSLMODE"),
	)
	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS sites (
		id SERIAL PRIMARY KEY,
		url TEXT UNIQUE NOT NULL,
		active BOOLEAN DEFAULT FALSE,
		last_check TIMESTAMP,
		response_time_ms BIGINT
		)`)

	if err != nil {
		log.Fatal(err)
	}
}

func checkSite(url string) (active bool, responseTime int64, err error) {
	start := time.Now()
	client := http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return false, 0, err
	}
	defer resp.Body.Close()
	elapsed := time.Since(start).Milliseconds()
	if resp.StatusCode >= 200 && resp.StatusCode < 400 {
		return true, elapsed, nil
	}
	return false, elapsed, nil
}

func addSiteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	url := r.FormValue("url")
	if url == "" {
	http.Error(w, "url is required", http.StatusBadRequest)
	return
	}

	_, err := db.Exec("INSERT INTO sites (url) VALUES ($1) ON CONFLICT DO NOTHING", url)
	if err != nil {
		http.Error(w, fmt.Sprintf("db error: %v", err), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func updateStatuses() {
	rows, err := db.Query("SELECT id, url FROM sites")
	if err != nil {
		log.Println("Error querying sites:", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
			var id int
			var url string
			err := rows.Scan(&id, &url)
			if err != nil {
				log.Println("Scan error:", err)
				continue
			}
		active, respTime, err := checkSite(url)
		if err != nil {
			active = false
			respTime = 0
		}
		_, err = db.Exec("UPDATE sites SET active=$1, last_check=NOW(), response_time_ms=$2 WHERE id=$3", active, respTime, id)
		if err != nil {
				log.Println("Update error:", err)
		}
	}
}

func dashboardHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT url, active FROM sites")
	if err != nil {
		http.Error(w, "DB error", 500)
		return
	}
	defer rows.Close()
	var sites []Site
	for rows.Next() {
		var s Site
		err := rows.Scan(&s.URL, &s.Active)
		if err != nil {
			continue
		}
		sites = append(sites, s)
	}
	templates.ExecuteTemplate(w, "dashboard.html", sites)
}

func responseTimeHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT url, response_time_ms FROM sites")
	if err != nil {
		http.Error(w, "DB error", 500)
		return
	}
	defer rows.Close()
	var sites []Site
	for rows.Next() {
		var s Site
		err := rows.Scan(&s.URL, &s.ResponseTimeMs)
		if err != nil {
			continue
		}
		sites = append(sites, s)
	}
	templates.ExecuteTemplate(w, "response_time.html", sites)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	templates.ExecuteTemplate(w, "index.html", nil)
}

func main() {
	initDB()
	templates = template.Must(template.ParseGlob("templates/*.html"))
	// Periodic update of site statuses
	go func() {
		for {
			updateStatuses()
			time.Sleep(1 * time.Minute)
		}
	}()

	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/add", addSiteHandler)
	http.HandleFunc("/dashboard", dashboardHandler)
	http.HandleFunc("/response-time", responseTimeHandler)

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	fmt.Println("Listening on :8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
