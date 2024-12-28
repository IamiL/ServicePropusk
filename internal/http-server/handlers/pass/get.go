package passhandler

import (
	"encoding/json"
	"fmt"
	"net/http"
	passService "rip/internal/service/pass"
	"strconv"
	"time"
	"unicode/utf8"
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
}

func PassesHandler(pService *passService.PassService) func(
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

		var beginDateFilter *time.Time = nil

		if params.Get("begin_date") != "" {
			beginDateFilterStr := params.Get("begin_date")
			if utf8.RuneCountInString(beginDateFilterStr) < 10 {

			} else {
				day, err := strconv.Atoi(beginDateFilterStr[0:2])
				if err != nil {
					fmt.Println("day: ", beginDateFilterStr[0:2])
					fmt.Println(err)
				}

				month, err := strconv.Atoi(beginDateFilterStr[3:5])
				if err != nil {
					fmt.Println("month: ", beginDateFilterStr[3:5])
					fmt.Println(err)
				}

				year, err := strconv.Atoi(beginDateFilterStr[6:10])
				if err != nil {
					fmt.Println("year: ", beginDateFilterStr[6:10])
				}

				timeTemp := time.Date(
					year,
					time.Month(month),
					day,
					23,
					59,
					59,
					7,
					time.UTC,
				)

				beginDateFilter = &timeTemp
			}
		}

		var endDateFilter *time.Time = nil

		if params.Get("end_date") != "" {
			endDateFilterStr := params.Get("end_date")
			if utf8.RuneCountInString(endDateFilterStr) < 10 {

			} else {
				day, err := strconv.Atoi(endDateFilterStr[0:2])
				if err != nil {
					fmt.Println("day: ", endDateFilterStr[0:2])
					fmt.Println(err)
				}

				month, err := strconv.Atoi(endDateFilterStr[3:5])
				if err != nil {
					fmt.Println("month: ", endDateFilterStr[3:5])
					fmt.Println(err)
				}

				year, err := strconv.Atoi(endDateFilterStr[6:10])
				if err != nil {
					fmt.Println("year: ", endDateFilterStr[6:10])
				}

				timeTemp := time.Date(
					year,
					time.Month(month),
					day,
					23,
					59,
					59,
					7,
					time.UTC,
				)

				endDateFilter = &timeTemp
			}
		}

		passes, err := pService.Passes(
			r.Context(),
			statusFilter,
			beginDateFilter,
			endDateFilter,
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		resp := make([]PassResponse, 0, len(*passes))

		for _, p := range *passes {
			resp = append(
				resp,
				PassResponse{
					User{p.User.Login},
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

func PassHandler(pService *passService.PassService) func(
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
