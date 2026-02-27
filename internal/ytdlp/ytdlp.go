package ytdlp

import (
	"context"
	"encoding/json"
	"github.com/baranovskis/go-ytdlp-bot/internal/config"
	"github.com/lrstanley/go-ytdlp"
	"github.com/rs/zerolog"
	"time"
)

type YtDlp struct {
	Command *ytdlp.Command
}

func Init(cfg *config.Config, log zerolog.Logger) *YtDlp {
	command := ytdlp.New().
		FormatSort("res,vcodec:h264").
		Format("bestvideo[vcodec^=avc1]+bestaudio[ext=m4a]/bestvideo[ext=mp4]+bestaudio[ext=m4a]/mp4").
		MergeOutputFormat("mp4").
		RecodeVideo("mp4").
		PostProcessorArgs("ffmpeg:-c:v libx264 -preset fast -crf 23 -pix_fmt yuv420p -c:a aac -movflags +faststart").
		NoOverwrites().
		NoPlaylist().
		PlaylistItems("1:1").
		ConcurrentFragments(5).
		Continue().
		NoProgress().
		ProgressFunc(100*time.Millisecond, func(prog ytdlp.ProgressUpdate) {
			log.Debug().
				Str("file", prog.Filename).
				Str("format", prog.Info.Format).
				Str("percent", prog.PercentString()).
				Dur("eta", prog.ETA()).
				Msgf("yt-dlp - %s", prog.Status)
		}).
		SetWorkDir(cfg.Storage.Path).
		Output("%(extractor)s_%(id)s.%(ext)s").
		PrintJSON()

	if cfg.Verbose {
		command.Verbose()
	}

	return &YtDlp{
		Command: command,
	}
}

func (b *YtDlp) Cookies(file string) {
	b.Command.Cookies(file)
}

func (b *YtDlp) Run(ctx context.Context, url ...string) (*Info, error) {
	r, err := b.Command.Run(ctx, url...)
	if err != nil {
		return nil, err
	}

	var info Info
	if err = json.Unmarshal([]byte(r.Stdout), &info); err != nil {
		return nil, err
	}

	return &info, nil
}
