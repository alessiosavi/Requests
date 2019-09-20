package requests

import (
	"testing"

	"github.com/alessiosavi/Requests/datastructure"
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

func TestSendRequest(t *testing.T) {
	var resp *datastructure.RequestResponse
	resp = makeBadRequestURL1()
	if resp == nil || resp.Error == nil {
		t.Fail()
	}
	t.Log("makeBadRequestURL1 Passed!")

	resp = makeBadRequestURL2()
	if resp == nil || resp.Error == nil {
		t.Fail()
	}
	t.Log("makeBadRequestURL2 Passed!")

	resp = makeOKRequestURL3()
	if resp == nil || resp.Error != nil || resp.StatusCode != 200 {
		t.Fail()
	}
	t.Log("makeOKRequestURL3 Passed!")

	t.Log("Tests Passed!")
}

func makeBadRequestURL1() *datastructure.RequestResponse {
	return SendRequest("tcp://google.it", "GET", nil, nil)
}
func makeBadRequestURL2() *datastructure.RequestResponse {
	return SendRequest("google.it", "GET", nil, nil)
}
func makeOKRequestURL3() *datastructure.RequestResponse {
	return SendRequest("https://google.it", "GET", nil, nil)
}

func dumpResponse(resp *datastructure.RequestResponse, t *testing.T) {
	t.Log(string(resp.Body))
	t.Log(resp.StatusCode)
	t.Log(resp.Headers)
	t.Log(resp.Error)
	t.Log(resp)
}
