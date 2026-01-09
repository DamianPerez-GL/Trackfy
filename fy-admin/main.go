package main

import (
	"bufio"
	"compress/gzip"
	"context"
	"database/sql"
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	_ "github.com/lib/pq"
)

//go:embed static/*
var staticFiles embed.FS

type Config struct {
	Port        string
	DatabaseURL string
	DBSyncURL   string
	AnalysisURL string
}

type Server struct {
	db          *sql.DB
	config      *Config
	client      *http.Client
	syncStatus  map[string]*SyncProgress
	syncMutex   sync.RWMutex
}

// SyncProgress rastrea el progreso de una sincronización
type SyncProgress struct {
	Source     string    `json:"source"`
	InProgress bool      `json:"in_progress"`
	StartedAt  time.Time `json:"started_at,omitempty"`
	Records    int64     `json:"records"`
	Errors     int64     `json:"errors"`
	Message    string    `json:"message"`
}

func main() {
	config := &Config{
		Port:        getEnv("PORT", "9092"),
		DatabaseURL: getEnv("DATABASE_URL", ""),
		DBSyncURL:   getEnv("DBSYNC_URL", "http://fy-dbsync:9091"),
		AnalysisURL: getEnv("ANALYSIS_URL", "http://fy-analysis:9090"),
	}

	var db *sql.DB
	if config.DatabaseURL != "" {
		var err error
		db, err = sql.Open("postgres", config.DatabaseURL)
		if err != nil {
			fmt.Printf("Warning: Failed to connect to database: %v\n", err)
		} else {
			db.SetMaxOpenConns(5)
			db.SetMaxIdleConns(2)
		}
	}

	server := &Server{
		db:     db,
		config: config,
		client: &http.Client{Timeout: 30 * time.Second},
		syncStatus: map[string]*SyncProgress{
			"urlhaus":   {Source: "urlhaus"},
			"openphish": {Source: "openphish"},
			"emails":    {Source: "emails"},
			"phones":    {Source: "phones"},
		},
	}

	mux := http.NewServeMux()

	// API endpoints
	mux.HandleFunc("/api/health", server.handleHealth)
	mux.HandleFunc("/api/stats/database", server.handleDatabaseStats)
	mux.HandleFunc("/api/stats/sources", server.handleSourcesStats)
	mux.HandleFunc("/api/stats/sync", server.handleSyncStatus)
	mux.HandleFunc("/api/actions/sync", server.handleForceSync)
	mux.HandleFunc("/api/actions/sync/progress", server.handleSyncProgress)
	mux.HandleFunc("/api/services/status", server.handleServicesStatus)

	// Data listing endpoints
	mux.HandleFunc("/api/data/domains", server.handleListDomains)
	mux.HandleFunc("/api/data/emails", server.handleListEmails)
	mux.HandleFunc("/api/data/phones", server.handleListPhones)
	mux.HandleFunc("/api/data/whitelist", server.handleListWhitelist)
	mux.HandleFunc("/api/data/reports", server.handleListReports)
	mux.HandleFunc("/api/data/reports/stats", server.handleReportsStats)

	// Manual entry endpoints
	mux.HandleFunc("/api/add/phone", server.handleAddPhone)
	mux.HandleFunc("/api/add/email", server.handleAddEmail)

	// Static files
	staticFS, _ := fs.Sub(staticFiles, "static")
	mux.Handle("/", http.FileServer(http.FS(staticFS)))

	httpServer := &http.Server{
		Addr:         ":" + config.Port,
		Handler:      corsMiddleware(mux),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	go func() {
		fmt.Printf("Fy Admin Panel running on http://localhost:%s\n", config.Port)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("Server error: %v\n", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	httpServer.Shutdown(ctx)

	if db != nil {
		db.Close()
	}
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
}

func (s *Server) handleDatabaseStats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if s.db == nil {
		json.NewEncoder(w).Encode(map[string]string{"error": "Database not connected"})
		return
	}

	stats := make(map[string]interface{})

	var domainsTotal, domainsActive int64
	s.db.QueryRow(`SELECT COUNT(*), COUNT(*) FILTER (WHERE (flags & 1) = 1) FROM threat_domains`).Scan(&domainsTotal, &domainsActive)
	stats["domains"] = map[string]int64{"total": domainsTotal, "active": domainsActive}

	var pathsTotal, pathsActive int64
	s.db.QueryRow(`SELECT COUNT(*), COUNT(*) FILTER (WHERE (flags & 1) = 1) FROM threat_paths`).Scan(&pathsTotal, &pathsActive)
	stats["paths"] = map[string]int64{"total": pathsTotal, "active": pathsActive}

	var emailsTotal int64
	s.db.QueryRow(`SELECT COUNT(*) FROM threat_emails`).Scan(&emailsTotal)
	stats["emails"] = map[string]int64{"total": emailsTotal}

	// Estadísticas de emails por fuente
	emailSourceStats := []map[string]interface{}{}
	emailRows, err := s.db.Query(`
		SELECT source::text, COUNT(*) as count
		FROM threat_emails
		GROUP BY source
		ORDER BY count DESC
	`)
	if err == nil {
		defer emailRows.Close()
		for emailRows.Next() {
			var source string
			var count int64
			if emailRows.Scan(&source, &count) == nil {
				emailSourceStats = append(emailSourceStats, map[string]interface{}{
					"source": source,
					"count":  count,
				})
			}
		}
	}
	stats["emails_by_source"] = emailSourceStats

	var phonesTotal int64
	s.db.QueryRow(`SELECT COUNT(*) FROM threat_phones`).Scan(&phonesTotal)
	stats["phones"] = map[string]int64{"total": phonesTotal}

	var whitelistTotal int64
	s.db.QueryRow(`SELECT COUNT(*) FROM whitelist_domains`).Scan(&whitelistTotal)
	stats["whitelist"] = map[string]int64{"total": whitelistTotal}

	sourceStats := []map[string]interface{}{}
	rows, err := s.db.Query(`
		SELECT source::text, COUNT(*) as count
		FROM threat_domains
		GROUP BY source
		ORDER BY count DESC
	`)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var source string
			var count int64
			if rows.Scan(&source, &count) == nil {
				sourceStats = append(sourceStats, map[string]interface{}{
					"source": source,
					"count":  count,
				})
			}
		}
	}
	stats["by_source"] = sourceStats

	threatStats := []map[string]interface{}{}
	rows2, err := s.db.Query(`
		SELECT threat_type::text, COUNT(*) as count
		FROM threat_domains
		WHERE (flags & 1) = 1
		GROUP BY threat_type
		ORDER BY count DESC
	`)
	if err == nil {
		defer rows2.Close()
		for rows2.Next() {
			var threatType string
			var count int64
			if rows2.Scan(&threatType, &count) == nil {
				threatStats = append(threatStats, map[string]interface{}{
					"type":  threatType,
					"count": count,
				})
			}
		}
	}
	stats["by_threat_type"] = threatStats

	json.NewEncoder(w).Encode(stats)
}

