package requests

import (
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/alessiosavi/Requests/datastructure"
)

// Remove comment for set the log at debug level
var req Request // = InitDebugRequest()

func TestCreateHeaderList(t *testing.T) {
	t.Parallel() // Create a simple headers
	headersKey := `Content-Type`
	headersValue := `application/json`

	err := req.CreateHeaderList(headersKey, headersValue)
	if err == nil {
		t.Error("Error, request is not initialized!")
	}

	request, err := InitRequest("http://", "POST", nil, false, false)
	if err != nil {
		t.Error("Error!: ", err)
	}

	err = request.CreateHeaderList(headersKey, headersValue)
	if err != nil {
		t.Error("Error!", err)
	}
	if strings.Compare(request.Req.Header.Get(headersKey), headersValue) != 0 {
		t.Error("Headers key mismatch!")
	}

}

func TestSendRequest(t *testing.T) {
	t.Parallel()
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
		r.SendRequest("http://127.0.0.1:9999", "GET", nil, []string{"Connection", "Close"}, false, 0)
	}
}

func BenchmarkRequestPOSTWithoutTLS(t *testing.B) {
	var r Request
	for i := 0; i < t.N; i++ {
		r.SendRequest("http://127.0.0.1:9999", "POST", []byte{}, []string{"Connection", "Close"}, false, 0)
	}
}

func BenchmarkParallelRequestGETWithoutTLS(t *testing.B) {
	var n = t.N
	var requests = make([]Request, n)
	for i := 0; i < n; i++ {
		req, err := InitRequest("http://127.0.0.1:9999", "GET", nil, true, false)
		if err == nil && req != nil {
			req.AddHeader("Connection", "Close")
			requests[i] = *req
		} else if err != nil {
			t.Error("error: ", err)
		}
	}
	for i := 0; i < t.N; i++ {
		ParallelRequest(requests, runtime.NumCPU())
	}
}

func BenchmarkParallelRequestPOSTWithoutTLS(t *testing.B) {
	var n = t.N
	var requests = make([]Request, n)
	for i := 0; i < n; i++ {
		req, err := InitRequest("http://127.0.0.1:9999", "POST", []byte{}, true, false)
		if err == nil && req != nil {
			req.AddHeader("Connection", "Close")
			requests[i] = *req
		} else if err != nil {
			t.Error("error: ", err)
		}
	}
	for i := 0; i < t.N; i++ {
		ParallelRequest(requests, runtime.NumCPU())
	}
}

func makeBadRequestURL1() *datastructure.Response {
	return req.SendRequest("tcp://google.it", "GET", nil, nil, true, 0)
}
func makeBadRequestURL2() *datastructure.Response {
	return req.SendRequest("google.it", "GET", nil, nil, true, 0)
}
func makeOKRequestURL3() *datastructure.Response {
	return req.SendRequest("https://google.it", "GET", nil, nil, true, 0)
}

type headerTestCase struct {
	input    []string
	expected bool
	number   int
}

