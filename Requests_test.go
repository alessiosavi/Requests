package requests

import (
	"errors"
	"log"
	"net"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/alessiosavi/Requests/datastructure"
)

var req Request // = InitDebugRequest()

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
	// t.Log(resp.Dump())

	resp = makeBadRequestURL2()
	if resp == nil || resp.Error == nil {
		t.Fail()
	} else {
		t.Log("makeBadRequestURL2 Passed!")
	}
	// t.Log(resp.Dump())

	resp = makeOKRequestURL3()
	if resp == nil || resp.Error != nil || resp.StatusCode != 200 {
		t.Fail()
	} else {
		t.Log("makeOKRequestURL3 Passed!")
	}
	// t.Log(resp.Dump())
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

	// create a listener with the desired port.
	l, err := net.Listen("tcp", "127.0.0.1:8081")
	if err != nil {
		log.Fatal(err)
	}

	ts := httptest.NewUnstartedServer(nil)
	// NewUnstartedServer creates a listener. Close that listener and replace
	// with the one we created.
	ts.Listener.Close()
	ts.Listener = l

	// Start the server.
	ts.Start()

	cases := []requestTestCase{

		// GET
		requestTestCase{host: "http://localhost:8081/", method: "GET", body: nil, skipTLS: false, expected: nil, number: 1},
		requestTestCase{host: "http://localhost:8081/", method: "GET", body: nil, skipTLS: true, expected: nil, number: 2},
		requestTestCase{host: "localhost:8081/", method: "GET", body: nil, skipTLS: false, expected: errors.New("PREFIX_URL_NOT_VALID"), number: 3},
		// POST
		requestTestCase{host: "localhost:8081/", method: "POST", body: []byte{}, skipTLS: true, expected: errors.New("PREFIX_URL_NOT_VALID"), number: 4},
		requestTestCase{host: "localhost:8081/", method: "POST", body: nil, skipTLS: true, expected: errors.New("PREFIX_URL_NOT_VALID"), number: 5},
		requestTestCase{host: "http://localhost:8081/", method: "HEAD", body: nil, skipTLS: false, expected: errors.New("HTTP_METHOD_NOT_MANAGED"), number: 6},
		requestTestCase{host: "http://localhost:8080/", method: "GET", body: nil, skipTLS: false, expected: errors.New("ERROR_SENDING_REQUEST"), number: 7},
	}

	for _, c := range cases {
		resp := request.SendRequest(c.host, c.method, c.body, c.skipTLS)
		if c.expected != resp.Error {
			if c.expected == nil && resp.Error != nil {
				t.Error("Url not reachable! Spawn a simple server (python3 -m http.server 8081 || python -m SimpleHTTPServer 8081)")
				continue
			}

			if !strings.Contains(resp.Error.Error(), c.expected.Error()) {
				t.Errorf("Expected %v, received %v [test n. %d]", c.expected, resp.Error, c.number)
			}
		}
	}

	// Cleanup.
	ts.Close()
}

func TestRequest_InitRequest(t *testing.T) {

	cases := []requestTestCase{

		// GET
		requestTestCase{host: "http://localhost:8081/", method: "GET", body: nil, skipTLS: false, expected: nil, number: 1},
		requestTestCase{host: "http://localhost:8081/", method: "GET", body: nil, skipTLS: true, expected: nil, number: 2},
		requestTestCase{host: "localhost:8081/", method: "GET", body: nil, skipTLS: false, expected: errors.New("PREFIX_URL_NOT_VALID"), number: 3},
		// POST
		requestTestCase{host: "localhost:8081/", method: "POST", body: []byte{}, skipTLS: true, expected: errors.New("PREFIX_URL_NOT_VALID"), number: 4},
		requestTestCase{host: "localhost:8081/", method: "POST", body: nil, skipTLS: true, expected: errors.New("PREFIX_URL_NOT_VALID"), number: 5},
		requestTestCase{host: "http://localhost:8081/", method: "HEAD", body: nil, skipTLS: false, expected: errors.New("HTTP_METHOD_NOT_MANAGED"), number: 6},
	}

	for _, c := range cases {
		_, err := InitRequest(c.host, c.method, c.body, nil, c.skipTLS)
		if c.expected != err {
			if c.expected.Error() != err.Error() {
				t.Errorf("Expected %v, received %v [test n. %d]", c.expected, err.Error(), c.number)
			}
		}
	}
}

func TestRequest_ExecuteRequest(t *testing.T) {

	// create a listener with the desired port.
	l, err := net.Listen("tcp", "127.0.0.1:8081")
	if err != nil {
		log.Fatal(err)
	}

	ts := httptest.NewUnstartedServer(nil)
	// NewUnstartedServer creates a listener. Close that listener and replace
	// with the one we created.
	ts.Listener.Close()
	ts.Listener = l

	// Start the server.
	ts.Start()

	cases := []requestTestCase{
		// GET
		requestTestCase{host: "http://localhost:8081/", method: "GET", body: nil, skipTLS: false, expected: nil, number: 1},
		requestTestCase{host: "http://localhost:8081/", method: "GET", body: nil, skipTLS: true, expected: nil, number: 2},
		requestTestCase{host: "localhost:8081/", method: "GET", body: nil, skipTLS: false, expected: errors.New("PREFIX_URL_NOT_VALID"), number: 3},
		// POST
		requestTestCase{host: "localhost:8081/", method: "POST", body: []byte{}, skipTLS: true, expected: errors.New("PREFIX_URL_NOT_VALID"), number: 4},
		requestTestCase{host: "localhost:8081/", method: "POST", body: nil, skipTLS: true, expected: errors.New("PREFIX_URL_NOT_VALID"), number: 5},
		requestTestCase{host: "http://localhost:8081/", method: "HEAD", body: nil, skipTLS: false, expected: errors.New("HTTP_METHOD_NOT_MANAGED"), number: 6},
		requestTestCase{host: "http://localhost:8080/", method: "GET", body: nil, skipTLS: false, expected: errors.New("ERROR_SENDING_REQUEST"), number: 7},
	}

	for _, c := range cases {
		req, err := InitRequest(c.host, c.method, c.body, nil, c.skipTLS)
		if err == nil {
			resp := req.ExecuteRequest()
			if c.expected != resp.Error {
				if c.expected == nil && resp.Error != nil {
					t.Error("Url not reachable! Spawn a simple server (python3 -m http.server 8081 || python -m SimpleHTTPServer 8081)")
					continue
				}

				if !strings.Contains(resp.Error.Error(), c.expected.Error()) {
					t.Errorf("Expected %v, received %v [test n. %d]", c.expected, resp.Error, c.number)
				}
			}

		}

	}

}
