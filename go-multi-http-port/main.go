package main

import (
	"context"
	"fmt"
	"net/http"
)

func main() {

	ctx := context.Background()

	mux1 := http.ServeMux{}
	mux2 := http.ServeMux{}

	mux1.HandleFunc("/server1", func(writer http.ResponseWriter, request *http.Request) {
		writer.Write([]byte("Hello server 1"))

	})

	mux2.HandleFunc("/server2", func(writer http.ResponseWriter, request *http.Request) {
		writer.Write([]byte("Hello server 2"))

	})

	ser1 := http.Server{
		Addr: "0.0.0.0:8080",
		Handler: http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			mux1.ServeHTTP(writer, request)
		}),
	}

	ser2 := http.Server{
		Addr: "0.0.0.0:8081",
		Handler: http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			mux2.ServeHTTP(writer, request)
		}),
	}

	go func() {
		fmt.Println("start server 1")
		err := ser1.ListenAndServe()
		if err != nil {
			fmt.Printf("server 1 failed: %v\n", err)
			return
		}

	}()

	go func() {
		fmt.Println("start server 2")
		err := ser2.ListenAndServe()
		if err != nil {
			fmt.Printf("server 2 failed: %v\n", err)
			return
		}

	}()

	<-ctx.Done()
	fmt.Println("server shout down")
	ser1.Shutdown(ctx)
	ser2.Shutdown(ctx)
}
