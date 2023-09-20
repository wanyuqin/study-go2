package main

import (
	"fmt"
	"log"
	"testing"
)

func TestNewSockListRequest(t *testing.T) {
	u, err := NewSockListRequest(150, 1, "b:BK1009", []string{"f12", "f14", "f2"})
	if err != nil {
		log.Fatal(err)
		return
	}

	fmt.Println(u.Path)

	response, err := u.DoRequest()
	if err != nil {
		log.Fatal(err)
		return
	}

	sockList := response.SockList()

	for _, v := range sockList {
		fmt.Printf("code %s name %s \n", v.Code, v.Name)
	}
}
