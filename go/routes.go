package friezechat

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

type Route struct {
	Name            string
	Method          string
	Pattern         string
	HandlerFunc     http.HandlerFunc
	applyMiddleware bool
}

type Routes []Route

func NewRouter() *mux.Router {
	router := mux.NewRouter().StrictSlash(true)
	for _, route := range routes {
		var handler http.Handler
		handler = route.HandlerFunc
		handler = Logger(handler, route.Name)
		if route.applyMiddleware {
			handler = authMiddleware(handler)
		}
		router.
			Methods(route.Method).
			Path(route.Pattern).
			Name(route.Name).
			Handler(handler)
	}

	return router
}

func Index(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello World!")
}

var routes = Routes{

	Route{
		"OpenChat",
		strings.ToUpper("Post"),
		"/chat/open",
		OpenChat,
		false,
	},
	Route{
		"GetMessages",
		strings.ToUpper("get"),
		"/chat/messages",
		GetMessages,
		true,
	},
	Route{
		"SendMessage",
		strings.ToUpper("Post"),
		"/chat/sendmesg",
		SendMessage,
		true,
	},
	/* 	Route{
		"ReceiveNotification",
		strings.ToUpper("Post"),
		"/_matrix/push/v1/notify",
		ReceiveNotification,
		true,
	}, */
	/* 	Route{
		"UpgradeWS",
		strings.ToUpper("get"),
		"/ws",
		UpgradeWS,
		true,
	}, */
}

func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqToken := r.Header.Get("Authorization")
		splitToken := strings.Split(reqToken, "Bearer")
		reqToken = splitToken[1]
		token, err := VerifyToken(strings.TrimSpace(reqToken))
		fmt.Println(token["username"])
		if err != nil {
			fmt.Println(err)
			fmt.Println("Token is not valid:", token)
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("Unauthorized"))
		} else {
			next.ServeHTTP(w, r)
		}
	})
}
