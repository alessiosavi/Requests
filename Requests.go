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
	"time"

	"github.com/alessiosavi/Requests/datastructure"
	"github.com/onrik/logrus/filename"
	log "github.com/sirupsen/logrus"
)

// AllowedMethod rappresent the HTTP method allowed in the request
var allowedMethod []string = []string{"GET", "POST", "HEAD", "PUT", "DELETE", "OPTIONS"}

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

func (req *Request) methodIsAllowed(method string) bool {
	for i := range allowedMethod {
		if method == allowedMethod[i] {
			req.Method = method
			return true
		}
	}
	return false
}

type Request struct {
	Req     *http.Request
	Method  string                 // HTTP method of the request
	Url     string                 // URL where send the request
	Data    []byte                 // BODY in case of POST, ARGS in case of GET
	Headers [][]string             // List of headers to send in the request
	Resp    datastructure.Response // Struct for save the response
	SkipTLS bool                   // Skip or not the SSL certificate validation
}

// CreateHeaderList is delegated to initialize a list of headers.
// Every row of the matrix contains [key,value]
func (req *Request) CreateHeaderList(headers ...string) bool {
	if headers == nil {
		return false
	}

	length := len(headers)

	if len(headers)%2 != 0 {
		log.Debug(`Headers have to be a "key:value" list`)
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
		//log.Debug("createHeaderList | ", i, ") Key: ", key, " Value: ", value)
		req.Headers[counter] = tmp
		counter++
	}
	//log.Debug("createHeaderList | LIST: ", list)
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
		if !strings.Contains(req.Url, "?") {
			// Overwrite the "/" with the provided params
			if strings.HasSuffix(req.Url, "/") {
				index := strings.LastIndex(req.Url, "/")
				req.Url = req.Url[:index]
			}
			req.Url += "?" + args
		} else { // adding additional parameter to the one provided in the URL
			req.Url += "&" + args
		}
	}
}

// ParallelRequest is delegated to run the given list of request in parallel, using N request at each time

func ParallelRequest(reqs []Request, N int) []datastructure.Response {
	var wg sync.WaitGroup
	var results []datastructure.Response = make([]datastructure.Response, len(reqs))
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

func InitRequest(url, method string, bodyData []byte, headers []string, skipTLS bool) (*Request, error) {
	var err error
	var req Request

	req.SkipTLS = skipTLS
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		err = errors.New("PREFIX_URL_NOT_VALID")
		log.Debug("sendRequest | Error! ", err, " URL: ", url)
		return nil, err
	}

	method = strings.ToUpper(method)

	// Validate HTTP method
	if !req.methodIsAllowed(method) {
		log.Debug("sendRequest | Method [" + method + "] is not allowed!")
		err = errors.New("METHOD_NOT_ALLOWED")
		return nil, err
	}

	req.Url = url
	req.Data = bodyData

	req.CreateHeaderList(headers...)
	switch req.Method {
	case "GET":
		req.initGetRequest()
		req.Req, err = http.NewRequest(req.Method, req.Url, nil)
	case "POST":
		req.initPostRequest()
		req.Req, err = http.NewRequest(req.Method, req.Url, bytes.NewReader(req.Data))
	case "PUT":
		req.Req, err = http.NewRequest(req.Method, req.Url, nil)
	case "DELETE":
		req.Req, err = http.NewRequest(req.Method, req.Url, nil)
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

	// If data are present and content lenght was not specified (only for POST)
	if req.Method == "POST" && bodyData != nil && !contentlengthPresent {
		contentlength := len(bodyData)
		log.Debug("sendRequest | Content-length not provided, setting new one -> ", contentlength)
		req.Req.Header.Add("Content-Length", strconv.Itoa(contentlength))
	}

	return &req, err
}

// ExecuteRequest is delegated to run a previously allocated request.
func (req *Request) ExecuteRequest() datastructure.Response {
	var response datastructure.Response
	var start time.Time = time.Now()
	var err error

	var tr *http.Transport
	if req.SkipTLS {
		// Accept not trusted SSL Certificates
		tr = &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	} else {
		tr = &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: false}}
	}

	log.Debug("sendRequest | Executing request ...")
	client := &http.Client{Transport: tr}
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
	t := time.Now()
	elapsed := t.Sub(start)
	response.Time = elapsed
	// log.Debug("sendRequest | Elapsed -> ", elapsed, " | STOP!")
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

	if skipTLS {
		// Accept not trusted SSL Certificates
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

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

	req.Url = url
	req.Data = bodyData

	switch req.Method {
	case "GET":
		req.initGetRequest()
		req.Req, err = http.NewRequest(req.Method, req.Url, nil)
	case "POST":
		req.initPostRequest()
		req.Req, err = http.NewRequest(req.Method, req.Url, bytes.NewReader(req.Data))
	case "PUT":
		req.Req, err = http.NewRequest(req.Method, req.Url, nil)
	case "DELETE":
		req.Req, err = http.NewRequest(req.Method, req.Url, nil)
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

	if req.Method == "POST" && bodyData != nil && !contentlengthPresent {
		contentlength := len(bodyData)
		// log.Debug("sendRequest | Content-length not provided, setting new one -> ", contentlength)
		req.Req.Header.Add("Content-Length", strconv.Itoa(contentlength))
	}
	// log.Debug("sendRequest | Executing request ...")
	client := &http.Client{}
	resp, err := client.Do(req.Req)
	if err != nil {
		log.Debug("Error executing request | ERR:", err)
		response.Error = errors.New("ERROR_SENDING_REQUEST -> " + err.Error())
		return &response
	}
	defer resp.Body.Close()
	//log.Debug("sendRequest | Request executed, reading response ...")
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
	t := time.Now()
	elapsed := t.Sub(start)
	response.Time = elapsed
	// log.Debug("sendRequest | Elapsed -> ", elapsed, " | STOP!")
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
