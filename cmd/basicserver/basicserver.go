package main

import (
	"encoding/json"
	"fmt"

	"github.com/BennettB123/simplehttp"
)

const PORT uint16 = 3030

// implement basic logger that writes to the console
type ConsoleLogger struct{}

func (ConsoleLogger) LogMessage(message string) {
	fmt.Println(message)
}

func main() {
	server := simplehttp.NewServer(PORT)
	server.Logger = ConsoleLogger{}

	server.Get("/", func(req simplehttp.Request, res *simplehttp.Response) error {
		fmt.Println("we're in the GET / callback!")

		res.SetHtml("<h1>Hello, world!</h1>")
		res.SetHeader("Custom-Header", "custom-header-value")
		return nil
	})

	server.Get("/json-string", func(req simplehttp.Request, res *simplehttp.Response) error {
		fmt.Println("we're in the GET /json-string callback!")

		type User struct {
			Name  string
			Age   uint
			Email string
		}

		myUser := User{"foo", 42, "baz@email.com"}
		myJson, err := json.Marshal(myUser)
		if err != nil {
			return err
		}

		res.SetJson(string(myJson))
		return nil
	})

	server.Get("/json-any", func(req simplehttp.Request, res *simplehttp.Response) error {
		fmt.Println("we're in the GET /json-any callback!")

		type User struct {
			Name  string
			Age   uint
			Email string
		}

		myUser := User{"foo", 42, "baz@email.com"}
		res.SetJson(myUser)
		return nil
	})

	server.Get("/file", func(req simplehttp.Request, res *simplehttp.Response) error {
		fmt.Println("we're in the GET /htmlfile callback!")

		err := res.SetFile("./files/NON-EXISTENT-FILE")
		if err != nil {
			fmt.Println(err)
		}

		err = res.SetFile("./files/index.html")
		if err != nil {
			fmt.Println(err)
		}

		return nil
	})

	server.Post("/", func(req simplehttp.Request, res *simplehttp.Response) error {
		fmt.Println("we're in the POST / callback!")

		res.SetStatus(201)
		return nil
	})

	server.Get("/error", func(req simplehttp.Request, res *simplehttp.Response) error {
		return fmt.Errorf("i am an error within a user callback")
	})

	err := server.Start()
	if err != nil {
		fmt.Println("There was an error starting the server:", err)
	}
}
