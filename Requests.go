package requests

import (
	"bytes"
	"crypto/tls"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	utils "github.com/alessiosavi/GoUtils"
	"github.com/alessiosavi/Requests/datastructure"
)

type Request struct {
	Headers [][]string
	Req     *http.Request
	Resp    datastructure.Response
}

// CreateHeaderList is delegated to initialize a list of headers.
// Every row of the matrix contains [key,value]
func (req *Request) CreateHeaderList(headers ...string) bool {
	lenght := len(headers)

	if len(headers)%2 != 0 {
		log.Println(`Headers have to be a "key:value" list`)
		return false
	}

	req.Headers = make([][]string, lenght/2)
	counter := 0
	for i := 0; i < lenght; i += 2 {
		tmp := make([]string, 2)
		key := headers[i]
		value := headers[i+1]
		tmp[0] = key
		tmp[1] = value
		//log.Println("createHeaderList | ", i, ") Key: ", key, " Value: ", value)
		req.Headers[counter] = tmp
		counter++
	}
	//log.Println("createHeaderList | LIST: ", list)
	return true
}

// SendRequest is delegated to initialize a new HTTP request.
// If the
func (req *Request) SendRequest(url, method string, bodyData []byte, skipTLS bool) *datastructure.Response {

	// Create a custom request
	var err error
	var response datastructure.Response

	start := time.Now()

	if skipTLS {
		// Accept not trusted SSL Certificates
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		_error := fmt.Errorf("URL [%s] have not a compliant prefix, use http or https", url)
		log.Println("sendRequest | Error! ", _error)
		response.Error = _error
		return &response
	}

	method = strings.ToUpper(method)
	switch method {
	case "GET":
		req.Req, err = http.NewRequest("GET", url, nil)
	case "POST":
		// TODO: Allow post request without argument?
		if bodyData == nil {
			log.Println("sendRequest | Unable to send post data without BODY data")
			err := errors.New("CALL POST without pass BODY data")
			response.Error = err
			return &response
		}
		req.Req, err = http.NewRequest("POST", url, bytes.NewBuffer(bodyData))
	case "PUT":
		req.Req, err = http.NewRequest("PUT", url, nil)
	case "DELETE":
		req.Req, err = http.NewRequest("DELETE", url, nil)
	default:
		log.Println("sendRequest | Unkown method -> ", method)
		err := errors.New("Unkow HTTP METHOD -> " + method)
		response.Error = err
		return &response
	}
	if err != nil {
		log.Println("sendRequest | Unable to create request! | Err: ", err)
		response.Error = err
		return &response
	}
	contentLenghtPresent := false
	for i := range req.Headers {
		// log.Println("sendRequest | Adding header: ", headers[i], " Len: ", len(headers[i]))
		key := req.Headers[i][0]
		value := req.Headers[i][1]
		if strings.Compare(`Authorization`, key) == 0 {
			req.Req.Header.Add(key, value)
		} else {
			req.Req.Header.Set(key, value)
		}
		if key == "Content-Length" {
			contentLenghtPresent = true
		}
		//log.Println("sendRequest | Adding header: {", key, "|", value, "}")

	}

	if bodyData != nil && !contentLenghtPresent {
		contentLenght := len(bodyData)
		log.Println("sendRequest | Content-Lenght not provided, setting new one -> ", contentLenght)
		req.Req.Header.Add("Content-Lenght", string(contentLenght))
	}
	log.Println("sendRequest | Executing request ...")
	client := &http.Client{}
	resp, err := client.Do(req.Req)
	if err != nil {
		log.Println("Error on response | ERR:", err)
		response.Error = err
		return &response
	}
	defer resp.Body.Close()
	//log.Println("sendRequest | Request executed, reading response ...")
	bodyResp, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("sendRequest | Unable to read response! | Err: ", err)
		response.Error = err
		return &response
	}
	var headersResp []string
	for k, v := range resp.Header {
		headersResp = append(headersResp, utils.Join(k, `:`, strings.Join(v, `,`)))
	}

	response.Body = bodyResp
	response.StatusCode = resp.StatusCode
	response.Headers = headersResp
	response.Error = nil
	t := time.Now()
	elapsed := t.Sub(start)
	response.Time = elapsed
	log.Println("sendRequest | Elapsed -> ", elapsed, " | STOP!")
	return &response
}
