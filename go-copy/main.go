package main

import (
	"fmt"

	"github.com/jinzhu/copier"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"study-go/go-copy/entity"
	"study-go/go-copy/po"
)

func main() {
	po := []po.Po{{
		Common: po.Common{
			ID: primitive.NewObjectID(),
		},
	}}

	e := make([]entity.Entity, 0)

	fmt.Println(copier.Copy(&e, po))
	fmt.Println(e)
}
