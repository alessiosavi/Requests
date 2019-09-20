package requests

import (
	"testing"
)

func TestCreateHeaderList(t *testing.T) {
	// Create a simple headers
	headersKey := `Content-Type`
	headersValue := `application/json`

	contentTypeHeaders := CreateHeaderList(headersKey, headersValue)

	t.Log("Headers => ", contentTypeHeaders)

	if len(contentTypeHeaders) != 1 {
		t.Error("size error!")
	}

	if len(contentTypeHeaders[0]) != 2 {
		t.Error("key value headers size mismatch")
	}

	headersKeyTest := contentTypeHeaders[0][0]
	headersValueTest := contentTypeHeaders[0][1]
	if headersKey != headersKeyTest {
		t.Error("Headers key mismatch!")
	}
	if headersValue != headersValueTest {
		t.Error("Headers value mismatch!")
	}
}
