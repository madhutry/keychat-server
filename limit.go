package main

import (
    "net/http"
	"fmt"
    "golang.org/x/time/rate"	
)

var limiter = rate.NewLimiter(2, 5)

func limit(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if limiter.Allow() == false {
			fmt.Println("Failure..")
            http.Error(w, http.StatusText(429), http.StatusTooManyRequests)
            return
        }

        next.ServeHTTP(w, r)
    })
}