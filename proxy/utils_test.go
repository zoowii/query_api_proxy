package proxy

import (
	"testing"
	"github.com/bitly/go-simplejson"
	"github.com/stretchr/testify/assert"
)

func TestCompareJSONIsSame(t *testing.T) {
	json1, err := simplejson.NewJson([]byte("{\"a\":1 , \"b\": [32, \"hi\", true, null], \"1\":{\"b\":2,\"a\":1}}"))
	assert.True(t, err == nil)
	json2, err := simplejson.NewJson([]byte("{\"a\":1, \"1\":{\"b\":2,\"a\":1}, \"b\": [32, \"hi\", true, null]}"))
	assert.True(t, err == nil)
	compareResult := CompareJSONIsSame(json1, json2)
	println("compare json is same result: ", compareResult)
	assert.True(t, compareResult)
}

func TestDigestJSONForEqual(t *testing.T) {
	json1, err := simplejson.NewJson([]byte("{\"a\":1 , \"b\": [32, \"hi\", true, null], \"1\":{\"b\":2,\"a\":1}}"))
	assert.True(t, err == nil)
	d1 := DigestJSONForEqual(json1)
	println("d1: ", string(d1))
	assert.True(t, string(d1) == "{\"1\":{\"a\":1,\"b\":2},\"a\":1,\"b\":[32,\"hi\",true,null]}")
}

func TestReadConfigFromYaml(t *testing.T) {
	config, err := ReadConfigFromYaml("../sample.yml")
	assert.True(t, err == nil)
	assert.True(t, config.Host=="0.0.0.0")
	assert.True(t, len(config.JSONRpcQueryMethods) == 2)
	assert.True(t, config.JSONRpcQueryMethods[0] == "hello")
}
