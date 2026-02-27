package dashboard

import (
	"net/http"
	"strconv"
	"strings"
)

func (s *Server) filtersPage(w http.ResponseWriter, r *http.Request) {
	filters, err := s.DB.ListFilters()
	if err != nil {
		s.Logger.Error().Str("reason", err.Error()).Msg("failed list filters")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	tmplMap["filters.html"].ExecuteTemplate(w, "layout", map[string]any{
		"Filters": filters,
	})
}

func (s *Server) addFilterHandler(w http.ResponseWriter, r *http.Request) {
	hostsRaw := r.FormValue("hosts")
	hosts := parseHostsInput(hostsRaw)
	if len(hosts) == 0 {
		http.Error(w, "At least one host is required", http.StatusBadRequest)
		return
	}

	excludeQP := r.FormValue("exclude_query_params") == "on"
	pathRegex := strings.TrimSpace(r.FormValue("path_regex"))
	cookiesFile := strings.TrimSpace(r.FormValue("cookies_file"))

	if _, err := s.DB.InsertFilter(hosts, excludeQP, pathRegex, cookiesFile); err != nil {
		s.Logger.Error().Str("reason", err.Error()).Msg("failed add filter")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/filters", http.StatusSeeOther)
}

func (s *Server) updateFilterHandler(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(r.FormValue("id"), 10, 64)
	if id == 0 {
		http.Error(w, "Invalid filter ID", http.StatusBadRequest)
		return
	}

	hostsRaw := r.FormValue("hosts")
	hosts := parseHostsInput(hostsRaw)
	if len(hosts) == 0 {
		http.Error(w, "At least one host is required", http.StatusBadRequest)
		return
	}

	excludeQP := r.FormValue("exclude_query_params") == "on"
	pathRegex := strings.TrimSpace(r.FormValue("path_regex"))
	cookiesFile := strings.TrimSpace(r.FormValue("cookies_file"))

	if err := s.DB.UpdateFilter(id, hosts, excludeQP, pathRegex, cookiesFile); err != nil {
		s.Logger.Error().Str("reason", err.Error()).Msg("failed update filter")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/filters", http.StatusSeeOther)
}

func (s *Server) deleteFilterHandler(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(r.FormValue("id"), 10, 64)
	if id == 0 {
		http.Error(w, "Invalid filter ID", http.StatusBadRequest)
		return
	}

	if err := s.DB.DeleteFilter(id); err != nil {
		s.Logger.Error().Str("reason", err.Error()).Msg("failed delete filter")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/filters", http.StatusSeeOther)
}

func parseHostsInput(raw string) []string {
	var hosts []string
	for _, line := range strings.Split(raw, "\n") {
		for _, h := range strings.Split(line, ",") {
			h = strings.TrimSpace(h)
			if h != "" {
				hosts = append(hosts, h)
			}
		}
	}
	return hosts
}
