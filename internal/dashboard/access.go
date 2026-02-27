package dashboard

import (
	"net/http"
	"strconv"
)

func (s *Server) accessPage(w http.ResponseWriter, r *http.Request) {
	pendingGroups, err := s.DB.ListPendingGroups()
	if err != nil {
		s.Logger.Error().Str("reason", err.Error()).Msg("failed list pending groups")
	}

	allowedGroups, err := s.DB.ListAllowedGroups()
	if err != nil {
		s.Logger.Error().Str("reason", err.Error()).Msg("failed list allowed groups")
	}

	pendingUsers, err := s.DB.ListPendingUsers()
	if err != nil {
		s.Logger.Error().Str("reason", err.Error()).Msg("failed list pending users")
	}

	allowedUsers, err := s.DB.ListAllowedUsers()
	if err != nil {
		s.Logger.Error().Str("reason", err.Error()).Msg("failed list allowed users")
	}

	data := map[string]any{
		"PendingGroups": pendingGroups,
		"AllowedGroups": allowedGroups,
		"PendingUsers":  pendingUsers,
		"AllowedUsers":  allowedUsers,
	}

	tmplMap["access.html"].ExecuteTemplate(w, "layout", data)
}

func (s *Server) approveGroupHandler(w http.ResponseWriter, r *http.Request) {
	chatID, _ := strconv.ParseInt(r.FormValue("chat_id"), 10, 64)
	if chatID == 0 {
		http.Error(w, "Invalid chat ID", http.StatusBadRequest)
		return
	}

	if err := s.DB.ApprovePendingGroup(chatID); err != nil {
		s.Logger.Error().Str("reason", err.Error()).Msg("failed approve group")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/access", http.StatusSeeOther)
}

func (s *Server) rejectGroupHandler(w http.ResponseWriter, r *http.Request) {
	chatID, _ := strconv.ParseInt(r.FormValue("chat_id"), 10, 64)
	if chatID == 0 {
		http.Error(w, "Invalid chat ID", http.StatusBadRequest)
		return
	}

	if err := s.DB.RejectPendingGroup(chatID); err != nil {
		s.Logger.Error().Str("reason", err.Error()).Msg("failed reject group")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/access", http.StatusSeeOther)
}

func (s *Server) removeGroupHandler(w http.ResponseWriter, r *http.Request) {
	chatID, _ := strconv.ParseInt(r.FormValue("chat_id"), 10, 64)
	if chatID == 0 {
		http.Error(w, "Invalid chat ID", http.StatusBadRequest)
		return
	}

	if err := s.DB.RemoveAllowedGroup(chatID); err != nil {
		s.Logger.Error().Str("reason", err.Error()).Msg("failed remove group")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/access", http.StatusSeeOther)
}

func (s *Server) approveUserHandler(w http.ResponseWriter, r *http.Request) {
	userID, _ := strconv.ParseInt(r.FormValue("user_id"), 10, 64)
	if userID == 0 {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	if err := s.DB.ApprovePendingUser(userID); err != nil {
		s.Logger.Error().Str("reason", err.Error()).Msg("failed approve user")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/access", http.StatusSeeOther)
}

func (s *Server) rejectUserHandler(w http.ResponseWriter, r *http.Request) {
	userID, _ := strconv.ParseInt(r.FormValue("user_id"), 10, 64)
	if userID == 0 {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	if err := s.DB.RejectPendingUser(userID); err != nil {
		s.Logger.Error().Str("reason", err.Error()).Msg("failed reject user")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/access", http.StatusSeeOther)
}

func (s *Server) removeUserHandler(w http.ResponseWriter, r *http.Request) {
	userID, _ := strconv.ParseInt(r.FormValue("user_id"), 10, 64)
	if userID == 0 {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	if err := s.DB.RemoveAllowedUser(userID); err != nil {
		s.Logger.Error().Str("reason", err.Error()).Msg("failed remove user")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/access", http.StatusSeeOther)
}
