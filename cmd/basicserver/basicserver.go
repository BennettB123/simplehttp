package main

import (
	"fmt"

	"github.com/BennettB123/simplehttp"
)

const PORT uint16 = 3030

func main() {
	server := simplehttp.NewServer(PORT)

	server.Get("/", func() {
		fmt.Println("we're in the GET / callback!")
	})

	server.Post("/", func() {
		fmt.Println("we're in the POST / callback!")
	})

	server.Put("/", func() {
		fmt.Println("we're in the PUT / callback!")
	})

	server.Delete("/", func() {
		fmt.Println("we're in the DELETE / callback!")
	})

	err := server.Start()
	if err != nil {
		fmt.Println("There was an error starting the server:", err)
	}
}
