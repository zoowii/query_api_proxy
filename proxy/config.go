package proxy

type Config struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`

	RequestTimeoutSeconds int `yaml:"request_timeout_seconds"`

	Workers                []string `yaml:"workers"`
	JSONRpcQueryMethods    []string `yaml:"jsonrpc_query_methods"`
	OpenMethodsWhitelist   bool     `yaml:"open_methods_whitelist"`
	CacheAllJSONRpcMethods bool     `yaml:"cache_all_jsonrpc_methods"`

	ResponseWhenFirstGotResult bool `yaml:"response_when_got_result"`

	LogPath string `yaml:"logpath"`
}
