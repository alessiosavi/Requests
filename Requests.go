package requests

import (
	"bytes"
	"crypto/tls"
	"errors"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/alessiosavi/Requests/datastructure"
	"github.com/onrik/logrus/filename"
	log "github.com/sirupsen/logrus"
)

// AllowedMethod rappresent the HTTP method allowed in the request
var allowedMethod []string = []string{"GET", "POST", "HEAD", "PUT", "DELETE", "OPTIONS"}

// Request will contains all the data related to the current HTTP request and response.
type Request struct {
	Req     *http.Request          // Request
	Tr      *http.Transport        // Transport layer, used for enable/disable TLS verification
	Method  string                 // HTTP method of the request
	URL     string                 // URL where send the request
	Data    []byte                 // BODY in case of POST, ARGS in case of GET
	Headers [][]string             // List of headers to send in the request
	Resp    datastructure.Response // Struct for save the response
	Timeout time.Duration          // Timeout of the request
}

// InitDebugRequest is delegated to set the log level in order to debug the flow
func InitDebugRequest() Request {
	Formatter := new(log.TextFormatter)
	Formatter.TimestampFormat = "Jan _2 15:04:05.000000000"
	Formatter.FullTimestamp = true
	Formatter.ForceColors = true
	log.AddHook(filename.NewHook()) // Print filename + line at every log
	log.SetFormatter(Formatter)
	log.SetLevel(log.DebugLevel)
	return Request{}
}

// methodIsAllowed is delegated to verify if the given HTTP Method is compliant
func (req *Request) methodIsAllowed(method string) bool {
	for i := range allowedMethod {
		if method == allowedMethod[i] {
			req.Method = method
			return true
		}
	}
	return false
}

// SetTimeoutString is delegated to set a timeout different from the default one.
// It take the unit of the time, and an integer related to the time to wait
// Ex: MS, 100 (100 milliseconds)
// NOTE: The default timeout (0) means no timeout waiting -> infinite timeout
func (req *Request) SetTimeoutString(unit string, value int) {
	var t time.Duration
	if value == 0 {
		log.Warning("WARNING! Setting a timeout of 0 means inifite timeout!!")
		req.Timeout = 0
		return
	} else if value < 0 {
		// Don't allow value < 0
		value = -value
	}

	mul := time.Duration(value)

	switch strings.ToUpper(unit) {
	case "NS":
		// Nanoseconds
		t = time.Nanosecond * mul
	case "MS":
		// Milliseconds
		t = time.Millisecond * mul
	case "S":
		// Seconds
		t = time.Second * mul
	case "M":
		// Minutes
		t = time.Minute * mul
	default:
		log.Warning("WARNING! [" + unit + "] is not a valid unit! Falling back to milliseconds!")
		unit = "MS"
		t = time.Millisecond * mul
	}

	log.Info("Setting a timeout of [", value, "] "+unit)
	req.Timeout = t
}

// SetTimeout is delegated to validate the given timeout and set to the request
func (req *Request) SetTimeout(t time.Duration) {

	value := t.Milliseconds()
	if value == 0 {
		log.Warning("WARNING! Setting a timeout of 0 means inifite timeout!!")
	} else if value < 0 {
		value = -value
		log.Warning("WARNING! Get a negative timeout, using absolute value")
	}
	req.Timeout = time.Duration(value)
}

// CreateHeaderList is delegated to initialize a list of headers.
// Every row of the matrix contains [key,value]
func (req *Request) CreateHeaderList(headers ...string) bool {
	if headers == nil {
		req.Headers = [][]string{}
		return false
	}

	length := len(headers)

	if len(headers)%2 != 0 {
		log.Debug(`Headers have to be a "key:value" list. Got instead a odd number of elements`)
		return false
	}

	req.Headers = make([][]string, length/2)
	counter := 0

	for i := 0; i < length; i += 2 {
		tmp := make([]string, 2)
		key := headers[i]
		value := headers[i+1]
		tmp[0] = key
		tmp[1] = value
		log.Debug("createHeaderList | ", counter, ") Key: ", key, " Value: ", value)
		req.Headers[counter] = tmp
		counter++
	}
	log.Debug("createHeaderList | LIST: ", req.Headers)
	return true
}

func (req *Request) initPostRequest() {
	if strings.ToUpper(req.Method) == "POST" {
		if req.Data == nil {
			req.Data = []byte("")
		}
	}
}

func (req *Request) initGetRequest() {
	if strings.ToUpper(req.Method) == "GET" && req.Data != nil {
		args := string(req.Data)
		// Arguments are not in the URL, concatenate the args in the URL
		if !strings.Contains(req.URL, "?") {
			// Overwrite the "/" with the provided params
			if strings.HasSuffix(req.URL, "/") {
				index := strings.LastIndex(req.URL, "/")
				req.URL = req.URL[:index]
			}
			req.URL += "?" + args
		} else {
			// adding additional parameter to the one provided in the URL
			req.URL += "&" + args
		}
	}
}