func (s *Server) handleSourcesStats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if s.db == nil {
		json.NewEncoder(w).Encode(map[string]string{"error": "Database not connected"})
		return
	}

	// Intervalos de sincronización por defecto (deben coincidir con fy-dbsync)
	// Nota: OpenPhish usa 'phishtank' en el enum de la BD
	syncIntervals := map[string]time.Duration{
		"urlhaus":   5 * time.Minute,
		"phishtank": 1 * time.Hour,  // OpenPhish usa 'phishtank' en sync_status
		"osint":     24 * time.Hour, // StopForumSpam
	}

	sources := []map[string]interface{}{}

	rows, err := s.db.Query(`
		SELECT source::text, last_sync, last_count, last_error
		FROM sync_status
		ORDER BY source
	`)
	if err == nil {
		defer rows.Close()
		now := time.Now()
		for rows.Next() {
			var source string
			var lastSync time.Time
			var lastCount int64
			var lastError sql.NullString

			if rows.Scan(&source, &lastSync, &lastCount, &lastError) == nil {
				// Mapear nombre para frontend (phishtank -> openphish)
				displayName := source
				if source == "phishtank" {
					displayName = "openphish"
				}

				src := map[string]interface{}{
					"name":       displayName,
					"last_sync":  lastSync.Format(time.RFC3339),
					"last_count": lastCount,
					"status":     "ok",
				}
				if lastError.Valid && lastError.String != "" {
					src["status"] = "error"
					src["error"] = lastError.String
				}

				// Calcular tiempo restante para próxima sincronización
				if interval, ok := syncIntervals[source]; ok {
					nextSync := lastSync.Add(interval)
					remaining := nextSync.Sub(now)
					if remaining < 0 {
						remaining = 0 // Ya debería sincronizarse
					}
					src["interval_seconds"] = int64(interval.Seconds())
					src["next_sync"] = nextSync.Format(time.RFC3339)
					src["remaining_seconds"] = int64(remaining.Seconds())
				}

				sources = append(sources, src)
			}
		}
	}

	json.NewEncoder(w).Encode(map[string]interface{}{"sources": sources})
}

func (s *Server) handleSyncStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	resp, err := s.client.Get(s.config.DBSyncURL + "/status")
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "offline",
			"error":  err.Error(),
		})
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var status map[string]interface{}
	json.Unmarshal(body, &status)
	status["status"] = "online"

	json.NewEncoder(w).Encode(status)
}

func (s *Server) handleForceSync(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if s.db == nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Database not connected",
		})
		return
	}

	source := r.URL.Query().Get("source")
	if source == "" {
		source = "all"
	}

	// Verificar si ya hay un sync en progreso para esta fuente
	s.syncMutex.RLock()
	if source != "all" {
		if status, ok := s.syncStatus[source]; ok && status.InProgress {
			s.syncMutex.RUnlock()
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error":   "Sync already in progress for: " + source,
			})
			return
		}
	}
	s.syncMutex.RUnlock()

	// Iniciar sync en background
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
		defer cancel()

		switch source {
		case "urlhaus":
			s.syncURLhaus(ctx)
		case "openphish":
			s.syncOpenPhish(ctx)
		case "emails", "stopforumspam":
			s.syncStopForumSpam(ctx)
		case "phones":
			s.syncPhones(ctx)
		case "all":
			// Sincronizar todas las fuentes en paralelo
			var wg sync.WaitGroup
			wg.Add(4)
			go func() { defer wg.Done(); s.syncURLhaus(ctx) }()
			go func() { defer wg.Done(); s.syncOpenPhish(ctx) }()
			go func() { defer wg.Done(); s.syncStopForumSpam(ctx) }()
			go func() { defer wg.Done(); s.syncPhones(ctx) }()
			wg.Wait()
		}
	}()

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Sync started for: " + source,
	})
}

func (s *Server) handleSyncProgress(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	s.syncMutex.RLock()
	defer s.syncMutex.RUnlock()

	json.NewEncoder(w).Encode(s.syncStatus)
}

func (s *Server) handleServicesStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	services := []map[string]interface{}{}

	dbsyncStatus := s.checkService(s.config.DBSyncURL + "/health")
	services = append(services, map[string]interface{}{
		"name":   "fy-dbsync",
		"url":    s.config.DBSyncURL,
		"status": dbsyncStatus,
	})

	analysisStatus := s.checkService(s.config.AnalysisURL + "/health")
	services = append(services, map[string]interface{}{
		"name":   "fy-analysis",
		"url":    s.config.AnalysisURL,
		"status": analysisStatus,
	})

	dbStatus := "offline"
	if s.db != nil {
		if err := s.db.Ping(); err == nil {
			dbStatus = "online"
		}
	}
	services = append(services, map[string]interface{}{
		"name":   "PostgreSQL",
		"status": dbStatus,
	})

	json.NewEncoder(w).Encode(map[string]interface{}{"services": services})
}

