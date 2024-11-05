package bot

import (
	"bufio"
	"context"
	"github.com/baranovskis/go-ytdlp-bot/internal/config"
	"github.com/baranovskis/go-ytdlp-bot/internal/ytdlp"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/rs/zerolog"
	"net/url"
	"os"
	"path"
	"regexp"
	"slices"
	"strings"
)

type Bot struct {
	API    *bot.Bot
	Config *config.Config
	Logger zerolog.Logger
}

func Init(config *config.Config, log zerolog.Logger) *Bot {
	return &Bot{
		Config: config,
		Logger: log,
	}
}

func (b *Bot) Run(ctx context.Context) {
	botAPI, err := b.NewBot()
	if err != nil {
		b.Logger.Fatal().
			Str("reason", err.Error()).
			Msg("failed create new bot api instance")
	}

	b.API = botAPI

	b.API.RegisterHandlerMatchFunc(b.matchVideoHostFunc, b.downloadVideoHandler)

	b.API.Start(ctx)
}

func (b *Bot) NewBot() (*bot.Bot, error) {
	opts := []bot.Option{
		bot.WithSkipGetMe(),
	}

	botAPI, err := bot.New(b.Config.Bot.Token, opts...)
	if err != nil {
		return nil, err
	}

	me, err := botAPI.GetMe(context.TODO())
	if err != nil {
		return nil, err
	}

	b.Logger.Info().
		Str("account", me.Username).
		Msg("authorized success, bot api instance created")
	return botAPI, nil
}

func (b *Bot) matchVideoHostFunc(update *models.Update) bool {
	if update.Message == nil {
		return false
	}

	u, err := url.Parse(update.Message.Text)
	if err != nil {
		return false
	}

	for _, filter := range b.Config.Bot.Filter {
		if slices.Contains(filter.Hosts, u.Host) {
			match := true

			if strings.TrimSpace(filter.PathRegEx) != "" {
				if match, err = regexp.MatchString(filter.PathRegEx, u.Path); err != nil {
					b.Logger.Error().
						Str("reason", err.Error()).
						Msg("failed regex match")
					return false
				}
			}

			return match
		}
	}

	return false
}

func (b *Bot) downloadVideoHandler(ctx context.Context, chat *bot.Bot, update *models.Update) {
	u, err := url.Parse(update.Message.Text)
	if err != nil {
		b.Logger.Error().
			Str("text", update.Message.Text).
			Str("reason", err.Error()).Msg("failed parse url")
		return
	}

	b.Logger.Info().
		Str("url", u.String()).
		Msg("triggered video download")

	command := ytdlp.Init(b.Config, b.Logger)

	// Remove query params if needed
	for _, filter := range b.Config.Bot.Filter {
		if slices.Contains(filter.Hosts, u.Host) {
			if filter.ExcludeQueryParams {
				u.RawQuery = ""
			}

			if len(filter.CookiesFile) > 0 {
				command.Cookies(filter.CookiesFile)
			}
		}
	}

	info, err := command.Run(context.TODO(), u.String())
	if err != nil {
		b.Logger.Error().
			Str("url", u.String()).
			Str("reason", err.Error()).
			Msg("failed video download")
		return
	}

	b.Logger.Info().
		Str("url", u.String()).
		Str("file", info.Filename).
		Msg("success video download")

	processedFile, err := os.Open(path.Join(b.Config.Storage.Path, info.Filename))
	if err != nil {
		b.Logger.Error().
			Str("path", path.Join(b.Config.Storage.Path, info.Filename)).
			Str("reason", err.Error()).
			Msg("failed video open")
		return
	}

	_, err = chat.SendMediaGroup(ctx, &bot.SendMediaGroupParams{
		ChatID: update.Message.Chat.ID,
		Media: []models.InputMedia{
			&models.InputMediaVideo{
				Media:           "attach://" + info.Filename,
				Caption:         info.Title,
				MediaAttachment: bufio.NewReader(processedFile),
			},
		},
		ReplyParameters: &models.ReplyParameters{
			MessageID: update.Message.ID,
			ChatID:    update.Message.Chat.ID,
		},
	})

	if err != nil {
		b.Logger.Error().
			Int64("chat_id", update.Message.Chat.ID).
			Str("path", processedFile.Name()).
			Str("error", err.Error()).
			Msg("failed video to chat upload")
		return
	}

	b.Logger.Info().
		Int("message_id", update.Message.ID).
		Str("file", info.Filename).
		Msg("success video upload")

	_ = processedFile.Close()

	if b.Config.Storage.RemoveAfterReply {
		if err = os.Remove(processedFile.Name()); err != nil {
			b.Logger.Error().
				Str("path", processedFile.Name()).
				Str("error", err.Error()).Msg("failed video remove")
		}
	}
}
