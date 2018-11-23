package proxy

import (
	"github.com/bitly/go-simplejson"
	"os"
)

const JSONRPC_PARSE_ERROR_CODE = -32700
const JSONRPC_INVALID_REQUEST_ERROR_CODE = -32600
const JSONRPC_METHOD_NOT_FOUND_ERROR_CODE = -32601
const JSONRPC_INVALID_REQUEST_PARAMS_ERROR_CODE = -32602
const JSONRPC_INTERNAL_ERROR_CODE = -32603
const JSONRPC_SERVER_ERROR_CODE_START = -32000 // [-32000, -32099]

func MakeJSONRpcSuccessResponse(id interface{}, result interface{}) ([]byte, error) {
	res := simplejson.New()
	res.Set("jsonrpc", "2.0")
	res.Set("id", id)
	res.Set("result", result)
	resBytes, err := res.Encode()
	return resBytes, err
}

func MakeJSONRpcErrorResponse(id interface{}, code int, message string, data interface{}) ([]byte, error) {
	res := simplejson.New()
	res.Set("jsonrpc", "2.0")
	res.Set("id", id)
	errorObj := simplejson.New()
	errorObj.Set("code", code)
	errorObj.Set("message", message)
	errorObj.Set("data", data)
	res.Set("error", errorObj)
	resBytes, err := res.Encode()
	return resBytes, err
}

func IsSuccessJSONRpcResponse(result *simplejson.Json) bool {
	if result == nil {
		return false
	}
	_, ok := result.CheckGet("error")
	if ok {
		return false
	} else {
		return true
	}
}

func CheckFileExists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}