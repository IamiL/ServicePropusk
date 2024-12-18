package handler_mux_v1

import (
	"encoding/json"
	"net/http"
	userService "rip/internal/service/user"
)

type Credentials struct {
	Password string `json:"password"`
	Login    string `json:"login"`
}

func SigninHandler(uService userService.UserService) func(
	http.ResponseWriter,
	*http.Request,
) {
	return func(w http.ResponseWriter, r *http.Request) {
		var creds Credentials
		// Get the JSON body and decode into credentials
		err := json.NewDecoder(r.Body).Decode(&creds)
		if err != nil {
			// If the structure of the body is wrong, return an HTTP error
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		sessionToken, expiresAt, err := uService.Auth(
			r.Context(),
			creds.Login,
			creds.Password,
		)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
		}

		http.SetCookie(
			w, &http.Cookie{
				Name:    "session_token",
				Value:   sessionToken,
				Expires: expiresAt,
			},
		)
		// Get the expected password from our in memory map
		//expectedPassword, ok := users[creds.Username]
		//
		//// If a password exists for the given user
		//// AND, if it is the same as the password we received, the we can move ahead
		//// if NOT, then we return an "Unauthorized" status
		//if !ok || expectedPassword != creds.Password {
		//	w.WriteHeader(http.StatusUnauthorized)
		//	return
		//}

		// Create a new random session token
		// we use the "github.com/google/uuid" library to generate UUIDs
		//sessionToken := uuid.NewString()
		//expiresAt := time.Now().Add(120 * time.Second)
		//
		//// Set the token in the session map, along with the session information
		//sessions[sessionToken] = session{
		//	username: creds.Username,
		//	expiry:   expiresAt,
		//}

		// Finally, we set the client cookie for "session_token" as the session token we just generated
		// we also set an expiry time of 120 seconds

	}
}
