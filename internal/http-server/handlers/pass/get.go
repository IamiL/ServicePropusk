package handler_mux_v1

import (
	"encoding/json"
	"fmt"
	"net/http"
	passService "rip/internal/service/pass"
	"strconv"
	"time"
)

type PassResponse struct {
	User        User      `json:"user"`
	ID          string    `json:"id"`
	VisitorName string    `json:"visitor_name"`
	DateVisit   time.Time `json:"date_visit"`
	Status      int       `json:"status"`
}

type User struct {
	Login string `json:"login"`
	ID    string `json:"id"`
}

func PassesHandler(pService passService.PassService) func(
	http.ResponseWriter,
	*http.Request,
) {
	return func(w http.ResponseWriter, r *http.Request) {
		params := r.URL.Query()

		var statusFilter *int = nil

		if params.Get("status") != "" {
			status, err := strconv.Atoi(params.Get("status"))
			if err != nil {

			} else {
				statusFilter = &status
			}
		}

		passes, err := pService.Passes(r.Context(), statusFilter)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		resp := make([]PassResponse, 0, len(*passes))

		for _, p := range *passes {
			resp = append(
				resp,
				PassResponse{
					User{p.User.Login, p.User.Id},
					p.ID,
					p.VisitorName,
					p.DateVisit,
					p.Status,
				},
			)
		}

		w.Header().Set("Content-Type", "application/json")

		if err := json.NewEncoder(w).Encode(
			resp,
		); err != nil {
			http.Error(
				w,
				fmt.Sprintf("error building the response, %v", err),
				http.StatusInternalServerError,
			)
			return
		}

	}
}

func PassHandler(pService passService.PassService) func(
	http.ResponseWriter,
	*http.Request,
) {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		pass, err := pService.Pass(r.Context(), id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")

		if err := json.NewEncoder(w).Encode(
			pass,
		); err != nil {
			http.Error(
				w,
				fmt.Sprintf("error building the response, %v", err),
				http.StatusInternalServerError,
			)
			return
		}

	}
}