// GetUlimitValue return the current and max value for ulimit
func GetUlimitValue() (uint64, uint64) {
	var rLimit syscall.Rlimit
	err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit)
	if err != nil {
		log.Error("Error Getting Rlimit: ", err)
	}
	log.Debug("Current Ulimit: ", rLimit.Cur)
	return rLimit.Cur, rLimit.Max
}

// ParallelRequest is delegated to run the given list of request in parallel, sending N request at each time
func ParallelRequest(reqs []Request, N int) []datastructure.Response {
	var wg sync.WaitGroup
	var results []datastructure.Response = make([]datastructure.Response, len(reqs))

	ulimitCurr, _ := GetUlimitValue()
	if uint64(N) > ulimitCurr {
		N = int(ulimitCurr)
		log.Warning("Provided a thread factor greater than current ulimit size, setting at MAX [", N, "] requests")
	}

	semaphore := make(chan struct{}, N)
	wg.Add(len(reqs))
	for i := 0; i < len(reqs); i++ {
		go func(i int) {
			semaphore <- struct{}{}
			defer func() { <-semaphore }()
			defer wg.Done()
			results[i] = reqs[i].ExecuteRequest()
		}(i)
	}
	wg.Wait()
	return results
}

// SetTLS is delegated to enable/disable TLS certificate validation
func (req *Request) SetTLS(skipTLS bool) {
	if skipTLS {
		// Accept not trusted SSL Certificates
		req.Tr = &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
		log.Debug("TLS certificate validation disabled")
	} else {
		req.Tr = &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: false}}
		log.Debug("TLS certificate validation enabled")
	}
}

// InitRequest is delegated to initialize a new request with the given parameter.
// NOTE: it will use the default timeout -> NO TIMEOUT. In order to specify a different timeout you can use the delegated method
func InitRequest(url, method string, bodyData []byte, headers []string, skipTLS bool) (*Request, error) {
	var err error
	var req Request

	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		err = errors.New("PREFIX_URL_NOT_VALID")
		log.Debug("sendRequest | Error! ", err, " URL: ", url)
		return nil, err
	}

	method = strings.ToUpper(method)

	// Validate HTTP method
	if !req.methodIsAllowed(method) {
		log.Warning("sendRequest | Method [" + method + "] is not allowed!")
		err = errors.New("METHOD_NOT_ALLOWED")
		return nil, err
	}

	// Manage TLS configuration
	req.SetTLS(skipTLS)
	// Set infinite timeout as default http/net
	req.SetTimeoutString("--", 0)
	// Create headers list
	req.CreateHeaderList(headers...)

	req.URL = url
	req.Data = bodyData

	switch req.Method {
	case "GET":
		req.initGetRequest()
		req.Req, err = http.NewRequest(req.Method, req.URL, nil)
	case "POST":
		req.initPostRequest()
		req.Req, err = http.NewRequest(req.Method, req.URL, bytes.NewReader(req.Data))
	case "PUT":
		req.Req, err = http.NewRequest(req.Method, req.URL, nil)
	case "DELETE":
		req.Req, err = http.NewRequest(req.Method, req.URL, nil)
	default:
		log.Debug("sendRequest | Unknown method -> " + method)
		err = errors.New("HTTP_METHOD_NOT_MANAGED")
	}

	if err != nil {
		log.Debug("sendRequest | Error while initializing a new request -> ", err)
		return nil, err
	}

	contentlengthPresent := false
	for i := range req.Headers {
		// log.Debug("sendRequest | Adding header: ", headers[i], " Len: ", len(headers[i]))
		key := req.Headers[i][0]
		value := req.Headers[i][1]
		if strings.EqualFold(`Authorization`, key) {
			req.Req.Header.Add(key, value)
		} else {
			req.Req.Header.Set(key, value)
		}
		if strings.EqualFold("Content-Length", key) {
			contentlengthPresent = true
		}
		//log.Debug("sendRequest | Adding header: {", key, "|", value, "}")
	}

	// If content lenght was not specified (only for POST)
	if req.Method == "POST" && !contentlengthPresent {
		contentlength := strconv.FormatInt(req.Req.ContentLength, 10)
		log.Debug("sendRequest | Content-length not provided, setting new one -> ", contentlength)
		req.Req.Header.Add("Content-Length", contentlength)
	}

	return &req, err
}

