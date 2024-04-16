package main

import (
	"fmt"

	"github.com/BennettB123/simplehttp"
)

const PORT uint16 = 3030

func main() {
	server := simplehttp.NewServer(PORT)

	err := server.Start()
	if err != nil {
		fmt.Println("There was an error starting the server:", err)
	}
}
