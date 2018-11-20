package proxy

import (
	"io/ioutil"
	"log"
	"gopkg.in/yaml.v2"
	"net/http"
	"fmt"
	"time"
	"bytes"
	"github.com/bitly/go-simplejson"
)

func ReadConfigFromYaml(yamlConfigFilePath string) (*Config, error) {
	conf := new(Config)
	yamlFile, err := ioutil.ReadFile(yamlConfigFilePath)
	log.Println("yamlFile:", yamlFile)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(yamlFile, conf)
	// err = yaml.Unmarshal(yamlFile, &resultMap)
	if err != nil {
		return nil, err
	}
	log.Println("conf", conf)
	return conf, nil
}

func writeErrorToJSONRpcResponse(w http.ResponseWriter, id interface{}, errorCode int, errMsg string) {
	resBytes, err := MakeJSONRpcErrorResponse(id, errorCode, errMsg, nil)
	if err != nil {
		w.Write([]byte(err.Error()))
	} else {
		w.Write(resBytes)
	}
}

func writeResultToJSONRpcResponse(w http.ResponseWriter, id interface{}, result interface{}) {
	resBytes, err := MakeJSONRpcSuccessResponse(id, result)
	if err != nil {
		w.Write([]byte(err.Error()))
	} else {
		w.Write(resBytes)
	}
}

type WorkerResponse struct {
	Error error
	Result []byte
	ResultJSON *simplejson.Json
	WorkerIndex int
	WorkerUri string
}

func StartServer(config *Config) {
	proxyHandlerFunc := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if(r.Method != http.MethodPost) {
			// only support POST json-rpc now
			writeErrorToJSONRpcResponse(w, 1, JSONRPC_PARSE_ERROR_CODE, "only support POST JSON-RPC now")
			return
		}
		defer r.Body.Close()
		reqBody, err := ioutil.ReadAll(r.Body)
		if err != nil {
			writeErrorToJSONRpcResponse(w, 1, JSONRPC_INVALID_REQUEST_ERROR_CODE, err.Error())
			return
		}
		var rpcReqId interface{} = 1
		reqBodyJSON, err := simplejson.NewJson(reqBody)
		if err == nil {
			tryGetReqId, err := reqBodyJSON.Get("id").Int()
			if err == nil {
				rpcReqId = tryGetReqId
			} else {
				tryGetReqId, err := reqBodyJSON.Get("id").String()
				if err == nil {
					rpcReqId = tryGetReqId
				}
			}
		}
		responsesChannel := make(chan *WorkerResponse, len(config.Workers))
		for workerIndex, workerUri := range config.Workers {
			go func(workerIndex int, workerUri string) {
				workerHttpRes, workerResErr := http.Post(workerUri, r.Header.Get("Content-Type"), bytes.NewReader(reqBody))
				res := new(WorkerResponse)
				res.WorkerIndex = workerIndex
				res.WorkerUri = workerUri
				if workerResErr != nil {
					res.Error = workerResErr
				} else {
					defer workerHttpRes.Body.Close()
					readBytes, readErr := ioutil.ReadAll(workerHttpRes.Body)
					if readErr != nil {
						res.Error = readErr
					} else {
						res.Result = readBytes
						resultJSON, jsonErr := simplejson.NewJson(readBytes)
						if jsonErr == nil {
							res.ResultJSON = resultJSON
						}
					}
				}
			}(workerIndex, workerUri)
		}
		timeout := false
		workerResponses := make([]*WorkerResponse, 0)
		for i:=0;i<len(config.Workers);i++ {
			if timeout {
				break
			}
			select {
			case res := <-responsesChannel:
				workerResponses = append(workerResponses, res)
			case <-time.After(time.Duration(config.RequestTimeoutSeconds) * time.Second):
				timeout = true
			}
		}
		// compare workerResponses to select most same responses
		hasSomeErrorInWorkerResponses := false
		if len(workerResponses) < len(config.Workers) {
			hasSomeErrorInWorkerResponses = true
		}
		if len(workerResponses) < 1 {
			hasSomeErrorInWorkerResponses = true
		}
		type WorkerResponseSameGroup struct {
			ResultJSON *simplejson.Json
			ResultBytes []byte
			Count int
		}
		var sameWorkerResponseGroups = make(map[string]*WorkerResponseSameGroup, 0)
		var maxCountGroup *WorkerResponseSameGroup = nil
		for _, workerRes := range workerResponses {
			if workerRes.ResultJSON == nil {
				hasSomeErrorInWorkerResponses = true
				continue
			}
			resultJSONDigest := DigestJSONForEqual(workerRes.ResultJSON)
			var group *WorkerResponseSameGroup
			var foundGroup bool
			if group, foundGroup = sameWorkerResponseGroups[resultJSONDigest]; foundGroup {
				group.Count += 1
			} else {
				group := new(WorkerResponseSameGroup)
				group.ResultJSON = workerRes.ResultJSON
				group.ResultBytes = workerRes.Result
				group.Count = 1
				sameWorkerResponseGroups[resultJSONDigest] = group
			}
			if maxCountGroup == nil {
				maxCountGroup = group
			} else {
				if group.Count > maxCountGroup.Count {
					maxCountGroup = group
				}
			}
		}
		if len(sameWorkerResponseGroups) < 1 || maxCountGroup == nil {
			hasSomeErrorInWorkerResponses = true
			errMsg := fmt.Sprintf("workers send zero responses when dispatch request %s\n", string(reqBody))
			log.Print(errMsg)
			writeErrorToJSONRpcResponse(w, rpcReqId, JSONRPC_INTERNAL_ERROR_CODE, "no responses until timeout")
			return
		}
		if len(sameWorkerResponseGroups)>1 {
			hasSomeErrorInWorkerResponses = true
			log.Printf("workers send some distinct responses when dispath request %s\n", string(reqBody))
		}
		if hasSomeErrorInWorkerResponses {
			log.Printf("some errors in worker responses when dispath request %s\n", string(reqBody))
		}
		writeResultToJSONRpcResponse(w, rpcReqId, maxCountGroup.ResultBytes)
	})
	s := &http.Server{
		Addr:           fmt.Sprintf("%s:%d", config.Host, config.Port),
		Handler:        proxyHandlerFunc,
		ReadTimeout:    50 * time.Second,
		WriteTimeout:   100 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	s.SetKeepAlivesEnabled(false)
	log.Printf("starting server at %s:%d", config.Host, config.Port)
	log.Fatal(s.ListenAndServe())
}