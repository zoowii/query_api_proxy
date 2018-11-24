package proxy

import (
	"github.com/pkg/errors"
	"strings"
)

type Config struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`

	RequestTimeoutSeconds int `yaml:"request_timeout_seconds"`

	Workers                []string `yaml:"workers"`
	JSONRpcQueryMethods    []string `yaml:"jsonrpc_query_methods"`
	OpenMethodsWhitelist   bool     `yaml:"open_methods_whitelist"`

	CacheAllJSONRpcMethods bool     `yaml:"cache_all_jsonrpc_methods"`
	CacheJSONRpcMethodsWithBlacklist bool `yaml:"cache_json_rpc_methods_with_black_list"`
	CacheJSONRpcMethodsWhitelist []string `yaml:"cache_jsonrpc_methods_whitelist"`
	CacheJSONRpcMethodsBlacklist []string `yaml:"cache_jsonrpc_methods_blacklist"`

	/**
	"fist_of_all", "most_of_all", "only_first", "only_once"
	fist_of_all: send request to all workers and response first valid json-rpc result
	most_of_all: send request to all workers and response most identical json-rpc result
	only_first: send request to workers one by one(seq by load balancing) until response first valid json-rpc result
	only_once: send request to only one worker selected by load balancing and response its result
	 */
	SelectWorkerMode string `yaml:"select_worker_mode"`


	LogPath string `yaml:"logpath"`
}

func NewConfig() *Config {
	config := new(Config)
	config.SelectWorkerMode = "only_first"
	return config
}

var firstOfAllSelectMode = "fist_of_all"
var mostOfAllSelectMode = "most_of_all"
var onlyFirstSelectMode = "only_first"
var onlyOnceSelectMode = "only_once"

func (config *Config) Validate() error {
	availableSelectWorkerModes := []string{firstOfAllSelectMode, mostOfAllSelectMode, onlyFirstSelectMode, onlyOnceSelectMode}
	if !ContainsString(availableSelectWorkerModes, config.SelectWorkerMode) {
		return errors.New("select_worker_mode config can only accept one of " + strings.Join(availableSelectWorkerModes, ", "))
	}
	return nil
}

func (config *Config) IsFirstOfAllSelectMode() bool {
	return config.SelectWorkerMode == firstOfAllSelectMode
}

func (config *Config) IsMostOfAllSelectMode() bool {
	return config.SelectWorkerMode == mostOfAllSelectMode
}

func (config *Config) IsOnlyFirstSelectMode() bool {
	return config.SelectWorkerMode == onlyFirstSelectMode
}

func (config *Config) IsOnlyOnceSelectMode() bool {
	return config.SelectWorkerMode == onlyOnceSelectMode
}
