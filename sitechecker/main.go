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

	templates = template.Must(template.ParseGlob("templates/*.html"))
}

func setNoCacheHeaders(w http.ResponseWriter) {
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, proxy-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")
}

func checkSite(url string) (active bool, responseTimeMs int64) {
	start := time.Now()

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)

	elapsed := time.Since(start).Milliseconds()

	if err != nil || resp == nil || resp.StatusCode < 200 || resp.StatusCode >= 400 {
		return false, elapsed
	}
	resp.Body.Close()

	return true, elapsed
}

func startSiteChecker() {
	go func() {
		log.Println("Site checker started (running every 5 minutes)")

		for {
			rows, err := db.Query("SELECT url FROM sites")
			if err != nil {
				log.Printf("Checker DB query error: %v", err)
				time.Sleep(1 * time.Minute)
				continue
			}

			var urls []string
			for rows.Next() {
				var url string
				if err := rows.Scan(&url); err != nil {
					log.Printf("Checker row scan error: %v", err)
					continue
				}
				urls = append(urls, url)
			}
			rows.Close()

			log.Printf("Checking %d sites...", len(urls))

			for _, url := range urls {
				active, responseTime := checkSite(url)

				_, err := db.Exec(
					"UPDATE sites SET active = $1, response_time_ms = $2, last_checked = $3 WHERE url = $4",
					active, responseTime, time.Now(), url,
				)
				if err != nil {
					log.Printf("Failed to update %s: %v", url, err)
				} else {
					log.Printf("Updated %s: active=%v, time=%dms", url, active, responseTime)
				}

				time.Sleep(2 * time.Second)
			}

			log.Println("Checker cycle complete. Next check in 5 minutes.")
			time.Sleep(5 * time.Minute)
		}
	}()
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	setNoCacheHeaders(w)

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

func responseTimeHandler(w http.ResponseWriter, r *http.Request) {
	setNoCacheHeaders(w)

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

	if err := templates.ExecuteTemplate(w, "response_time.html", sites); err != nil {
		http.Error(w, "Template render failed", http.StatusInternalServerError)
		log.Printf("Template error: %v", err)
		return
	}
}

func dashboardHandler(w http.ResponseWriter, r *http.Request) {
	setNoCacheHeaders(w)

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

	if err := templates.ExecuteTemplate(w, "dashboard.html", sites); err != nil {
		http.Error(w, "Template render failed", http.StatusInternalServerError)
		log.Printf("Template error: %v", err)
		return
	}
}

func main() {
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/response-time", responseTimeHandler)
	http.HandleFunc("/dashboard", dashboardHandler)

	startSiteChecker()

	port := ":8080"
	log.Printf("Server starting on http://localhost%s", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
