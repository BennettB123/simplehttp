package simplehttp

type response struct {
	statusLine statusLine
	httpMessage
}

type statusLine struct {
	httpversion  string
	statusCode   uint
	reasonPhrase string
}
