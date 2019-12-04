package requests

import (
	"errors"
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
	t.Log(resp.Dump())

	resp = makeBadRequestURL2()
	if resp == nil || resp.Error == nil {
		t.Fail()
	} else {
		t.Log("makeBadRequestURL2 Passed!")
	}
	t.Log(resp.Dump())

	resp = makeOKRequestURL3()
	if resp == nil || resp.Error != nil || resp.StatusCode != 200 {
		t.Fail()
	} else {
		t.Log("makeOKRequestURL3 Passed!")
	}
	t.Log(resp.Dump())
}

func BenchmarkRequestGETWithoutTLS(t *testing.B) {
	var r Request
	for i := 0; i < t.N; i++ {
		r.SendRequest("http://127.0.0.1:8080", "GET", nil, false)
	}
}

func BenchmarkRequestGETWithTLS(t *testing.B) {
	var r Request
	for i := 0; i < t.N; i++ {
		r.SendRequest("http://127.0.0.1:8080", "GET", nil, true)
	}
}

func BenchmarkRequestPOSTWithoutTLS(t *testing.B) {
	var r Request
	for i := 0; i < t.N; i++ {
		r.SendRequest("http://127.0.0.1:8080", "POST", []byte{}, false)
	}
}

func BenchmarkRequestPOSTWithTLS(t *testing.B) {
	var r Request
	for i := 0; i < t.N; i++ {
		r.SendRequest("http://127.0.0.1:8080", "POST", []byte{}, true)
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

type headerTestCase struct {
	input    []string
	expected bool
	number   int
}

func TestRequest_CreateHeaderList(t *testing.T) {
	var request Request
	cases := []headerTestCase{
		headerTestCase{input: []string{"Content-Type", "text/plain"}, expected: true, number: 1},
		headerTestCase{input: []string{"Content-Type"}, expected: false, number: 2},
		headerTestCase{input: []string{"Content-Type", "text/plain", "Error"}, expected: false, number: 3},
	}
	for _, c := range cases {
		if c.expected != request.CreateHeaderList(c.input...) {
			t.Errorf("Expected %v for input %v [test n. %d]", c.expected, c.input, c.number)
		}
	}
}

type requestTestCase struct {
	host     string
	method   string
	body     []byte
	skipTLS  bool
	expected error
	number   int
}

func TestRequest_SendRequest(t *testing.T) {
	var request Request

	cases := []requestTestCase{

		// GET
		requestTestCase{host: "http://localhost:8081/", method: "GET", body: nil, skipTLS: false, expected: nil, number: 1},
		requestTestCase{host: "http://localhost:8081/", method: "GET", body: nil, skipTLS: true, expected: nil, number: 2},
		requestTestCase{host: "localhost:8081/", method: "GET", body: nil, skipTLS: false, expected: errors.New("PREFIX_URL_NOT_VALID"), number: 3},
		// POST
		requestTestCase{host: "http://localhost:8081/", method: "POST", body: nil, skipTLS: false, expected: errors.New("BODY_NULL"), number: 4},
		requestTestCase{host: "http://localhost:8081/", method: "POST", body: nil, skipTLS: true, expected: errors.New("BODY_NULL"), number: 5},
		requestTestCase{host: "localhost:8081/", method: "POST", body: []byte{}, skipTLS: true, expected: errors.New("PREFIX_URL_NOT_VALID"), number: 6},
		requestTestCase{host: "http://localhost:8081/", method: "HEAD", body: nil, skipTLS: false, expected: errors.New("HTTP_METHOD_NOT_MANAGED"), number: 7},
		requestTestCase{host: "http://localhost:8080/", method: "GET", body: nil, skipTLS: false, expected: errors.New("ERROR_SENDING_REQUEST"), number: 8},
	}

	for _, c := range cases {
		resp := request.SendRequest(c.host, c.method, c.body, c.skipTLS)
		if c.expected != resp.Error {
			if c.expected == nil && resp.Error != nil {
				t.Error("Url not reachable! Spawn a simple server (python3 -m http.server 8081 || python -m SimpleHTTPServer 8081)")
				continue
			}
			if c.expected.Error() != resp.Error.Error() {
				t.Errorf("Expected %v, recived %v [test n. %d]", c.expected, resp.Error, c.number)
			}
		}
	}

}