func (s *Server) handleListDomains(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if s.db == nil {
		json.NewEncoder(w).Encode(map[string]string{"error": "Database not connected"})
		return
	}

	limit := getQueryInt(r, "limit", 50)
	offset := getQueryInt(r, "offset", 0)
	search := r.URL.Query().Get("search")
	source := r.URL.Query().Get("source")
	threatType := r.URL.Query().Get("threat_type")

	query := `
		SELECT domain, threat_type::text, severity::text, confidence, source::text,
		       first_seen, last_seen, hit_count
		FROM threat_domains
		WHERE (flags & 1) = 1
	`
	args := []interface{}{}
	argCount := 0

	if search != "" {
		argCount++
		query += fmt.Sprintf(" AND domain ILIKE $%d", argCount)
		args = append(args, "%"+search+"%")
	}
	if source != "" {
		argCount++
		query += fmt.Sprintf(" AND source::text = $%d", argCount)
		args = append(args, source)
	}
	if threatType != "" {
		argCount++
		query += fmt.Sprintf(" AND threat_type::text = $%d", argCount)
		args = append(args, threatType)
	}

	query += " ORDER BY last_seen DESC"
	query += fmt.Sprintf(" LIMIT %d OFFSET %d", limit, offset)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	defer rows.Close()

	domains := []map[string]interface{}{}
	for rows.Next() {
		var domain, threatType, severity, source string
		var confidence int
		var firstSeen, lastSeen time.Time
		var hitCount int

		if rows.Scan(&domain, &threatType, &severity, &confidence, &source, &firstSeen, &lastSeen, &hitCount) == nil {
			domains = append(domains, map[string]interface{}{
				"domain":      domain,
				"threat_type": threatType,
				"severity":    severity,
				"confidence":  confidence,
				"source":      source,
				"first_seen":  firstSeen.Format(time.RFC3339),
				"last_seen":   lastSeen.Format(time.RFC3339),
				"hit_count":   hitCount,
			})
		}
	}

	// Get total count
	countQuery := `SELECT COUNT(*) FROM threat_domains WHERE (flags & 1) = 1`
	var total int64
	s.db.QueryRow(countQuery).Scan(&total)

	json.NewEncoder(w).Encode(map[string]interface{}{
		"data":   domains,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

func (s *Server) handleListEmails(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if s.db == nil {
		json.NewEncoder(w).Encode(map[string]string{"error": "Database not connected"})
		return
	}

	limit := getQueryInt(r, "limit", 50)
	offset := getQueryInt(r, "offset", 0)
	search := r.URL.Query().Get("search")
	source := r.URL.Query().Get("source")
	threatType := r.URL.Query().Get("threat_type")

	query := `
		SELECT email, threat_type::text, severity::text, confidence, source::text,
		       impersonates, first_seen, last_seen, report_count
		FROM threat_emails
		WHERE (flags & 1) = 1
	`
	args := []interface{}{}
	argCount := 0

	if search != "" {
		argCount++
		query += fmt.Sprintf(" AND email ILIKE $%d", argCount)
		args = append(args, "%"+search+"%")
	}
	if source != "" {
		argCount++
		query += fmt.Sprintf(" AND source::text = $%d", argCount)
		args = append(args, source)
	}
	if threatType != "" {
		argCount++
		query += fmt.Sprintf(" AND threat_type::text = $%d", argCount)
		args = append(args, threatType)
	}

	query += " ORDER BY last_seen DESC"
	query += fmt.Sprintf(" LIMIT %d OFFSET %d", limit, offset)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	defer rows.Close()

	emails := []map[string]interface{}{}
	for rows.Next() {
		var email, threatType, severity, source string
		var confidence, reportCount int
		var impersonates sql.NullString
		var firstSeen, lastSeen time.Time

		if rows.Scan(&email, &threatType, &severity, &confidence, &source, &impersonates, &firstSeen, &lastSeen, &reportCount) == nil {
			item := map[string]interface{}{
				"email":        email,
				"threat_type":  threatType,
				"severity":     severity,
				"confidence":   confidence,
				"source":       source,
				"first_seen":   firstSeen.Format(time.RFC3339),
				"last_seen":    lastSeen.Format(time.RFC3339),
				"report_count": reportCount,
			}
			if impersonates.Valid {
				item["impersonates"] = impersonates.String
			}
			emails = append(emails, item)
		}
	}

	// Get total count with same filters
	countQuery := `SELECT COUNT(*) FROM threat_emails WHERE (flags & 1) = 1`
	countArgs := []interface{}{}
	argCount = 0
	if search != "" {
		argCount++
		countQuery += fmt.Sprintf(" AND email ILIKE $%d", argCount)
		countArgs = append(countArgs, "%"+search+"%")
	}
	if source != "" {
		argCount++
		countQuery += fmt.Sprintf(" AND source::text = $%d", argCount)
		countArgs = append(countArgs, source)
	}
	if threatType != "" {
		argCount++
		countQuery += fmt.Sprintf(" AND threat_type::text = $%d", argCount)
		countArgs = append(countArgs, threatType)
	}

	var total int64
	s.db.QueryRow(countQuery, countArgs...).Scan(&total)

	json.NewEncoder(w).Encode(map[string]interface{}{
		"data":   emails,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

func (s *Server) handleListPhones(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if s.db == nil {
		json.NewEncoder(w).Encode(map[string]string{"error": "Database not connected"})
		return
	}

	limit := getQueryInt(r, "limit", 50)
	offset := getQueryInt(r, "offset", 0)
	search := r.URL.Query().Get("search")

	query := `
		SELECT phone_national, country_code, threat_type::text, severity::text,
		       confidence, source::text, description, first_seen, last_seen
		FROM threat_phones
		WHERE (flags & 1) = 1
	`
	args := []interface{}{}

	if search != "" {
		query += " AND phone_national LIKE $1"
		args = append(args, "%"+search+"%")
	}

	query += " ORDER BY last_seen DESC"
	query += fmt.Sprintf(" LIMIT %d OFFSET %d", limit, offset)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	defer rows.Close()

	phones := []map[string]interface{}{}
	for rows.Next() {
		var phone, countryCode, threatType, severity, source string
		var confidence int
		var description sql.NullString
		var firstSeen, lastSeen time.Time

		if rows.Scan(&phone, &countryCode, &threatType, &severity, &confidence, &source, &description, &firstSeen, &lastSeen) == nil {
			item := map[string]interface{}{
				"phone":        phone,
				"country_code": countryCode,
				"threat_type":  threatType,
				"severity":     severity,
				"confidence":   confidence,
				"source":       source,
				"first_seen":   firstSeen.Format(time.RFC3339),
				"last_seen":    lastSeen.Format(time.RFC3339),
			}
			if description.Valid {
				item["description"] = description.String
			}
			phones = append(phones, item)
		}
	}

	var total int64
	s.db.QueryRow(`SELECT COUNT(*) FROM threat_phones WHERE (flags & 1) = 1`).Scan(&total)

	json.NewEncoder(w).Encode(map[string]interface{}{
		"data":   phones,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

func (s *Server) handleListWhitelist(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if s.db == nil {
		json.NewEncoder(w).Encode(map[string]string{"error": "Database not connected"})
		return
	}

	limit := getQueryInt(r, "limit", 50)
	offset := getQueryInt(r, "offset", 0)
	search := r.URL.Query().Get("search")

	query := `
		SELECT domain, category, brand, country, official_name, created_at
		FROM whitelist_domains
		WHERE 1=1
	`
	args := []interface{}{}

	if search != "" {
		query += " AND (domain ILIKE $1 OR brand ILIKE $1)"
		args = append(args, "%"+search+"%")
	}

	query += " ORDER BY brand, domain"
	query += fmt.Sprintf(" LIMIT %d OFFSET %d", limit, offset)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	defer rows.Close()

	whitelist := []map[string]interface{}{}
	for rows.Next() {
		var domain string
		var category, brand, country, officialName sql.NullString
		var createdAt time.Time

		if rows.Scan(&domain, &category, &brand, &country, &officialName, &createdAt) == nil {
			item := map[string]interface{}{
				"domain":     domain,
				"created_at": createdAt.Format(time.RFC3339),
			}
			if category.Valid {
				item["category"] = category.String
			}
			if brand.Valid {
				item["brand"] = brand.String
			}
			if country.Valid {
				item["country"] = country.String
			}
			if officialName.Valid {
				item["official_name"] = officialName.String
			}
			whitelist = append(whitelist, item)
		}
	}

	var total int64
	s.db.QueryRow(`SELECT COUNT(*) FROM whitelist_domains`).Scan(&total)

	json.NewEncoder(w).Encode(map[string]interface{}{
		"data":   whitelist,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

func (s *Server) handleAddPhone(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if s.db == nil {
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "Database not connected"})
		return
	}

	var input struct {
		Phone       string `json:"phone"`
		CountryCode string `json:"country_code"`
		ThreatType  string `json:"threat_type"`
		Severity    string `json:"severity"`
		Description string `json:"description"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "Invalid JSON"})
		return
	}

	if input.Phone == "" || input.CountryCode == "" {
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "Phone and country_code are required"})
		return
	}

	// Valores por defecto
	if input.ThreatType == "" {
		input.ThreatType = "scam"
	}
	if input.Severity == "" {
		input.Severity = "medium"
	}

	now := time.Now()
	_, err := s.db.Exec(`
		INSERT INTO threat_phones (phone_national, country_code, threat_type, severity, confidence, source, description, first_seen, last_seen, flags)
		VALUES ($1, $2, $3::threat_type_enum, $4::severity_enum, 80, 'manual'::source_enum, $5, $6, $7, 1)
		ON CONFLICT (phone_national) DO UPDATE SET
			last_seen = EXCLUDED.last_seen,
			report_count = threat_phones.report_count + 1
	`, input.Phone, input.CountryCode, input.ThreatType, input.Severity, input.Description, now, now)

	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": err.Error()})
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "message": "Phone added successfully"})
}

func (s *Server) handleAddEmail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if s.db == nil {
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "Database not connected"})
		return
	}

	var input struct {
		Email       string `json:"email"`
		ThreatType  string `json:"threat_type"`
		Severity    string `json:"severity"`
		Impersonates string `json:"impersonates"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "Invalid JSON"})
		return
	}

	if input.Email == "" {
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "Email is required"})
		return
	}

	// Valores por defecto
	if input.ThreatType == "" {
		input.ThreatType = "phishing"
	}
	if input.Severity == "" {
		input.Severity = "medium"
	}

	// Extraer dominio del email
	domain := ""
	if idx := strings.LastIndex(input.Email, "@"); idx > 0 {
		domain = input.Email[idx+1:]
	}

	now := time.Now()
	_, err := s.db.Exec(`
		INSERT INTO threat_emails (email_hash, email, domain_hash, threat_type, severity, confidence, source, impersonates, first_seen, last_seen, flags)
		VALUES (sha256_bytea($1), $1, sha256_bytea($2), $3::threat_type_enum, $4::severity_enum, 80, 'manual'::source_enum, $5, $6, $7, 1)
		ON CONFLICT (email_hash) DO UPDATE SET
			last_seen = EXCLUDED.last_seen,
			report_count = threat_emails.report_count + 1
	`, input.Email, domain, input.ThreatType, input.Severity, input.Impersonates, now, now)

	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": err.Error()})
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "message": "Email added successfully"})
}

func (s *Server) checkService(url string) string {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
	resp, err := s.client.Do(req)
	if err != nil {
		return "offline"
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return "online"
	}
	return "error"
}

func getEnv(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}

func getQueryInt(r *http.Request, key string, defaultValue int) int {
	val := r.URL.Query().Get(key)
	if val == "" {
		return defaultValue
	}
	i, err := strconv.Atoi(val)
	if err != nil {
		return defaultValue
	}
	return i
}

// ============================================================================
// SYNC FUNCTIONS - Importación directa de fuentes de amenazas
// ============================================================================

const (
	urlhausDownloadURL      = "https://urlhaus.abuse.ch/downloads/text/"
	openPhishURL            = "https://openphish.com/feed.txt"
	stopForumSpamEmailsURL  = "https://www.stopforumspam.com/downloads/listed_email_365_all.gz"
	listaHuPhonesURL        = "https://listahu.org/descargar/csv"
)

// syncURLhaus descarga e importa datos de URLhaus
func (s *Server) syncURLhaus(ctx context.Context) {
	source := "urlhaus"
	s.updateSyncStatus(source, true, "Downloading URLhaus feed...")

	startTime := time.Now()
	var records, errors int64

	defer func() {
		duration := time.Since(startTime)
		s.updateSyncStatusComplete(source, records, errors, fmt.Sprintf("Completed in %v", duration.Round(time.Second)))
	}()

	req, err := http.NewRequestWithContext(ctx, "GET", urlhausDownloadURL, nil)
	if err != nil {
		s.updateSyncStatusComplete(source, 0, 1, "Failed to create request: "+err.Error())
		return
	}
	req.Header.Set("User-Agent", "Fy-Admin/1.0")

	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		s.updateSyncStatusComplete(source, 0, 1, "Failed to download: "+err.Error())
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		s.updateSyncStatusComplete(source, 0, 1, fmt.Sprintf("Download failed with status: %d", resp.StatusCode))
		return
	}

	s.updateSyncStatus(source, true, "Parsing and importing...")

	scanner := bufio.NewScanner(resp.Body)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	lineNum := 0
	now := time.Now()

	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return
		default:
		}

		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		lineNum++

		parsedURL, err := url.Parse(line)
		if err != nil {
			errors++
			continue
		}

		domain := strings.ToLower(parsedURL.Hostname())
		if domain == "" || len(domain) < 3 || !strings.Contains(domain, ".") {
			errors++
			continue
		}

		// Remover puerto
		if idx := strings.Index(domain, ":"); idx != -1 {
			domain = domain[:idx]
		}

		// Saltar IPs
		if net.ParseIP(domain) != nil {
			continue
		}

		tld := extractTLD(domain)
		sourceID := fmt.Sprintf("urlhaus-%d", lineNum)

		_, err = s.db.ExecContext(ctx, `
			INSERT INTO threat_domains (domain_hash, domain, threat_type, severity, confidence, source, source_id, tld, first_seen, last_seen, flags)
			VALUES (sha256_bytea($1), $1, 'malware'::threat_type_enum, 'high'::severity_enum, 85, 'urlhaus'::source_enum, $2, $3, $4, $5, 1)
			ON CONFLICT (domain_hash) DO UPDATE SET
				last_seen = EXCLUDED.last_seen,
				hit_count = threat_domains.hit_count + 1,
				confidence = GREATEST(threat_domains.confidence, EXCLUDED.confidence)
		`, domain, sourceID, tld, now, now)

		if err != nil {
			errors++
			continue
		}
		records++

		// Insertar path si existe
		path := parsedURL.Path
		if path != "" && path != "/" {
			fullPath := domain + path
			s.db.ExecContext(ctx, `
				INSERT INTO threat_paths (path_hash, domain_hash, path, threat_type, severity, confidence, source, first_seen, last_seen, flags)
				VALUES (sha256_bytea($1), sha256_bytea($2), $3, 'malware'::threat_type_enum, 'high'::severity_enum, 85, 'urlhaus'::source_enum, $4, $5, 1)
				ON CONFLICT (path_hash) DO UPDATE SET last_seen = EXCLUDED.last_seen
			`, fullPath, domain, path, now, now)
		}

		if lineNum%1000 == 0 {
			s.updateSyncStatus(source, true, fmt.Sprintf("Imported %d domains...", records))
		}
	}

	// Actualizar sync_status en BD
	s.db.ExecContext(ctx, `
		INSERT INTO sync_status (source, last_sync, last_count)
		VALUES ('urlhaus'::source_enum, NOW(), $1)
		ON CONFLICT (source) DO UPDATE SET
			last_sync = NOW(),
			last_count = $1
	`, records)
}

// syncOpenPhish descarga e importa datos de OpenPhish
func (s *Server) syncOpenPhish(ctx context.Context) {
	source := "openphish"
	s.updateSyncStatus(source, true, "Downloading OpenPhish feed...")

	startTime := time.Now()
	var records, errors int64

	defer func() {
		duration := time.Since(startTime)
		s.updateSyncStatusComplete(source, records, errors, fmt.Sprintf("Completed in %v", duration.Round(time.Second)))
	}()

	req, err := http.NewRequestWithContext(ctx, "GET", openPhishURL, nil)
	if err != nil {
		s.updateSyncStatusComplete(source, 0, 1, "Failed to create request: "+err.Error())
		return
	}
	req.Header.Set("User-Agent", "Fy-Admin/1.0")

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		s.updateSyncStatusComplete(source, 0, 1, "Failed to download: "+err.Error())
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		s.updateSyncStatusComplete(source, 0, 1, fmt.Sprintf("Download failed with status: %d", resp.StatusCode))
		return
	}

	s.updateSyncStatus(source, true, "Parsing and importing...")

	scanner := bufio.NewScanner(resp.Body)
	lineNum := 0
	now := time.Now()

	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return
		default:
		}

		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		lineNum++

		parsedURL, err := url.Parse(line)
		if err != nil {
			errors++
			continue
		}

		domain := strings.ToLower(parsedURL.Hostname())
		if domain == "" || len(domain) < 3 || !strings.Contains(domain, ".") {
			errors++
			continue
		}

		// Saltar IPs
		if net.ParseIP(domain) != nil {
			continue
		}

		tld := extractTLD(domain)
		sourceID := fmt.Sprintf("openphish-%d", lineNum)

		_, err = s.db.ExecContext(ctx, `
			INSERT INTO threat_domains (domain_hash, domain, threat_type, severity, confidence, source, source_id, tld, first_seen, last_seen, flags)
			VALUES (sha256_bytea($1), $1, 'phishing'::threat_type_enum, 'high'::severity_enum, 90, 'phishtank'::source_enum, $2, $3, $4, $5, 1)
			ON CONFLICT (domain_hash) DO UPDATE SET
				last_seen = EXCLUDED.last_seen,
				hit_count = threat_domains.hit_count + 1,
				confidence = GREATEST(threat_domains.confidence, EXCLUDED.confidence)
		`, domain, sourceID, tld, now, now)

		if err != nil {
			errors++
			continue
		}
		records++

		// Insertar path si existe
		path := parsedURL.Path
		if path != "" && path != "/" {
			fullPath := domain + path
			s.db.ExecContext(ctx, `
				INSERT INTO threat_paths (path_hash, domain_hash, path, threat_type, severity, confidence, source, first_seen, last_seen, flags)
				VALUES (sha256_bytea($1), sha256_bytea($2), $3, 'phishing'::threat_type_enum, 'high'::severity_enum, 90, 'phishtank'::source_enum, $4, $5, 1)
				ON CONFLICT (path_hash) DO UPDATE SET last_seen = EXCLUDED.last_seen
			`, fullPath, domain, path, now, now)
		}
	}

	// Actualizar sync_status en BD (usa 'phishtank' como en el enum)
	s.db.ExecContext(ctx, `
		INSERT INTO sync_status (source, last_sync, last_count)
		VALUES ('phishtank'::source_enum, NOW(), $1)
		ON CONFLICT (source) DO UPDATE SET
			last_sync = NOW(),
			last_count = $1
	`, records)
}

