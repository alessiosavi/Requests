package requests

import (
	"testing"

	"github.com/alessiosavi/Requests/datastructure"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestCreateHeaderList(t *testing.T) {
	// Create a simple headers
	headersKey := `Content-Type`
	headersValue := `application/json`

	contentTypeHeaders := CreateHeaderList(headersKey, headersValue)

	if len(contentTypeHeaders) != 1 {
		t.Error("size error!")
	}

	if len(contentTypeHeaders[0]) != 2 {
		t.Error("key value headers size mismatch")
	}

	headersKeyTest := contentTypeHeaders[0][0]
	headersValueTest := contentTypeHeaders[0][1]
	if headersKey != headersKeyTest {
		t.Error("Headers key mismatch!")
	}
	if headersValue != headersValueTest {
		t.Error("Headers value mismatch!")
	}
}

func initZapLog() *zap.Logger {
	config := zap.NewDevelopmentConfig()
	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	logger, _ := config.Build()
	return logger
}

func TestSendRequest(t *testing.T) {

	// loggerMgr := initZapLog()
	// zap.ReplaceGlobals(loggerMgr)
	// defer loggerMgr.Sync() // flushes buffer, if any
	// logger := loggerMgr.Sugar()
	// logger.Debug("START")
	var resp *datastructure.RequestResponse
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

func makeBadRequestURL1() *datastructure.RequestResponse {
	return SendRequest("tcp://google.it", "GET", nil, nil, true)
}
func makeBadRequestURL2() *datastructure.RequestResponse {
	return SendRequest("google.it", "GET", nil, nil, true)
}
func makeOKRequestURL3() *datastructure.RequestResponse {
	return SendRequest("https://google.it", "GET", nil, nil, true)
}

func dumpResponse(resp *datastructure.RequestResponse, t *testing.T) {
	t.Log(string(resp.Body))
	t.Log(resp.StatusCode)
	t.Log(resp.Headers)
	t.Log(resp.Error)
	t.Log(resp)
}