// ExecuteRequest is delegated to run a previously allocated request.
func (req *Request) ExecuteRequest() datastructure.Response {
	var response datastructure.Response
	var start time.Time = time.Now()
	var err error

	log.Debug("ExecuteRequest | Executing request ...")
	client := &http.Client{Transport: req.Tr, Timeout: req.Timeout}
	log.Debugf("Request: %+v\n", req.Req)
	resp, err := client.Do(req.Req)

	if err != nil {
		log.Error("Error executing request | ERR:", err)
		err = errors.New("ERROR_SENDING_REQUEST -> " + err.Error())
		response.Error = err
		return response
	}
	defer resp.Body.Close()
	//log.Debug("sendRequest | Request executed, reading response ...")
	bodyResp, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error("Unable to read response! | Err: ", err)
		err = errors.New("ERROR_READING_RESPONSE -> " + err.Error())
		response.Error = err
		return response
	}
	var headersResp []string
	for k, v := range resp.Header {
		headersResp = append(headersResp, join(k, `:`, strings.Join(v, `,`)))
	}

	response.Body = bodyResp
	response.StatusCode = resp.StatusCode
	response.Headers = headersResp
	response.Error = nil
	elapsed := time.Since(start)
	response.Time = elapsed
	log.Debug("ExecuteRequest | Elapsed -> ", elapsed, " | STOP!")
	return response
}

// SendRequest is delegated to initialize a new HTTP request.
func (req *Request) SendRequest(url, method string, bodyData []byte, skipTLS bool) *datastructure.Response {

	// Create a custom request
	var (
		err      error
		response datastructure.Response
		start    time.Time
	)

	start = time.Now()

	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		_error := errors.New("PREFIX_URL_NOT_VALID")
		log.Debug("sendRequest | Error! ", _error, " URL: ", url)
		response.Error = _error
		return &response
	}

	method = strings.ToUpper(method)

	// Validate method
	if !req.methodIsAllowed(method) {
		log.Debug("sendRequest | Method [" + method + "] is not allowed!")
		_error := errors.New("METHOD_NOT_ALLOWED")
		response.Error = _error
		return &response
	}

	// Manage TLS configuration
	req.SetTLS(skipTLS)
	// Set infinite timeout as default http/net
	req.SetTimeoutString("--", 0)

	req.URL = url
	req.Data = bodyData

	switch req.Method {
	case "GET":
		req.initGetRequest()
		req.Req, err = http.NewRequest(req.Method, req.URL, nil)
	case "POST":
		req.initPostRequest()
		req.Req, err = http.NewRequest(req.Method, req.URL, bytes.NewReader(req.Data))
	case "PUT":
		req.Req, err = http.NewRequest(req.Method, req.URL, nil)
	case "DELETE":
		req.Req, err = http.NewRequest(req.Method, req.URL, nil)
	default:
		log.Debug("sendRequest | Unknown method -> " + method)
		err = errors.New("HTTP_METHOD_NOT_MANAGED")
	}

	if err != nil {
		log.Debug("sendRequest | Error while initializing a new request -> ", err)
		response.Error = err
		return &response
	}

	contentlengthPresent := false
	for i := range req.Headers {
		log.Debug("sendRequest | Adding header: ", req.Headers[i], " Len: ", len(req.Headers[i]))
		key := req.Headers[i][0]
		value := req.Headers[i][1]
		if strings.EqualFold(`Authorization`, key) {
			req.Req.Header.Add(key, value)
		} else {
			req.Req.Header.Set(key, value)
		}
		if strings.EqualFold("Content-Length", key) {
			contentlengthPresent = true
		}
		log.Debug("sendRequest | Adding header: {", key, "|", value, "}")
	}

	if req.Method == "POST" && !contentlengthPresent {
		contentlength := strconv.FormatInt(req.Req.ContentLength, 10)
		log.Debug("sendRequest | Content-length not provided, setting new one -> ", contentlength)
		req.Req.Header.Add("Content-Length", contentlength)
	}

	log.Debug("sendRequest | Executing request ...")
	client := &http.Client{Transport: req.Tr, Timeout: req.Timeout}
	log.Debugf("sendRequest | Request: %+v\n", req.Req)

	resp, err := client.Do(req.Req)

	if err != nil {
		log.Debug("Error executing request | ERR:", err)
		response.Error = errors.New("ERROR_SENDING_REQUEST -> " + err.Error())
		return &response
	}
	defer resp.Body.Close()
	log.Debug("sendRequest | Request executed, reading response ...")
	bodyResp, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Debug("sendRequest | Unable to read response! | Err: ", err)
		response.Error = errors.New("ERROR_READING_RESPONSE -> " + err.Error())
		return &response
	}
	var headersResp []string
	for k, v := range resp.Header {
		headersResp = append(headersResp, join(k, `:`, strings.Join(v, `,`)))
	}

	response.Body = bodyResp
	response.StatusCode = resp.StatusCode
	response.Headers = headersResp
	response.Error = nil
	elapsed := time.Since(start)
	response.Time = elapsed
	log.Debug("sendRequest | Elapsed -> ", elapsed, " | STOP!")
	return &response
}

// Join is a quite efficient string concatenator
func join(strs ...string) string {
	var sb strings.Builder
	for _, str := range strs {
		sb.WriteString(str)
	}
	return sb.String()
}
