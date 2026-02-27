package dashboard

import (
	"math"
	"net/http"
	"strconv"

	"github.com/baranovskis/go-ytdlp-bot/internal/database"
)

const downloadsPerPage = 50

func (s *Server) downloadsPage(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	status := r.URL.Query().Get("status")
	userID, _ := strconv.ParseInt(r.URL.Query().Get("user_id"), 10, 64)

	downloads, total, err := s.DB.ListDownloads(database.DownloadFilter{
		Status: status,
		UserID: userID,
		Limit:  downloadsPerPage,
		Offset: (page - 1) * downloadsPerPage,
	})
	if err != nil {
		s.Logger.Error().Str("reason", err.Error()).Msg("failed list downloads")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	totalPages := int(math.Ceil(float64(total) / float64(downloadsPerPage)))

	data := map[string]any{
		"Downloads":  downloads,
		"Total":      total,
		"Page":       page,
		"TotalPages": totalPages,
		"Status":     status,
		"UserID":     userID,
	}

	tmplMap["downloads.html"].ExecuteTemplate(w, "layout", data)
}
