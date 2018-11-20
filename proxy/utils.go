package proxy

import (
	"encoding/json"
	"github.com/bitly/go-simplejson"
	"bytes"
	"sort"
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

func DigestJSONForEqual(jsonVal *simplejson.Json) string {
	if jsonVal == nil {
		return "nil"
	}
	encoded, err := jsonVal.Encode()
	if err != nil {
		return "error"
	}
	encodedStr := string(encoded)
	if len(encodedStr) < 1 || (encodedStr[0] != '{' && encodedStr[0] != '[') {
		return encodedStr
	}
	if encodedStr[0] == '[' {
		jsonArray := jsonVal.MustArray()
		var digestBuffer bytes.Buffer
		digestBuffer.WriteString("[")
		for idx, _ := range jsonArray {
			if idx > 0 {
				digestBuffer.WriteString(",")
			}
			itemJson := jsonVal.GetIndex(idx)
			digestBuffer.WriteString(DigestJSONForEqual(itemJson))
		}
		digestBuffer.WriteString("]")
		return digestBuffer.String()
	} else if encodedStr[0] == '{' {
		jsonMap := jsonVal.MustMap()
		var digestBuffer bytes.Buffer
		digestBuffer.WriteString("{")
		keys := make([]string, 0)
		for k, _ := range jsonMap {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for idx, key := range keys {
			if idx > 0 {
				digestBuffer.WriteString(",")
			}
			keyEncode, err := json.Marshal(key)
			if err != nil {
				digestBuffer.WriteString("\"error\":\"error\"")
				continue
			}
			item := jsonVal.Get(key)
			digestBuffer.WriteString(string(keyEncode))
			digestBuffer.WriteString(":")
			digestBuffer.WriteString(DigestJSONForEqual(item))
		}
		digestBuffer.WriteString("}")
		return digestBuffer.String()
	} else {
		return "error"
	}
}

// whether json a and json b have the same value
func CompareJSONIsSame(a *simplejson.Json, b *simplejson.Json) bool {
	if a == b {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return DigestJSONForEqual(a) == DigestJSONForEqual(b)
}