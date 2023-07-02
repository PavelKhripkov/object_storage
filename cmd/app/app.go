package main

import (
	"github.com/PavelKhripkov/object_storage/internal/handler/api/http/v1"
	"github.com/julienschmidt/httprouter"
	"log"
	"net/http"
)

func main() {
	log.Println("Starting app")
	router := httprouter.New()
	itemHandler := v1.NewItemHandler(nil)

	log.Println("Registering handlers")
	itemHandler.Register(router)

	log.Println("Listening on port 11111")
	if err := http.ListenAndServe(":11111", router); err != nil {
		log.Fatal(err)
	}
}
