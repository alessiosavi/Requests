package requests

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"strings"

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

func SendRequest(url, method string, headers [][]string, jsonStr []byte) *datastructure.RequestResponse {

	// Create a custom request
	var req *http.Request
	var err error

	switch method {
	case "GET":
		req, err = http.NewRequest("GET", url, nil)
	case "POST":
		req, err = http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	case "PUT":
		req, err = http.NewRequest("PUT", url, nil)
	case "DELETE":
		req, err = http.NewRequest("DELETE", url, nil)
	default:
		zap.S().Warn("sendRequest | Unkown method -> ", method)
		return nil
	}
	if err != nil {
		zap.S().Error("AuthenticateRequest | Unable to create request! ", err)
		return nil
	}

	lenght := len(headers)
	for i := 0; i < lenght; i++ {
		//zap.S().Debug("sendRequest | Adding header: ", headers[i], " Len: ", len(headers[i]))

		key := headers[i][0]
		value := headers[i][1]
		if strings.Compare(`Authorization`, key) == 0 {
			req.Header.Add(key, value)
		} else {
			req.Header.Set(key, value)
		}
		//zap.S().Debug("sendRequest | Adding header: {", key, "|", value, "}")
	}
	zap.S().Debug("sendRequest | Executing request ...")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		zap.S().Debug("Error on response | ERR:", err)
		return nil
	}
	defer resp.Body.Close()
	//zap.S().Debug("sendRequest | Request executed, reading response ...")
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		zap.S().Error("sendRequest | Unable to read response! ", err)
		return nil
	}
	var headersResp []string
	for k, v := range resp.Header {
		headersResp = append(headersResp, utils.Join(k, `:`, strings.Join(v, `,`)))
	}
	response := datastructure.RequestResponse{}
	response.Body = body
	response.StatusCode = resp.StatusCode
	response.Headers = headersResp
	zap.S().Debug("sendRequest | Response saved")
	return &response
}
