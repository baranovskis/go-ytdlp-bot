package bot

import (
	"bufio"
	"context"
	"net/url"
	"os"
	"path"
	"regexp"
	"slices"
	"strings"

	"github.com/baranovskis/go-ytdlp-bot/internal/cache"
	"github.com/baranovskis/go-ytdlp-bot/internal/config"
	"github.com/baranovskis/go-ytdlp-bot/internal/database"
	"github.com/baranovskis/go-ytdlp-bot/internal/ytdlp"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/rs/zerolog"
)

type Bot struct {
	API    *bot.Bot
	Config *config.Config
	Logger zerolog.Logger
	Cache  *cache.Cache
	DB     *database.DB
}

func Init(config *config.Config, log zerolog.Logger, db *database.DB) *Bot {
	return &Bot{
		Config: config,
		Logger: log,
		Cache:  cache.New(config.Cache.GetTTL(), config.Storage.RemoveAfterReply, log),
		DB:     db,
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
	b.API.RegisterHandlerMatchFunc(b.matchMyChatMember, b.myChatMemberHandler)

	b.API.Start(ctx)
}

func (b *Bot) NewBot() (*bot.Bot, error) {
	opts := []bot.Option{
		bot.WithSkipGetMe(),
		bot.WithDebugHandler(func(format string, args ...any) {}),
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

	filters, err := b.DB.ListFilters()
	if err != nil {
		b.Logger.Error().Str("reason", err.Error()).Msg("failed load filters from db")
		return false
	}

	for _, filter := range filters {
		if slices.Contains(filter.Hosts, u.Host) {
			match := true

			if strings.TrimSpace(filter.PathRegex) != "" {
				if match, err = regexp.MatchString(filter.PathRegex, u.Path); err != nil {
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
	if update.Message.From == nil {
		return
	}

	uname := update.Message.From.Username
	if uname == "" {
		uname = update.Message.From.FirstName
	}

	chatID := update.Message.Chat.ID
	userID := update.Message.From.ID
	isGroup := update.Message.Chat.Type == "group" || update.Message.Chat.Type == "supergroup"

	if isGroup {
		groupAllowed, _ := b.DB.IsGroupAllowed(chatID)
		if !groupAllowed {
			title := update.Message.Chat.Title
			if err := b.DB.AddPendingGroup(chatID, title); err != nil {
				b.Logger.Error().
					Int64("chat_id", chatID).
					Str("reason", err.Error()).
					Msg("failed add pending group")
			}
			b.Logger.Warn().
				Int64("chat_id", chatID).
				Str("title", title).
				Msg("access denied: group not allowed, registered as pending")
			return
		}
	} else {
		b.DB.RegisterUser(userID, uname)

		userAllowed, _ := b.DB.IsUserAllowed(userID)
		if !userAllowed {
			b.Logger.Warn().
				Int64("user_id", userID).
				Msg("access denied: user not allowed")
			return
		}
	}

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

	var cookiesFile string
	dbFilters, _ := b.DB.ListFilters()
	for _, filter := range dbFilters {
		if slices.Contains(filter.Hosts, u.Host) {
			if filter.ExcludeQueryParams {
				u.RawQuery = ""
			}

			if len(filter.CookiesFile) > 0 {
				cookiesFile = filter.CookiesFile
			}
		}
	}

	cleanURL := u.String()

	downloadID, _ := b.DB.InsertDownload(cleanURL, update.Message.From.ID, uname, update.Message.Chat.ID, "pending", "", "")

	result, err := b.Cache.GetOrDownload(ctx, cleanURL, func(_ context.Context) (*cache.Result, error) {
		command := ytdlp.Init(b.Config, b.Logger)

		if cookiesFile != "" {
			command.Cookies(cookiesFile)
		}

		info, err := command.Run(context.TODO(), cleanURL)
		if err != nil {
			return nil, err
		}

		return &cache.Result{
			FilePath: path.Join(b.Config.Storage.Path, info.Filename),
			Filename: info.Filename,
			Title:    info.Title,
		}, nil
	})

	if err != nil {
		b.Logger.Error().
			Str("url", cleanURL).
			Str("reason", err.Error()).
			Msg("failed video download")
		if downloadID > 0 {
			b.DB.UpdateDownloadStatus(downloadID, "failed", "", err.Error())
		}
		chat.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Failed to download video.",
			ReplyParameters: &models.ReplyParameters{
				MessageID: update.Message.ID,
				ChatID:    update.Message.Chat.ID,
			},
		})
		return
	}

	if downloadID > 0 {
		b.DB.UpdateDownloadStatus(downloadID, "success", result.Filename, "")
	}

	b.Logger.Info().
		Str("url", cleanURL).
		Str("file", result.Filename).
		Msg("success video download")

	processedFile, err := os.Open(result.FilePath)
	if err != nil {
		b.Logger.Error().
			Str("path", result.FilePath).
			Str("reason", err.Error()).
			Msg("failed video open")
		chat.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Failed to process downloaded video.",
			ReplyParameters: &models.ReplyParameters{
				MessageID: update.Message.ID,
				ChatID:    update.Message.Chat.ID,
			},
		})
		return
	}
	defer processedFile.Close()

	_, err = chat.SendMediaGroup(ctx, &bot.SendMediaGroupParams{
		ChatID: update.Message.Chat.ID,
		Media: []models.InputMedia{
			&models.InputMediaVideo{
				Media:           "attach://" + result.Filename,
				Caption:         result.Title,
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
		chat.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Failed to upload video. File may be too large.",
			ReplyParameters: &models.ReplyParameters{
				MessageID: update.Message.ID,
				ChatID:    update.Message.Chat.ID,
			},
		})
		return
	}

	b.Logger.Info().
		Int("message_id", update.Message.ID).
		Str("file", result.Filename).
		Msg("success video upload")
}

func (b *Bot) matchMyChatMember(update *models.Update) bool {
	return update.MyChatMember != nil
}

func (b *Bot) myChatMemberHandler(ctx context.Context, chat *bot.Bot, update *models.Update) {
	member := update.MyChatMember
	if member == nil {
		return
	}

	newType := member.NewChatMember.Type
	if newType != models.ChatMemberTypeMember && newType != models.ChatMemberTypeAdministrator {
		return
	}

	chatID := member.Chat.ID
	title := member.Chat.Title

	b.Logger.Info().
		Int64("chat_id", chatID).
		Str("title", title).
		Str("type", string(newType)).
		Msg("bot added to group")

	if err := b.DB.AddPendingGroup(chatID, title); err != nil {
		b.Logger.Error().
			Int64("chat_id", chatID).
			Str("reason", err.Error()).
			Msg("failed add pending group")
	}
}
