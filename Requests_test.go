package requests

import (
	"testing"

	"github.com/alessiosavi/Requests/datastructure"
)

var req Request

func TestCreateHeaderList(t *testing.T) {
	// Create a simple headers
	headersKey := `Content-Type`
	headersValue := `application/json`

	if !req.CreateHeaderList(headersKey, headersValue) {
		t.Error("Unable to create headers list")
	}

	if len(req.Headers) != 1 {
		t.Error("size error!")
	}

	if len(req.Headers[0]) != 2 {
		t.Error("key value headers size mismatch")
	}

	headersKeyTest := req.Headers[0][0]
	headersValueTest := req.Headers[0][1]
	if headersKey != headersKeyTest {
		t.Error("Headers key mismatch!")
	}
	if headersValue != headersValueTest {
		t.Error("Headers value mismatch!")
	}
}

func TestSendRequest(t *testing.T) {

	var resp *datastructure.Response
	resp = makeBadRequestURL1()
	if resp == nil || resp.Error == nil {
		t.Fail()
	} else {
		t.Log("makeBadRequestURL1 Passed!")
	}
	resp = makeBadRequestURL2()
	if resp == nil || resp.Error == nil {
		t.Fail()
	} else {
		t.Log("makeBadRequestURL2 Passed!")
	}
	resp = makeOKRequestURL3()
	if resp == nil || resp.Error != nil || resp.StatusCode != 200 {
		t.Fail()
	} else {
		t.Log("makeOKRequestURL3 Passed!")
	}
}

func makeBadRequestURL1() *datastructure.Response {
	return req.SendRequest("tcp://google.it", "GET", nil, true)
}
func makeBadRequestURL2() *datastructure.Response {
	return req.SendRequest("google.it", "GET", nil, true)
}
func makeOKRequestURL3() *datastructure.Response {
	return req.SendRequest("https://google.it", "GET", nil, true)
}

func dumpResponse(resp *datastructure.Response, t *testing.T) {
	t.Log(string(resp.Body))
	t.Log(resp.StatusCode)
	t.Log(resp.Headers)
	t.Log(resp.Error)
	t.Log(resp)
}
