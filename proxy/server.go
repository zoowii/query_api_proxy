package proxy

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/zoowii/query_api_proxy/cache"

	"github.com/bitly/go-simplejson"
	"github.com/zoowii/betterjson"
	"gopkg.in/yaml.v2"
	"sync/atomic"
)

func ReadConfigFromYaml(yamlConfigFilePath string) (*Config, error) {
	conf := NewConfig()
	yamlFile, err := ioutil.ReadFile(yamlConfigFilePath)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(yamlFile, conf)
	if err != nil {
		return nil, err
	}
	return conf, nil
}

func writeErrorToJSONRpcResponse(w http.ResponseWriter, id interface{}, errorCode int, errMsg string) {
	resBytes, err := MakeJSONRpcErrorResponse(id, errorCode, errMsg, nil)
	if err != nil {
		w.Write([]byte(err.Error()))
	} else {
		w.Header().Set("Content-Type", "application/json")
		w.Write(resBytes)
	}
}

func writeResultToJSONRpcResponse(w http.ResponseWriter, id interface{}, result interface{}) {
	resBytes, err := MakeJSONRpcSuccessResponse(id, result)
	if err != nil {
		w.Write([]byte(err.Error()))
	} else {
		w.Header().Set("Content-Type", "application/json")
		w.Write(resBytes)
	}
}

func writeDirectlyToResponse(w http.ResponseWriter, data []byte) {
	w.Write(data)
}

type WorkerResponse struct {
	Error       error
	Result      []byte
	ResultJSON  *simplejson.Json
	WorkerIndex int
	WorkerUri   string
}

func isNeedCacheMethod(config *Config, rpcReqMethod string) bool {
	if config.CacheAllJSONRpcMethods {
		return true
	}
	if config.CacheJSONRpcMethodsWithBlacklist {
		for _, m := range config.CacheJSONRpcMethodsBlacklist {
			if m == rpcReqMethod {
				return false
			}
		}
		return true
	}
	return false
}

func useWorkerToProvideService(config *Config, workerIndex int, workerUri string, rpcReqMethod string, reqBody []byte) *WorkerResponse {
	res := new(WorkerResponse)
	res.WorkerIndex = workerIndex
	res.WorkerUri = workerUri

	cache1Key := workerUri
	cache2Key := string(reqBody)
	// because of there will not be any '^' in workerUri, so join cache1Key and cache2Key by '^'
	cacheKey := cache1Key + "^" + cache2Key

	if isNeedCacheMethod(config, rpcReqMethod) {
		if cacheValue, ok := cache.Get(cacheKey); ok {
			resultBytes := cacheValue.([]byte)
			resultJSON, jsonErr := simplejson.NewJson(resultBytes)
			if jsonErr == nil {
				res.Result = resultBytes
				res.ResultJSON = resultJSON
				// TODO: digest result json and when got > 1/2 same results, just break the loop
				return res
			}
		}
	}

	workerHttpRes, workerResErr := http.Post(workerUri, "application/json", bytes.NewReader(reqBody))

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
				// TODO: digest result json and when got > 1/2 same results, just break the loop
				if isNeedCacheMethod(config, rpcReqMethod) || IsSuccessJSONRpcResponse(resultJSON) {
					cacheValue := readBytes
					cache.SetWithDefaultExpire(cacheKey, cacheValue)
				}
			}
		}
	}
	return res
}

func selectWorkerToProvideService(config *Config, triedWorkerUris []string) (workerUri string, err error) {
	// TODO
	return "", nil
}

var workerLoadBalanceIndex uint32 = 0

// teturns the order of workers according to the mode in the configuration
func getWorkersSequenceBySelectMode(config *Config, workerUris []string) []string {
	if config.IsMostOfAllSelectMode() || config.IsFirstOfAllSelectMode() {
		return workerUris
	} else if config.IsOnlyFirstSelectMode() ||  config.IsOnlyOnceSelectMode() {
		firstIdx := atomic.AddUint32(&workerLoadBalanceIndex, 1)
		firstIdx = firstIdx % uint32(len(workerUris))
		newSeq := []string{workerUris[firstIdx]}
		beforeWorkers := workerUris[0:firstIdx]
		afterWorkers := workerUris[firstIdx+1:]
		newSeq = append(newSeq, beforeWorkers...)
		newSeq = append(newSeq, afterWorkers...)
		return newSeq
	} else {
		panic("not supported config select_worker_mode")
		return nil
	}
}

