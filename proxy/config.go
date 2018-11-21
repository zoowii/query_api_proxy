package proxy

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

	ResponseWhenFirstGotResult bool `yaml:"response_when_got_result"`

	LogPath string `yaml:"logpath"`
}
