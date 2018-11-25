package proxy

import simplejson "github.com/bitly/go-simplejson"

type WorkerResponse struct {
	Error       error
	Result      []byte
	ResultJSON  *simplejson.Json
	WorkerIndex int
	WorkerUri   string
}

func (res *WorkerResponse) IsValidJSONRpcResult() bool {
	if res.Error != nil {
		return false
	}
	if res.ResultJSON == nil {
		return false
	}
	return true
}
