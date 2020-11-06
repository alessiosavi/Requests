package requests

import (
	"bytes"
	"crypto/tls"
	"errors"
	"io/ioutil"
	"net/http"
	netURL "net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/alessiosavi/Requests/datastructure"
	"github.com/onrik/logrus/filename"
	log "github.com/sirupsen/logrus"
)

// AllowedMethod represent the HTTP method allowed in the request
var allowedMethod = []string{"GET", "POST", "HEAD", "PUT", "DELETE", "OPTIONS"}

// Request will contains all the data related to the current HTTP request and response.
type Request struct {
	Req     *http.Request          // Request
	Tr      *http.Transport        // Transport layer, used for enable/disable TLS verification
	Method  string                 // HTTP method of the request
	URL     string                 // URL where send the request
	Data    []byte                 // BODY in case of POST, ARGS in case of GET
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

// SetTimeout is delegated to validate the given timeout and set to the request
func (req *Request) SetTimeout(t time.Duration) {
	value := t.Milliseconds()
	if value == 0 {
		log.Debug("WARNING! Setting a timeout of 0 means infinite timeout!!")
	} else if value < 0 {
		value = -value
		log.Warning("WARNING! Get a negative timeout, using absolute value")
	}
	req.Timeout = time.Duration(value) * time.Millisecond
}

// AddCookie is delegated to add the given list of cookie to the request
func (req *Request) AddCookie(c ...*http.Cookie) error {
	if req.Req == nil {
		return errors.New("request not initialized")
	}
	for i := range c {
		req.Req.AddCookie(c[i])
	}
	return nil
}

// CreateHeaderList is delegated to initialize a list of headers.
// Every row of the matrix contains [key,value]
func (req *Request) CreateHeaderList(headers ...string) error {
	if headers == nil {
		return nil
	}
	if req.Req == nil {
		return errors.New("request is not initialized, please use the `InitRequest` method before apply the headers")
	}

	length := len(headers)

	if len(headers)%2 != 0 {
		err := errors.New(`headers have to be a "key:value" list, got instead a odd number of elements`)
		log.Debug(err)
		return err
	}

	counter := 0

	for i := 0; i < length; i += 2 {
		key := headers[i]
		value := headers[i+1]
		log.Debug("createHeaderList | ", counter, ") Key: ", key, " Value: ", value)
		counter++
		if strings.EqualFold(`Authorization`, key) {
			req.Req.Header.Set(key, value)
		} else {
			req.Req.Header.Add(key, value)
		}
		//log.Debug("sendRequest | Adding header: {", key, "|", value, "}")
	}
	log.Debug("createHeaderList | LIST: ", req.Req.Header)
	return nil
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

// ParallelRequest is delegated to run the given list of request in parallel, sending N request at each time
func ParallelRequest(reqs []Request, N int) []datastructure.Response {
	var wg sync.WaitGroup
	var results = make([]datastructure.Response, len(reqs))
	// Need to fix the hardcoded limit
	ulimitCurr := 512
	if N >= ulimitCurr {
		N = int(float64(ulimitCurr) * 0.7)
		log.Warning("Provided a thread factor greater than current ulimit size, setting at MAX [", N, "] requests")
	}
	for i := range reqs {
		reqs[i].Tr.MaxIdleConns = N
		reqs[i].Tr.MaxIdleConnsPerHost = N
		reqs[i].Tr.MaxConnsPerHost = N
	}

	semaphore := make(chan struct{}, N)
	wg.Add(len(reqs))
	client := &http.Client{}
	for i := 0; i < len(reqs); i++ {
		go func(i int) {
			semaphore <- struct{}{}
			results[i] = reqs[i].ExecuteRequest(client)
			wg.Done()
			func() { <-semaphore }()
		}(i)
	}
	wg.Wait()
	return results
}

// SetTLS is delegated to enable/disable TLS certificate validation
func (req *Request) SetTLS(skipTLS bool) {
	var transport *http.Transport = &http.Transport{DisableKeepAlives: false}
	if skipTLS {
		// Accept not trusted SSL Certificates
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		req.Tr = transport
		log.Debug("SetTLS | TLS certificate validation disabled")
	} else {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: false}
		req.Tr = transport
		log.Debug("SetTLS | TLS certificate validation enabled")
	}
}

func (req *Request) SetTransportLayer(tl *http.Transport) {
	req.Tr = tl
	log.Debugf("SetTransportLayer | The following transport layer have been set: [%+v]\n", tl)
}

// InitRequest is delegated to initialize a new request with the given parameter.
// NOTE: it will use the default timeout -> NO TIMEOUT. In order to specify a different timeout you can use the delegated method
// NOTE: headers have to be set with the delegated method
func InitRequest(url, method string, bodyData []byte, skipTLS, debug bool) (*Request, error) {
	var err error
	var req Request

	if debug {
		req = InitDebugRequest()
	}

	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		err = errors.New("PREFIX_URL_NOT_VALID")
		log.Debug("InitRequest | Error! ", err, " URL: ", url)
		return nil, err
	}

	method = strings.ToUpper(method)

	// Validate HTTP method
	if !req.methodIsAllowed(method) {
		log.Warning("InitRequest | Method [" + method + "] is not allowed!")
		err = errors.New("METHOD_NOT_ALLOWED")
		return nil, err
	}

	url = escapeURL(url)

	// Manage TLS configuration
	req.SetTLS(skipTLS)
	// Set infinite timeout as default http/net
	req.SetTimeout(time.Duration(0))

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
		log.Debug("InitRequest | Unknown method -> " + method)
		err = errors.New("HTTP_METHOD_NOT_MANAGED")
	}

	if err != nil {
		log.Debug("InitRequest | Error while initializing a new request -> ", err)
		return nil, err
	}

	return &req, err
}

