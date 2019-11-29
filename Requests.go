package requests

import (
	"bytes"
	"crypto/tls"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	utils "github.com/alessiosavi/GoUtils"
	"github.com/alessiosavi/Requests/datastructure"
	"go.uber.org/zap"
)

// CreateHeaderList is delegated to initialize a list of headers.
// Every row of the matrix contains [key,value]
func CreateHeaderList(headers ...string) [][]string {
	var list [][]string
	lenght := len(headers)

	list = make([][]string, lenght/2)
	counter := 0
	for i := 0; i < lenght; i += 2 {
		tmp := make([]string, 2)
		key := headers[i]
		value := headers[i+1]
		tmp[0] = key
		tmp[1] = value
		//zap.S().Debug("createHeaderList | ", i, ") Key: ", key, " Value: ", value)
		list[counter] = tmp
		counter++
	}
	//zap.S().Debug("createHeaderList | LIST: ", list)
	return list
}

// SendRequest is delegated to initialize a new HTTP request.
// If the
func SendRequest(url, method string, headers [][]string, bodyData []byte, skipTLS bool) *datastructure.RequestResponse {

	// Create a custom request
	var req *http.Request
	var err error
	var response datastructure.RequestResponse

	start := time.Now()

	if skipTLS {
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		_error := fmt.Errorf("URL [%s] have not a compliant prefix, use http or https", url)
		zap.S().Debug("sendRequest | ", _error)
		response.Error = _error
		return &response
	}

	method = strings.ToUpper(method)
	switch method {
	case "GET":
		req, err = http.NewRequest("GET", url, nil)
	case "POST":
		// TODO: Allow post request without argument?
		if bodyData == nil {
			zap.S().Error("sendRequest | Unable to send post data without BODY data")
			err := errors.New("CALL POST without pass BODY data")
			response.Error = err
			return &response
		}
		req, err = http.NewRequest("POST", url, bytes.NewBuffer(bodyData))
	case "PUT":
		req, err = http.NewRequest("PUT", url, nil)
	case "DELETE":
		req, err = http.NewRequest("DELETE", url, nil)
	default:
		zap.S().Warn("sendRequest | Unkown method -> ", method)
		err := errors.New("Unkow HTTP METHOD -> " + method)
		response.Error = err
		return &response
	}
	if err != nil {
		zap.S().Error("sendRequest | Unable to create request! | Err: ", err)
		response.Error = err
		return &response
	}
	contentLenghtPresent := false
	for i := range headers {
		// zap.S().Debug("sendRequest | Adding header: ", headers[i], " Len: ", len(headers[i]))
		key := headers[i][0]
		value := headers[i][1]
		if strings.Compare(`Authorization`, key) == 0 {
			req.Header.Add(key, value)
		} else {
			req.Header.Set(key, value)
		}
		if key == "Content-Length" {
			contentLenghtPresent = true
		}
		//zap.S().Debug("sendRequest | Adding header: {", key, "|", value, "}")

	}

	if bodyData != nil && !contentLenghtPresent {
		contentLenght := len(bodyData)
		zap.S().Debug("sendRequest | Content-Lenght not provided, setting new one -> ", contentLenght)
		req.Header.Add("Content-Lenght", string(contentLenght))
	}
	zap.S().Debug("sendRequest | Executing request ...")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		zap.S().Debug("Error on response | ERR:", err)
		response.Error = err
		return &response
	}
	defer resp.Body.Close()
	//zap.S().Debug("sendRequest | Request executed, reading response ...")
	bodyResp, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		zap.S().Error("sendRequest | Unable to read response! | Err: ", err)
		response.Error = err
		return &response
	}
	var headersResp []string
	for k, v := range resp.Header {
		headersResp = append(headersResp, utils.Join(k, `:`, strings.Join(v, `|`)))
	}

	response.Body = bodyResp
	response.StatusCode = resp.StatusCode
	response.Headers = headersResp
	response.Error = nil
	t := time.Now()
	elapsed := t.Sub(start)
	zap.S().Debug("sendRequest | Elapsed -> ", elapsed, " | STOP!")
	return &response
}
