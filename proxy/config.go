package proxy

type Config struct {
	Host string `yaml:"host"`
	Port int `yaml:"port"`

	RequestTimeoutSeconds int `yaml:"request_timeout_seconds"`

	Workers []string `yaml:"workers"`
	JSONRpcQueryMethods []string `yaml:"jsonrpc_query_methods"`

	LogPath string `yaml:"logpath"`
}