// TODO: use jsonrpcmethods whitelist if enabled
// TODO: fault handler
// TODO: rate limit
func StartServer(config *Config) {
	if config.LogPath=="" {
		config.LogPath = "./query_api_proxy.log"
	}
	logger, err := NewLogger(config.LogPath)
	if err != nil {
		panic("error happen when open log " + err.Error())
		return
	}
	defer logger.Close()
	proxyHandlerFunc := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
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
		var rpcReqMethod string = ""
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
			method, err := reqBodyJSON.Get("method").String()
			if err == nil {
				rpcReqMethod = method
			} else {
				writeErrorToJSONRpcResponse(w, 1, JSONRPC_INVALID_REQUEST_ERROR_CODE, err.Error())
				return
			}
		}

		responsesChannel := make(chan *WorkerResponse, len(config.Workers))
		// TODO: workers health check
		workerUris := getWorkersSequenceBySelectMode(config, config.Workers)
		if config.IsOnlyFirstSelectMode() {
			// TODO: send request to workers one by one. now just send to all workers
			for workerIndex, workerUri := range workerUris {
				go func(workerIndex int, workerUri string) {
					res := useWorkerToProvideService(config, workerIndex, workerUri, rpcReqMethod, reqBody)
					responsesChannel <- res
				}(workerIndex, workerUri)
			}
		} else {
			for workerIndex, workerUri := range workerUris {
				go func(workerIndex int, workerUri string) {
					res := useWorkerToProvideService(config, workerIndex, workerUri, rpcReqMethod, reqBody)
					responsesChannel <- res
				}(workerIndex, workerUri)
				if config.IsOnlyOnceSelectMode() {
					break
				}
			}
		}
		timeout := false
		breakIterWorkerResponses := false
		workerResponses := make([]*WorkerResponse, 0)
		for i := 0; i < len(workerUris); i++ {
			if timeout {
				break
			}
			select {
			case res := <-responsesChannel:
				workerResponses = append(workerResponses, res)
				if config.IsOnlyOnceSelectMode() {
					breakIterWorkerResponses = true
					break
				}
				if config.IsOnlyFirstSelectMode() && res.ResultJSON != nil {
					breakIterWorkerResponses = true
					break
				}
				if !config.IsMostOfAllSelectMode() && res.ResultJSON != nil {
					breakIterWorkerResponses = true
				}
			case <-time.After(time.Duration(config.RequestTimeoutSeconds) * time.Second):
				timeout = true
			}
			if breakIterWorkerResponses {
				break
			}
		}
		// compare workerResponses to select most same responses
		hasSomeErrorInWorkerResponses := false
		if (config.IsFirstOfAllSelectMode() || config.IsMostOfAllSelectMode()) || len(workerResponses) < len(config.Workers) {
			hasSomeErrorInWorkerResponses = true
		}
		if len(workerResponses) < 1 {
			hasSomeErrorInWorkerResponses = true
		}
		type WorkerResponseSameGroup struct {
			ResultJSON  *simplejson.Json
			ResultBytes []byte
			Count       int
		}
		if !config.IsMostOfAllSelectMode() && len(workerResponses) > 0 {
			// find first not empty result json and final response
			for _, workerRes := range workerResponses {
				if workerRes.ResultJSON != nil {
					writeDirectlyToResponse(w, workerRes.Result)
					return
				}
			}
		}
		var sameWorkerResponseGroups = make(map[string]*WorkerResponseSameGroup, 0)
		var maxCountGroup *WorkerResponseSameGroup = nil
		for _, workerRes := range workerResponses {
			if workerRes.ResultJSON == nil {
				hasSomeErrorInWorkerResponses = true
				continue
			}
			resultJSONDigest := betterjson.FromNotEmptySimpleJson(workerRes.ResultJSON).DigestJSONForEqual()
			var group *WorkerResponseSameGroup
			var foundGroup bool
			if group, foundGroup = sameWorkerResponseGroups[resultJSONDigest]; foundGroup {
				group.Count += 1
			} else {
				group = new(WorkerResponseSameGroup)
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
			logger.Print(errMsg)
			writeErrorToJSONRpcResponse(w, rpcReqId, JSONRPC_INTERNAL_ERROR_CODE, "no responses until timeout")
			return
		}
		if len(sameWorkerResponseGroups) > 1 {
			hasSomeErrorInWorkerResponses = true
			logger.Printf("workers send some distinct responses when dispath request %s\n", string(reqBody))
		}
		if hasSomeErrorInWorkerResponses {
			logger.Printf("some errors in worker responses when dispath request %s\n", string(reqBody))
		}
		writeDirectlyToResponse(w, maxCountGroup.ResultBytes)
	})
	var logRequest = func (handler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			timer1 := time.NewTimer(time.Millisecond)
			logger.Printf("%s %s %s\n", r.RemoteAddr, r.Method, r.URL)
			handler.ServeHTTP(w, r)
			timer1.Stop()
			usedTime := <- timer1.C
			logger.Printf("using %.2f seconds\n", (float64(usedTime.Nanosecond())*1.0/1000000000))
		})
	}
	_ = logRequest
	s := &http.Server{
		Addr:           fmt.Sprintf("%s:%d", config.Host, config.Port),
		Handler:        proxyHandlerFunc, // logRequest(proxyHandlerFunc),
		ReadTimeout:    50 * time.Second,
		WriteTimeout:   100 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	s.SetKeepAlivesEnabled(false)
	logger.Printf("starting server at %s:%d\n", config.Host, config.Port)
	logger.Fatal(s.ListenAndServe())
}
