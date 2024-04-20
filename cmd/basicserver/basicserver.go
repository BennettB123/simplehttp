package main

import (
	"fmt"

	"github.com/BennettB123/simplehttp"
)

const PORT uint16 = 3030

func main() {
	server := simplehttp.NewServer(PORT)

	server.Get("/", func() error {
		fmt.Println("we're in the GET / callback!")
		return nil
	})

	server.Post("/", func() error {
		fmt.Println("we're in the POST / callback!")
		return nil
	})

	server.Put("/", func() error {
		fmt.Println("we're in the PUT / callback!")
		return nil
	})

	server.Delete("/", func() error {
		fmt.Println("we're in the DELETE / callback!")
		return nil
	})

	server.Get("/error", func() error {
		return fmt.Errorf("i am an error within a user callback")
	})

	err := server.Start()
	if err != nil {
		fmt.Println("There was an error starting the server:", err)
	}
}
