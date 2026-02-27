package dashboard

import (
	"context"
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"time"

	"github.com/baranovskis/go-ytdlp-bot/internal/config"
	"github.com/baranovskis/go-ytdlp-bot/internal/database"
	"github.com/baranovskis/go-ytdlp-bot/internal/logger"
	"github.com/rs/zerolog"
)

//go:embed templates/*.html
var templateFS embed.FS

//go:embed static/*
var staticFS embed.FS

var tmplMap map[string]*template.Template

type Server struct {
	Config    config.Dashboard
	DB        *database.DB
	Logger    zerolog.Logger
	LogWriter *logger.DBWriter
	srv       *http.Server
}

func NewServer(cfg config.Dashboard, db *database.DB, log zerolog.Logger, logWriter *logger.DBWriter) *Server {
	return &Server{
		Config:    cfg,
		DB:        db,
		Logger:    log,
		LogWriter: logWriter,
	}
}

func (s *Server) Run(ctx context.Context) {
	funcMap := template.FuncMap{
		"add":      func(a, b int) int { return a + b },
		"subtract": func(a, b int) int { return a - b },
	}

	pages := []string{"home.html", "downloads.html", "logs.html", "stats.html", "access.html", "filters.html", "login.html"}
	tmplMap = make(map[string]*template.Template, len(pages))
	for _, page := range pages {
		t, err := template.New("").Funcs(funcMap).ParseFS(templateFS, "templates/layout.html", "templates/"+page)
		if err != nil {
			s.Logger.Fatal().Str("page", page).Str("reason", err.Error()).Msg("failed parse dashboard template")
		}
		tmplMap[page] = t
	}

	mux := http.NewServeMux()

	staticContent, _ := fs.Sub(staticFS, "static")
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticContent))))

	mux.HandleFunc("GET /login", s.loginPage)
	mux.HandleFunc("POST /login", s.loginHandler)
	mux.HandleFunc("POST /logout", s.requireAuth(s.logoutHandler))

	mux.HandleFunc("GET /downloads", s.requireAuth(s.downloadsPage))
	mux.HandleFunc("GET /logs", s.requireAuth(s.logsPage))
	mux.HandleFunc("GET /api/logs/stream", s.requireAuth(s.logsStreamHandler))
	mux.HandleFunc("GET /stats", s.requireAuth(s.statsPage))
	mux.HandleFunc("GET /api/stats/stream", s.requireAuth(s.statsStreamHandler))
	mux.HandleFunc("GET /access", s.requireAuth(s.accessPage))
	mux.HandleFunc("POST /access/groups/approve", s.requireAuth(s.approveGroupHandler))
	mux.HandleFunc("POST /access/groups/reject", s.requireAuth(s.rejectGroupHandler))
	mux.HandleFunc("POST /access/groups/remove", s.requireAuth(s.removeGroupHandler))
	mux.HandleFunc("POST /access/users/approve", s.requireAuth(s.approveUserHandler))
	mux.HandleFunc("POST /access/users/reject", s.requireAuth(s.rejectUserHandler))
	mux.HandleFunc("POST /access/users/remove", s.requireAuth(s.removeUserHandler))
	mux.HandleFunc("GET /filters", s.requireAuth(s.filtersPage))
	mux.HandleFunc("POST /filters/add", s.requireAuth(s.addFilterHandler))
	mux.HandleFunc("POST /filters/update", s.requireAuth(s.updateFilterHandler))
	mux.HandleFunc("POST /filters/delete", s.requireAuth(s.deleteFilterHandler))

	mux.HandleFunc("GET /", s.requireAuth(s.homePage))

	port := s.Config.Port
	if port == 0 {
		port = 8080
	}

	s.srv = &http.Server{
		Addr:        fmt.Sprintf(":%d", port),
		Handler:     mux,
		ReadTimeout: 10 * time.Second,
		IdleTimeout: 60 * time.Second,
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		s.srv.Shutdown(shutdownCtx)
	}()

	s.Logger.Info().Int("port", port).Msg("dashboard server started")

	if err := s.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		s.Logger.Error().Str("reason", err.Error()).Msg("dashboard server error")
	}
}

func (s *Server) homePage(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	stats, _ := s.DB.GetStats()
	tmplMap["home.html"].ExecuteTemplate(w, "layout", map[string]any{
		"Stats": stats,
	})
}