func escapeURL(url string) string {
	// Escape GET parameters after first slash `/` and then concatenate it
	if firstSlash := strings.LastIndex(url, "?"); firstSlash > 0 {
		firstSlash++
		urlRune := []rune(url)
		urlFront := string(urlRune[:firstSlash])
		urlBack := string(urlRune[firstSlash:])

		urlBack = netURL.PathEscape(urlBack)

		// Concate front and back
		url = urlFront + urlBack
	}
	return url
}

// ExecuteRequest is delegated to run a previously allocated request.
func (req *Request) ExecuteRequest(client *http.Client) datastructure.Response {
	var response datastructure.Response
	var start = time.Now()
	var err error

	if client == nil {
		client = http.DefaultClient
	}

	log.Debug("ExecuteRequest | Executing request ...")
	//client := &http.Client{Transport: req.Tr, Timeout: req.Timeout}
	req.Tr.DisableKeepAlives = false
	client.Transport = req.Tr
	client.Timeout = req.Timeout
	log.Debugf("Request: %+v\n", req.Req)
	log.Debugf("Client: %+v\n", client)

	// If content length was not specified (only for POST) add an headers with the lenght of the request
	if req.Method == "POST" && req.Req.Header.Get("Content-Length") == "" {
		contentLength := strconv.FormatInt(req.Req.ContentLength, 10)
		req.Req.Header.Set("Content-Length", contentLength)
		log.Debug("ExecuteRequest | Setting Content-Length -> ", contentLength)

	}
	resp, err := client.Do(req.Req)

	if err != nil {
		log.Error("Error executing request | ERR:", err)
		err = errors.New("ERROR_SENDING_REQUEST -> " + err.Error())
		response.Error = err
		return response
	}

	defer resp.Body.Close()
	response.Headers = make(map[string]string, len(resp.Header))
	for k, v := range resp.Header {
		response.Headers[k] = strings.Join(v, `,`)
	}
	response.Cookie = resp.Cookies()

	//log.Debug("ExecuteRequest | Request executed, reading response ...")
	bodyResp, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error("Unable to read response! | Err: ", err)
		err = errors.New("ERROR_READING_RESPONSE -> " + err.Error())
		response.Error = err
		return response
	}

	response.Body = bodyResp
	response.StatusCode = resp.StatusCode
	response.Error = nil
	elapsed := time.Since(start)
	response.Time = elapsed
	response.Respnse = resp
	log.Debug("ExecuteRequest | Elapsed -> ", elapsed, " | STOP!")
	return response
}

// SendRequest is delegated to initialize a new HTTP request.
func (req *Request) SendRequest(url, method string, bodyData []byte, headers []string, skipTLS bool, timeout time.Duration) *datastructure.Response {

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
	req.SetTimeout(timeout)

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
	err = req.CreateHeaderList(headers...)
	if err != nil {
		log.Debug("sendRequest | Error while initializing the headers -> ", err)
		response.Error = err
		return &response
	}

	contentlengthPresent := false
	if strings.Compare(req.Req.Header.Get("Content-Length"), "") == 0 {
		contentlengthPresent = true
	}

	if req.Method == "POST" && !contentlengthPresent {
		contentLength := strconv.FormatInt(req.Req.ContentLength, 10)
		log.Debug("sendRequest | Content-length not provided, setting new one -> ", contentLength)
		req.Req.Header.Add("Content-Length", contentLength)
	}

	log.Debugf("sendRequest | Executing request .. %+v\n", req.Req)
	client := &http.Client{Transport: req.Tr, Timeout: req.Timeout}

	resp, err := client.Do(req.Req)

	if err != nil {
		log.Debug("Error executing request | ERR:", err)
		response.Error = errors.New("ERROR_SENDING_REQUEST -> " + err.Error())
		return &response
	}
	defer resp.Body.Close()

	response.Headers = make(map[string]string, len(resp.Header))
	for k, v := range resp.Header {
		response.Headers[k] = strings.Join(v, `,`)
	}
	response.Cookie = resp.Cookies()
	log.Debug("sendRequest | Request executed, reading response ...")
	bodyResp, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Debug("sendRequest | Unable to read response! | Err: ", err)
		response.Error = errors.New("ERROR_READING_RESPONSE -> " + err.Error())
		return &response
	}

	response.Body = bodyResp
	response.StatusCode = resp.StatusCode
	response.Error = nil
	elapsed := time.Since(start)
	response.Time = elapsed
	response.Respnse = resp
	log.Debug("sendRequest | Elapsed -> ", elapsed, " | STOP!")
	return &response
}

// BasicAuth is compute the basic auth value for the given data
func (req *Request) SetBasicAuth(username, password string) {
	req.Req.SetBasicAuth(username, password)
}

func (req *Request) SetBearerAuth(token string) error {
	if req.Req == nil {
		return errors.New("request is not initialized, call InitRequest")
	}
	req.Req.Header.Set("Authorization", "Bearer "+token)
	return nil
}

func (req *Request) AddHeader(key, value string) error {
	if req.Req == nil {
		return errors.New("request is not initialized, call InitRequest")
	}
	req.Req.Header.Set(key, value)
	return nil

}
