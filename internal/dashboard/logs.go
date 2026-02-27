package dashboard

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strconv"

	"github.com/baranovskis/go-ytdlp-bot/internal/database"
)

const logsPerPage = 100

func (s *Server) logsPage(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	level := r.URL.Query().Get("level")
	search := r.URL.Query().Get("search")

	logs, total, err := s.DB.ListLogs(database.LogFilter{
		Level:  level,
		Search: search,
		Limit:  logsPerPage,
		Offset: (page - 1) * logsPerPage,
	})
	if err != nil {
		s.Logger.Error().Str("reason", err.Error()).Msg("failed list logs")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	totalPages := int(math.Ceil(float64(total) / float64(logsPerPage)))

	data := map[string]any{
		"Logs":       logs,
		"Total":      total,
		"Page":       page,
		"TotalPages": totalPages,
		"Level":      level,
		"Search":     search,
	}

	tmplMap["logs.html"].ExecuteTemplate(w, "layout", data)
}

func (s *Server) logsStreamHandler(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	ch := s.LogWriter.Subscribe()
	defer s.LogWriter.Unsubscribe(ch)

	for {
		select {
		case <-r.Context().Done():
			return
		case event, ok := <-ch:
			if !ok {
				return
			}
			data, _ := json.Marshal(event)
			fmt.Fprintf(w, "data: %s\n\n", data)
			flusher.Flush()
		}
	}
}
