package simplehttp

import (
	"testing"
	"time"
)

func TestFormatHttpDate(t *testing.T) {
	// example format:
	//   Sun, 06 Nov 1994 08:49:37 GMT
	timestamps := make(map[time.Time]string)
	parsed, _ := time.Parse(time.RFC3339, "2006-01-02T15:04:05Z")
	timestamps[parsed] = "Mon, 02 Jan 2006 15:04:05 GMT"

	parsed, _ = time.Parse(time.RFC3339, "1999-12-25T01:02:03Z")
	timestamps[parsed] = "Sat, 25 Dec 1999 01:02:03 GMT"

	parsed, _ = time.Parse(time.RFC3339, "2024-05-12T05:53:22Z")
	timestamps[parsed] = "Sun, 12 May 2024 05:53:22 GMT"

	for timestamp, expected := range timestamps {
		actual := formatHttpDate(timestamp)
		if actual != expected {
			t.Fatalf("expected '%s' | actual '%s'", expected, actual)
		}
	}
}

func TestParseHeadersError(t *testing.T) {
	input := "No-Colon-In-Header"
	_, err := parseHeaders(input)
	if err == nil {
		t.Fatalf("expected an error, but it was nil")
	}
}

func TestParseHeaders(t *testing.T) {
	input := "Header-1: value-1" + lineEnd +
		"Header-2: value:with:colons" + lineEnd +
		"Header-3: \"in a string\"" + lineEnd +
		"Header-4:     extra-whitespace      "
	headers, err := parseHeaders(input)

	if err != nil {
		t.Fatalf("did not expect an error, but received the following: %v", err)
	}

	if headers["Header-1"] != "value-1" {
		t.Fatalf("incorrect header value. Expected 'value-1' | Actual '%s'", headers["Header-1"])
	}
	if headers["Header-2"] != "value:with:colons" {
		t.Fatalf("incorrect header value. Expected 'value:with:colons' | Actual '%s'", headers["Header-2"])
	}
	if headers["Header-3"] != "\"in a string\"" {
		t.Fatalf("incorrect header value. Expected '\"in a string\"' | Actual '%s'", headers["Header-3"])
	}
	if headers["Header-4"] != "extra-whitespace" {
		t.Fatalf("incorrect header value. Expected 'extra-whitespace' | Actual '%s'", headers["Header-4"])
	}
}
