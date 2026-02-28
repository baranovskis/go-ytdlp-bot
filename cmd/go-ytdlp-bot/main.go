package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/baranovskis/go-ytdlp-bot/internal/bot"
	"github.com/baranovskis/go-ytdlp-bot/internal/config"
	"github.com/baranovskis/go-ytdlp-bot/internal/dashboard"
	"github.com/baranovskis/go-ytdlp-bot/internal/database"
	"github.com/baranovskis/go-ytdlp-bot/internal/logger"
	"github.com/lrstanley/go-ytdlp"
)

func main() {
	ytdlp.MustInstall(context.TODO(), nil)

	configPath := flag.String("c", "./config.yaml", "path to go-ytdlp-bot config")
	flag.Parse()

	// Bootstrap logger (console only) for early startup
	log := logger.GetLogger()

	cfg, err := config.GetConfiguration(*configPath)
	if err != nil {
		log.Fatal().Str("reason", err.Error()).Msg("failed get configuration")
	}

	if err := os.MkdirAll(cfg.Storage.Path, os.ModePerm); err != nil {
		log.Fatal().
			Str("reason", err.Error()).
			Str("path", cfg.Storage.Path).
			Msg("failed create storage folder")
	}

	dbPath := cfg.Database.Path
	if dbPath == "" {
		dbPath = "data/bot.db"
	}

	db, err := database.Open(dbPath)
	if err != nil {
		log.Fatal().Str("reason", err.Error()).Msg("failed open database")
	}
	defer db.Close()

	if err := db.Migrate(); err != nil {
		log.Fatal().Str("reason", err.Error()).Msg("failed run database migrations")
	}

	// Seed config filters into DB (only if table is empty)
	var seedFilters []struct {
		Hosts              []string
		ExcludeQueryParams bool
		PathRegex          string
		CookiesFile        string
	}
	for _, f := range cfg.Bot.Filter {
		seedFilters = append(seedFilters, struct {
			Hosts              []string
			ExcludeQueryParams bool
			PathRegex          string
			CookiesFile        string
		}{
			Hosts:              f.Hosts,
			ExcludeQueryParams: f.ExcludeQueryParams,
			PathRegex:          f.PathRegEx,
			CookiesFile:        f.CookiesFile,
		})
	}
	if err := db.SeedFilters(seedFilters); err != nil {
		log.Fatal().Str("reason", err.Error()).Msg("failed seed filters")
	}

	// Re-create logger with DB writer for log capture
	dbWriter := logger.NewDBWriter(db)
	log = logger.GetLoggerWithDB(dbWriter)

	log.Info().Str("path", dbPath).Msg("database initialized")

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	dash := dashboard.NewServer(cfg.Dashboard, db, log, dbWriter)
	go dash.Run(ctx)

	botApi := bot.Init(cfg, log, db)
	botApi.Run(ctx)
}
