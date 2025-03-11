package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http"

	_ "github.com/mattn/go-sqlite3"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type EmailTrafficCollector struct {
	db *sql.DB
}

func (c *EmailTrafficCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- prometheus.NewDesc("email_upload_bytes_total", "Total bytes uploaded by each email.", []string{"email", "enable"}, nil)
	ch <- prometheus.NewDesc("email_download_bytes_total", "Total bytes downloaded by each email.", []string{"email", "enable"}, nil)
}

func (c *EmailTrafficCollector) Collect(ch chan<- prometheus.Metric) {
	rows, err := c.db.Query("SELECT email, up, down, enable FROM client_traffics")
	if err != nil {
		log.Printf("Error querying database: %v", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var email string
		var up, down int64
		var enable int
		if err := rows.Scan(&email, &up, &down, &enable); err != nil {
			log.Printf("Error scanning row: %v", err)
			continue
		}

		enableLabel := fmt.Sprintf("%d", enable)

		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc("email_upload_bytes_total", "Total bytes uploaded by each email.", []string{"email", "enable"}, nil),
			prometheus.CounterValue,
			float64(up),
			email,
			enableLabel,
		)

		ch <- prometheus.MustNewConstMetric(
			prometheus.NewDesc("email_download_bytes_total", "Total bytes downloaded by each email.", []string{"email", "enable"}, nil),
			prometheus.CounterValue,
			float64(down),
			email,
			enableLabel,
		)
	}

	if err := rows.Err(); err != nil {
		log.Printf("Error iterating rows: %v", err)
	}
}

func main() {
	port := flag.Int("port", 9833, "Port to listen on")
	dbPath := flag.String("dbPath", "/etc/x-ui/x-ui.db", "Path to the SQLite database")
	flag.Parse()

	db, err := sql.Open("sqlite3", *dbPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	collector := &EmailTrafficCollector{db: db}

	prometheus.MustRegister(collector)

	http.Handle("/metrics", promhttp.Handler())
	log.Printf("Starting server on port %d", *port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *port), nil))
}
