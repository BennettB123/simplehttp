package simplehttp

type response struct {
	statusLine statusLine
	httpMessage
}

type statusLine struct {
	httpVersion  string
	statusCode   uint
	reasonPhrase string
}