// cleanEmail limpia un email de caracteres no deseados
func cleanEmail(email string) string {
	// Quitar comillas simples y dobles
	email = strings.Trim(email, "\"'`")

	// Quitar espacios al inicio y final
	email = strings.TrimSpace(email)

	// Quitar caracteres de control y no imprimibles
	var cleaned strings.Builder
	for _, r := range email {
		if r >= 32 && r < 127 { // Solo ASCII imprimible
			cleaned.WriteRune(r)
		}
	}

	return cleaned.String()
}

// syncStopForumSpam descarga e importa emails de spam
func (s *Server) syncStopForumSpam(ctx context.Context) {
	source := "emails"
	s.updateSyncStatus(source, true, "Downloading StopForumSpam emails...")

	startTime := time.Now()
	var records, errors int64

	defer func() {
		duration := time.Since(startTime)
		s.updateSyncStatusComplete(source, records, errors, fmt.Sprintf("Completed in %v", duration.Round(time.Second)))
	}()

	req, err := http.NewRequestWithContext(ctx, "GET", stopForumSpamEmailsURL, nil)
	if err != nil {
		s.updateSyncStatusComplete(source, 0, 1, "Failed to create request: "+err.Error())
		return
	}
	req.Header.Set("User-Agent", "Fy-Admin/1.0")

	client := &http.Client{Timeout: 180 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		s.updateSyncStatusComplete(source, 0, 1, "Failed to download: "+err.Error())
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		s.updateSyncStatusComplete(source, 0, 1, fmt.Sprintf("Download failed with status: %d", resp.StatusCode))
		return
	}

	// Descomprimir gzip
	gzReader, err := gzip.NewReader(resp.Body)
	if err != nil {
		s.updateSyncStatusComplete(source, 0, 1, "Failed to decompress: "+err.Error())
		return
	}
	defer gzReader.Close()

	s.updateSyncStatus(source, true, "Parsing and importing emails...")

	scanner := bufio.NewScanner(gzReader)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	lineNum := 0
	now := time.Now()

	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return
		default:
		}

		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		// Formato: email,count,lastseen
		parts := strings.Split(line, ",")
		if len(parts) < 1 {
			continue
		}

		email := strings.ToLower(strings.TrimSpace(parts[0]))

		// Limpiar el email de caracteres no deseados
		email = cleanEmail(email)

		if email == "" || !strings.Contains(email, "@") {
			continue
		}

		// Validar formato básico de email
		if len(email) < 5 || len(email) > 254 {
			continue
		}

		// Verificar que no contenga datos corruptos
		if strings.Contains(email, "%") || strings.Contains(email, "spinfile") ||
			strings.Contains(email, " ") || strings.Contains(email, "\t") {
			continue
		}

		lineNum++

		// Extraer dominio del email
		atIndex := strings.LastIndex(email, "@")
		if atIndex < 1 {
			errors++
			continue
		}
		domain := email[atIndex+1:]

		// Validar que el dominio tenga al menos un punto
		if !strings.Contains(domain, ".") {
			errors++
			continue
		}

		// Calcular confianza basado en el count si está disponible
		confidence := int16(70)
		if len(parts) >= 2 && parts[1] != "" {
			var count int
			fmt.Sscanf(parts[1], "%d", &count)
			if count > 100 {
				confidence = 95
			} else if count > 50 {
				confidence = 90
			} else if count > 10 {
				confidence = 80
			}
		}

		_, err := s.db.ExecContext(ctx, `
			INSERT INTO threat_emails (email_hash, email, domain_hash, threat_type, severity, confidence, source, first_seen, last_seen, flags)
			VALUES (sha256_bytea($1), $1, sha256_bytea($2), 'spam'::threat_type_enum, 'medium'::severity_enum, $3, 'osint'::source_enum, $4, $5, 1)
			ON CONFLICT (email_hash) DO UPDATE SET
				last_seen = EXCLUDED.last_seen,
				report_count = threat_emails.report_count + 1,
				confidence = GREATEST(threat_emails.confidence, EXCLUDED.confidence)
		`, email, domain, confidence, now, now)

		if err != nil {
			errors++
			continue
		}
		records++

		if lineNum%10000 == 0 {
			s.updateSyncStatus(source, true, fmt.Sprintf("Imported %d emails...", records))
		}
	}

	// Actualizar sync_status en BD (usa 'osint' para StopForumSpam)
	s.db.ExecContext(ctx, `
		INSERT INTO sync_status (source, last_sync, last_count)
		VALUES ('osint'::source_enum, NOW(), $1)
		ON CONFLICT (source) DO UPDATE SET
			last_sync = NOW(),
			last_count = $1
	`, records)
}