func TestRequest_CreateHeaderList(t *testing.T) {
	t.Parallel()
	var request *Request
	request, err := InitRequest("http://", "POST", nil, false, false)
	if err != nil {
		t.Error("Error!", err)
	}
	cases := []headerTestCase{
		{input: []string{"Content-Type", "text/plain"}, expected: true, number: 1},
		{input: []string{"Content-Type"}, expected: false, number: 2},
		{input: []string{"Content-Type", "text/plain", "Error"}, expected: false, number: 3},
	}
	for _, c := range cases {
		err := request.CreateHeaderList(c.input...)
		if (c.expected && err != nil) || (!c.expected && err == nil) {
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
	t.Parallel()
	var request Request

	// create a listener with the desired port.
	l, err := net.Listen("tcp", "127.0.0.1:8082")
	if err != nil {
		t.Fatal(err)
	}

	ts := httptest.NewUnstartedServer(nil)
	// NewUnstartedServer creates a listener. Close that listener and replace
	// with the one we created.
	_ = ts.Listener.Close()
	ts.Listener = l

	// Start the server.
	ts.Start()

	cases := []requestTestCase{

		// GET
		{host: "http://localhost:8082/", method: "GET", body: nil, skipTLS: false, expected: nil, number: 1},
		{host: "http://localhost:8082/", method: "GET", body: nil, skipTLS: true, expected: nil, number: 2},
		{host: "localhost:8082/", method: "GET", body: nil, skipTLS: false, expected: errors.New("PREFIX_URL_NOT_VALID"), number: 3},
		// POST
		{host: "localhost:8082/", method: "POST", body: []byte{}, skipTLS: true, expected: errors.New("PREFIX_URL_NOT_VALID"), number: 4},
		{host: "localhost:8082/", method: "POST", body: nil, skipTLS: true, expected: errors.New("PREFIX_URL_NOT_VALID"), number: 5},
		{host: "http://localhost:8082/", method: "HEAD", body: nil, skipTLS: false, expected: errors.New("HTTP_METHOD_NOT_MANAGED"), number: 6},
		{host: "http://localhost:8080/", method: "GET", body: nil, skipTLS: false, expected: errors.New("ERROR_SENDING_REQUEST"), number: 7},
	}

	for _, c := range cases {
		resp := request.SendRequest(c.host, c.method, c.body, nil, c.skipTLS, 0)
		if c.expected != resp.Error {
			if c.expected != nil && resp.Error != nil {
				if !strings.Contains(resp.Error.Error(), c.expected.Error()) {
					t.Errorf("Expected %v, received %v [test n. %d]", c.expected, resp.Error, c.number)
				}
			} else {
				t.Error("Url not reachable! Spawn a simple server (python3 -m http.server 8081 || python -m SimpleHTTPServer 8081)")
			}
		}
	}

	// Cleanup.
	ts.Close()
}

func TestRequest_InitRequest(t *testing.T) {
	t.Parallel()
	cases := []requestTestCase{

		// GET
		{host: "http://localhost:8081/", method: "GET", body: nil, skipTLS: false, expected: nil, number: 1},
		{host: "http://localhost:8081/", method: "GET", body: nil, skipTLS: true, expected: nil, number: 2},
		{host: "localhost:8081/", method: "GET", body: nil, skipTLS: false, expected: errors.New("PREFIX_URL_NOT_VALID"), number: 3},
		// POST
		{host: "localhost:8081/", method: "POST", body: []byte{}, skipTLS: true, expected: errors.New("PREFIX_URL_NOT_VALID"), number: 4},
		{host: "localhost:8081/", method: "POST", body: nil, skipTLS: true, expected: errors.New("PREFIX_URL_NOT_VALID"), number: 5},
		{host: "http://localhost:8081/", method: "HEAD", body: nil, skipTLS: false, expected: errors.New("HTTP_METHOD_NOT_MANAGED"), number: 6},
	}

	for _, c := range cases {
		_, err := InitRequest(c.host, c.method, c.body, c.skipTLS, false)
		if c.expected != err {
			if c.expected.Error() != err.Error() {
				t.Errorf("Expected %v, received %v [test n. %d]", c.expected, err.Error(), c.number)
			}
		}
	}
}

func Test_Headers(t *testing.T) {
	t.Parallel()
	var req Request
	// create a listener with the desired port.
	l, err := net.Listen("tcp", "127.0.0.1:8083")
	if err != nil {
		log.Fatal(err)
	}

	f := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("1", "1")
		w.Header().Set("2", "2")
		w.Header().Set("3", "3")
		w.Header().Set("4", "4")
		w.Header().Set("5", "5")
		w.Header().Set("6", "6")
		_, _ = fmt.Fprintf(w, "Hello, %s", r.Proto)
	})
	ts := httptest.NewUnstartedServer(f)
	// NewUnstartedServer creates a listener. Close that listener and replace
	// with the one we created.
	_ = ts.Listener.Close()
	ts.Listener = l

	// Start the server.
	ts.Start()
	time.Sleep(1 * time.Millisecond)

	url := `http://127.0.0.1:8083`
	resp := req.SendRequest(url, "GET", nil, nil, true, 1*time.Second)

	if resp.Error != nil {
		t.Error("Request failed: ", resp.Error)
	}
	if len(resp.Headers) < 6 {
		t.Error("Not enough headers: ", len(resp.Headers))
		t.Error(resp.Headers)
	}
	ts.CloseClientConnections()
	ts.Close()

}
func TestRequest_ExecuteRequest(t *testing.T) {
	t.Parallel() // create a listener with the desired port.
	l, err := net.Listen("tcp", "127.0.0.1:8084")
	if err != nil {
		log.Fatal(err)
	}

	ts := httptest.NewUnstartedServer(nil)
	// NewUnstartedServer creates a listener. Close that listener and replace
	// with the one we created.
	_ = ts.Listener.Close()
	ts.Listener = l

	// Start the server.
	ts.Start()

	cases := []requestTestCase{
		// GET
		{host: "http://localhost:8084/", method: "GET", body: nil, skipTLS: false, expected: nil, number: 1},
		{host: "http://localhost:8084/", method: "GET", body: nil, skipTLS: true, expected: nil, number: 2},
		{host: "localhost:8084/", method: "GET", body: nil, skipTLS: false, expected: errors.New("PREFIX_URL_NOT_VALID"), number: 3},
		// POST
		{host: "localhost:8084/", method: "POST", body: []byte{}, skipTLS: true, expected: errors.New("PREFIX_URL_NOT_VALID"), number: 4},
		{host: "localhost:8084/", method: "POST", body: nil, skipTLS: true, expected: errors.New("PREFIX_URL_NOT_VALID"), number: 5},
		{host: "http://localhost:8084/", method: "HEAD", body: nil, skipTLS: false, expected: errors.New("HTTP_METHOD_NOT_MANAGED"), number: 6},
		{host: "http://localhost:8080/", method: "GET", body: nil, skipTLS: false, expected: errors.New("ERROR_SENDING_REQUEST"), number: 7},
	}

	client := &http.Client{}
	for _, c := range cases {
		req, err := InitRequest(c.host, c.method, c.body, c.skipTLS, false)
		if err == nil {
			resp := req.ExecuteRequest(client)

			if c.expected != nil && resp.Error != nil {
				if !strings.Contains(resp.Error.Error(), c.expected.Error()) {
					t.Errorf("Expected %v, received %v [test n. %d]", c.expected, resp.Error, c.number)
				}
			}
		}
	}
	// Cleanup.
	ts.Close()
}

