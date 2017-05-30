package handlers_test

import (
	"net/http"
	. "spacecafe/handlers"
	"testing"
)

func TestAssure(t *testing.T) {
	req, err := http.NewRequest("GET", "http://example.com/foo?body=1234&num=string&empty=", nil)
	if err != nil {
		t.Error("Failed to create request")
	}

	err = Assure(req, "doesnotexist")
	if err == nil {
		t.Error("Failed to assure existance")
	}

	err = Assure(req, "empty")
	if err == nil {
		t.Error("Shouldn't be empty")
	}

	err = Assure(req, "body:max=3")
	if err == nil {
		t.Error("Failed to assure length")
	}

	err = Assure(req, "num:numeric")
	if err == nil {
		t.Error("Failed to assure numeric string")
	}
}
