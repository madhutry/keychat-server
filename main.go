/*
 * eList
 *
 * This api document for eList.
 *
 * API version: 1.0.0
 * Contact: ***REMOVED***.tryambake@gmail.io
 * Generated by: Swagger Codegen (https://github.com/swagger-api/swagger-codegen.git)
 */

package main

import (
	"log"
	"net/http"

	sw "./go"
	_ "github.com/lib/pq"
)

func main() {
	log.Printf("Server started")

	router := sw.NewRouter()
	sw.InitSqllite()
	log.Fatal(http.ListenAndServe(":6060", router))
}