type timeoutTestCase struct {
	host    string
	method  string
	body    []byte
	skipTLS bool
	time    int
	number  int
}

func TestRequest_Timeout(t *testing.T) {
	t.Parallel()
	// Need to run the server present in example/server_example.py
	cases := []timeoutTestCase{
		// GET
		{host: "https://localhost:5000/timeout", method: "GET", body: nil, skipTLS: true, time: 11, number: 1},
	}

	for _, c := range cases {
		var req Request // = InitDebugRequest()
		req.SetTimeout(time.Second * time.Duration(c.time))
		start := time.Now()
		resp := req.SendRequest(c.host, c.method, c.body, nil, c.skipTLS, 0)
		elapsed := time.Since(start)
		if resp.Error != nil {
			t.Errorf("Received an error -> %v [test n. %d].\n Be sure that the python server on ./example folder is up and running", resp.Error, c.number)
		}
		if time.Duration(c.time)*time.Second < elapsed {
			t.Error("Error timeout")
		}
	}
}

func TestParallelRequest(t *testing.T) {
	t.Parallel()
	start := time.Now()
	// This array will contains the list of request
	var reqs []Request
	// This array will contains the response from the given request
	var response []datastructure.Response

	// Set to run at max N request in parallel (use CPU count for best effort)
	var N = runtime.NumCPU()
	// Create the list of request
	for i := 0; i < 1000; i++ {
		// Run against the `server_example.py` present in this folder
		req, err := InitRequest("https://127.0.0.1:5000", "GET", nil, true, false) // Alternate cert validation
		if err != nil {
			t.Error("Error request [", i, "]. Error: ", err)
		} else {
			req.SetTimeout(10 * time.Second)
			reqs = append(reqs, *req)
		}
	}

	// Run the request in parallel
	response = ParallelRequest(reqs, N)

	elapsed := time.Since(start)

	for i := range response {
		if response[i].Error != nil {
			t.Error("Error request [", i, "]. Error: ", response[i].Error)
		}
	}
	t.Logf("Sending %d Requests took %s", len(reqs), elapsed)
}

func Test_escapeURL(t *testing.T) {
	type args struct {
		url string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO: Add test cases.
		{
			name: "testOK",
			args: args{"https://example.com/api/items?lang=en&search=escape this path"},
			want: "https://example.com/api/items?lang=en&search=escape%20this%20path",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := escapeURL(tt.args.url); got != tt.want {
				t.Errorf("escapeURL() = %v, want %v", got, tt.want)
			}
		})
	}
}