// syncPhones descarga e importa números de teléfono de estafa desde Lista Hũ
func (s *Server) syncPhones(ctx context.Context) {
	source := "phones"
	s.updateSyncStatus(source, true, "Downloading Lista Hũ phone database...")

	startTime := time.Now()
	var records, errors int64

	defer func() {
		duration := time.Since(startTime)
		s.updateSyncStatusComplete(source, records, errors, fmt.Sprintf("Completed in %v", duration.Round(time.Second)))
	}()

	req, err := http.NewRequestWithContext(ctx, "GET", listaHuPhonesURL, nil)
	if err != nil {
		s.updateSyncStatusComplete(source, 0, 1, "Failed to create request: "+err.Error())
		return
	}
	req.Header.Set("User-Agent", "Fy-Admin/1.0")

	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		s.updateSyncStatusComplete(source, 0, 1, "Failed to download: "+err.Error())
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		s.updateSyncStatusComplete(source, 0, 1, fmt.Sprintf("Download failed with status: %d", resp.StatusCode))
		return
	}

	s.updateSyncStatus(source, true, "Parsing and importing phones...")

	scanner := bufio.NewScanner(resp.Body)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	lineNum := 0
	now := time.Now()

	// Saltar cabecera
	if scanner.Scan() {
		// Primera línea es header: "#","Numero","Tipo","Comentarios","Captura","Fecha_Denuncia"
	}

	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return
		default:
		}

		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		lineNum++

		// Formato CSV: "#","Numero","Tipo","Comentarios","Captura","Fecha_Denuncia"
		// Parsear CSV manualmente (campos entre comillas)
		parts := parseCSVLine(line)
		if len(parts) < 3 {
			errors++
			continue
		}

		phone := strings.TrimSpace(parts[1])
		tipoRaw := strings.ToLower(strings.TrimSpace(parts[2]))

		// Validar número
		if phone == "" || len(phone) < 6 {
			errors++
			continue
		}

		// Determinar código de país y número nacional
		countryCode := "XX"
		phoneNational := phone

		// Detectar prefijos conocidos
		if strings.HasPrefix(phone, "34") && len(phone) >= 11 {
			countryCode = "ES"
			phoneNational = phone[2:] // Remover 34
		} else if strings.HasPrefix(phone, "595") && len(phone) >= 12 {
			countryCode = "PY" // Paraguay
			phoneNational = phone[3:]
		} else if strings.HasPrefix(phone, "54") && len(phone) >= 11 {
			countryCode = "AR" // Argentina
			phoneNational = phone[2:]
		} else if strings.HasPrefix(phone, "52") && len(phone) >= 11 {
			countryCode = "MX" // México
			phoneNational = phone[2:]
		} else if strings.HasPrefix(phone, "57") && len(phone) >= 11 {
			countryCode = "CO" // Colombia
			phoneNational = phone[2:]
		} else if strings.HasPrefix(phone, "56") && len(phone) >= 11 {
			countryCode = "CL" // Chile
			phoneNational = phone[2:]
		} else if strings.HasPrefix(phone, "51") && len(phone) >= 11 {
			countryCode = "PE" // Perú
			phoneNational = phone[2:]
		}

		// Mapear tipo de amenaza
		threatType := "scam"
		severity := "medium"
		switch {
		case strings.Contains(tipoRaw, "estafa"):
			threatType = "scam"
			severity = "high"
		case strings.Contains(tipoRaw, "extorsion") || strings.Contains(tipoRaw, "extorsión"):
			threatType = "scam"
			severity = "critical"
		case strings.Contains(tipoRaw, "spam"):
			threatType = "spam"
			severity = "low"
		case strings.Contains(tipoRaw, "phishing"):
			threatType = "phishing"
			severity = "high"
		}

		// Obtener descripción si existe
		description := ""
		if len(parts) >= 4 {
			description = strings.TrimSpace(parts[3])
		}

		_, err := s.db.ExecContext(ctx, `
			INSERT INTO threat_phones (phone_national, country_code, threat_type, severity, confidence, source, description, first_seen, last_seen, flags)
			VALUES ($1, $2, $3::threat_type_enum, $4::severity_enum, 75, 'osint'::source_enum, $5, $6, $7, 1)
			ON CONFLICT (phone_national) DO UPDATE SET
				last_seen = EXCLUDED.last_seen,
				report_count = threat_phones.report_count + 1,
				confidence = GREATEST(threat_phones.confidence, EXCLUDED.confidence)
		`, phoneNational, countryCode, threatType, severity, description, now, now)

		if err != nil {
			errors++
			continue
		}
		records++

		if lineNum%1000 == 0 {
			s.updateSyncStatus(source, true, fmt.Sprintf("Imported %d phones...", records))
		}
	}

	// No hay sync_status específico para phones, pero podemos usar 'manual' o no actualizar
	fmt.Printf("[Phones] Import completed: %d records, %d errors\n", records, errors)
}

