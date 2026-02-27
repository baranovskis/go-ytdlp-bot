package dashboard

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

func (s *Server) statsPage(w http.ResponseWriter, r *http.Request) {
	stats, err := s.DB.GetStats()
	if err != nil {
		s.Logger.Error().Str("reason", err.Error()).Msg("failed get stats")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	tmplMap["stats.html"].ExecuteTemplate(w, "layout", stats)
}

func (s *Server) statsStreamHandler(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	// Send initial stats
	s.sendStats(w, flusher)

	for {
		select {
		case <-r.Context().Done():
			return
		case <-ticker.C:
			s.sendStats(w, flusher)
		}
	}
}

func (s *Server) sendStats(w http.ResponseWriter, flusher http.Flusher) {
	stats, err := s.DB.GetStats()
	if err != nil {
		return
	}
	data, _ := json.Marshal(stats)
	fmt.Fprintf(w, "data: %s\n\n", data)
	flusher.Flush()
}
