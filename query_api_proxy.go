package main

import (
	"os"
	"log"
	"github.com/zoowii/query_api_proxy/proxy"
)

func main() {
	argsWithoutProg := os.Args[1:]
	if len(argsWithoutProg) < 1 {
		log.Println("please pass config yaml file path as argument")
		os.Exit(1)
		return
	}
	configFilePath := argsWithoutProg[0]
	config, err := proxy.ReadConfigFromYaml(configFilePath)
	if err != nil {
		log.Fatal(err.Error())
		os.Exit(2)
		return
	}
	proxy.StartServer(config)
}