package main

import (
	"log"
	"net/http"
	_ "net/http/pprof"
)

func main() {
	println("hello world")

	log.Println(http.ListenAndServe("localhost:8080", nil))
}
