package main

import (
	"context"
	"flag"
	"github.com/baranovskis/go-ytdlp-bot/internal/bot"
	"github.com/baranovskis/go-ytdlp-bot/internal/config"
	"github.com/baranovskis/go-ytdlp-bot/internal/logger"
	"github.com/lrstanley/go-ytdlp"
	"os"
	"os/signal"
)

func main() {
	ytdlp.MustInstall(context.TODO(), nil)

	configPath := flag.String("c", "./config.yaml", "path to go-ytdlp-bot config")
	flag.Parse()

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

	botApi := bot.Init(cfg, log)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	botApi.Run(ctx)
}