// parseCSVLine parsea una línea CSV con campos entre comillas
func parseCSVLine(line string) []string {
	var result []string
	var current strings.Builder
	inQuotes := false

	for i := 0; i < len(line); i++ {
		c := line[i]
		if c == '"' {
			inQuotes = !inQuotes
		} else if c == ',' && !inQuotes {
			result = append(result, current.String())
			current.Reset()
		} else {
			current.WriteByte(c)
		}
	}
	result = append(result, current.String())
	return result
}

// updateSyncStatus actualiza el estado de sincronización
func (s *Server) updateSyncStatus(source string, inProgress bool, message string) {
	s.syncMutex.Lock()
	defer s.syncMutex.Unlock()

	if status, ok := s.syncStatus[source]; ok {
		status.InProgress = inProgress
		status.Message = message
		if inProgress {
			status.StartedAt = time.Now()
		}
	}
}

// updateSyncStatusComplete marca una sincronización como completada
func (s *Server) updateSyncStatusComplete(source string, records, errors int64, message string) {
	s.syncMutex.Lock()
	defer s.syncMutex.Unlock()

	if status, ok := s.syncStatus[source]; ok {
		status.InProgress = false
		status.Records = records
		status.Errors = errors
		status.Message = message
	}
}

// handleListReports lista los reportes de usuarios
func (s *Server) handleListReports(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if s.db == nil {
		json.NewEncoder(w).Encode(map[string]string{"error": "Database not connected"})
		return
	}

	limit := getQueryInt(r, "limit", 50)
	offset := getQueryInt(r, "offset", 0)
	search := r.URL.Query().Get("search")
	status := r.URL.Query().Get("status")

	query := `
		SELECT ru.url, ru.domain, ru.primary_threat_type::text, ru.aggregated_score,
		       ru.total_reports, ru.unique_reporters, ru.status::text,
		       ru.first_reported_at, ru.last_reported_at, ru.promoted_to_threats
		FROM reported_urls ru
		WHERE (ru.flags & 1) = 1
	`
	args := []interface{}{}
	argCount := 0

	if search != "" {
		argCount++
		query += fmt.Sprintf(" AND (ru.url ILIKE $%d OR ru.domain ILIKE $%d)", argCount, argCount)
		args = append(args, "%"+search+"%")
	}
	if status != "" {
		argCount++
		query += fmt.Sprintf(" AND ru.status::text = $%d", argCount)
		args = append(args, status)
	}

	query += " ORDER BY ru.last_reported_at DESC"
	query += fmt.Sprintf(" LIMIT %d OFFSET %d", limit, offset)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}
	defer rows.Close()

	reports := []map[string]interface{}{}
	for rows.Next() {
		var urlStr, domain, status string
		var threatType sql.NullString
		var score, totalReports, uniqueReporters int
		var firstReported, lastReported time.Time
		var promoted bool

		if rows.Scan(&urlStr, &domain, &threatType, &score, &totalReports, &uniqueReporters,
			&status, &firstReported, &lastReported, &promoted) == nil {
			item := map[string]interface{}{
				"url":              urlStr,
				"domain":           domain,
				"aggregated_score": score,
				"total_reports":    totalReports,
				"unique_reporters": uniqueReporters,
				"status":           status,
				"first_reported":   firstReported.Format(time.RFC3339),
				"last_reported":    lastReported.Format(time.RFC3339),
				"promoted":         promoted,
			}
			if threatType.Valid {
				item["threat_type"] = threatType.String
			}
			reports = append(reports, item)
		}
	}

	// Get total count with same filters
	countQuery := `SELECT COUNT(*) FROM reported_urls WHERE (flags & 1) = 1`
	countArgs := []interface{}{}
	argCount = 0
	if search != "" {
		argCount++
		countQuery += fmt.Sprintf(" AND (url ILIKE $%d OR domain ILIKE $%d)", argCount, argCount)
		countArgs = append(countArgs, "%"+search+"%")
	}
	if status != "" {
		argCount++
		countQuery += fmt.Sprintf(" AND status::text = $%d", argCount)
		countArgs = append(countArgs, status)
	}

	var total int64
	s.db.QueryRow(countQuery, countArgs...).Scan(&total)

	json.NewEncoder(w).Encode(map[string]interface{}{
		"data":   reports,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// handleReportsStats devuelve estadísticas de los reportes
func (s *Server) handleReportsStats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if s.db == nil {
		json.NewEncoder(w).Encode(map[string]string{"error": "Database not connected"})
		return
	}

	stats := make(map[string]interface{})

	// Estadísticas generales de reported_urls
	var total, pending, confirmed, rejected, highScore, promoted int64
	s.db.QueryRow(`
		SELECT
			COUNT(*),
			COUNT(*) FILTER (WHERE status = 'pending'),
			COUNT(*) FILTER (WHERE status = 'confirmed'),
			COUNT(*) FILTER (WHERE status = 'rejected'),
			COUNT(*) FILTER (WHERE aggregated_score >= 70),
			COUNT(*) FILTER (WHERE promoted_to_threats = true)
		FROM reported_urls WHERE (flags & 1) = 1
	`).Scan(&total, &pending, &confirmed, &rejected, &highScore, &promoted)

	stats["total_reported_urls"] = total
	stats["pending"] = pending
	stats["confirmed"] = confirmed
	stats["rejected"] = rejected
	stats["high_score"] = highScore
	stats["promoted"] = promoted

	// Total de reportes individuales
	var totalIndividual int64
	s.db.QueryRow(`SELECT COUNT(*) FROM user_url_reports`).Scan(&totalIndividual)
	stats["total_individual_reports"] = totalIndividual

	// Reportes por tipo de amenaza
	byThreat := []map[string]interface{}{}
	rows, err := s.db.Query(`
		SELECT primary_threat_type::text, COUNT(*) as count
		FROM reported_urls
		WHERE (flags & 1) = 1 AND primary_threat_type IS NOT NULL
		GROUP BY primary_threat_type
		ORDER BY count DESC
	`)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var threatType string
			var count int64
			if rows.Scan(&threatType, &count) == nil {
				byThreat = append(byThreat, map[string]interface{}{
					"type":  threatType,
					"count": count,
				})
			}
		}
	}
	stats["by_threat_type"] = byThreat

	// Top reportadores (usuarios con más reportes)
	topReporters := []map[string]interface{}{}
	rows2, err := s.db.Query(`
		SELECT user_id, trust_score, total_reports, confirmed_reports
		FROM user_trust_scores
		WHERE total_reports > 0
		ORDER BY total_reports DESC
		LIMIT 10
	`)
	if err == nil {
		defer rows2.Close()
		for rows2.Next() {
			var userID string
			var trustScore, totalReports, confirmedReports int
			if rows2.Scan(&userID, &trustScore, &totalReports, &confirmedReports) == nil {
				// Anonimizar user_id parcialmente
				anonID := userID
				if len(userID) > 8 {
					anonID = userID[:4] + "..." + userID[len(userID)-4:]
				}
				topReporters = append(topReporters, map[string]interface{}{
					"user_id":           anonID,
					"trust_score":       trustScore,
					"total_reports":     totalReports,
					"confirmed_reports": confirmedReports,
				})
			}
		}
	}
	stats["top_reporters"] = topReporters

	json.NewEncoder(w).Encode(stats)
}

func extractTLD(domain string) string {
	parts := strings.Split(domain, ".")
	if len(parts) >= 2 {
		return parts[len(parts)-1]
	}
	return ""
}
