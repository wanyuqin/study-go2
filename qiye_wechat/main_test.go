package main

import (
	"fmt"
	"log"
	"testing"
)

func Test_getUserInfo(t *testing.T) {
	iamUsers, err := getUserInfo()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%#v", iamUsers)
}
