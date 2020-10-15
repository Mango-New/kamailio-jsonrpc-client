package main

import (
	"encoding/json"
	"flag"
	"os"

	"github.com/romana/rlog"
	"github.com/voipxswitch/kamailio-jsonrpc-client/internal/client"
	"github.com/voipxswitch/kamailio-jsonrpc-client/serverhttp"
	"goji.io"
)

func main() {
	rlog.Debug("debug enabled")
	confPath := "config.json"

	configFilePath := flag.String("config", "", "path to config file")
	flag.Parse()
	if *configFilePath != "" {
		confPath = *configFilePath
	}
	rlog.Infof("loading config from file [%s]", confPath)
	c, err := loadConfigFile(confPath)
	if err != nil {
		rlog.Errorf("could load config [%s]", err.Error())
		return
	}

	client, err := client.New(c.Kamailio.ServerAddr)
	if err != nil {
		rlog.Errorf("could not setup client [%s]", err.Error())
		return
	}

	// setup http server
	err = serverhttp.ListenAndServe(goji.NewMux(), c.HTTP.ListenAddr, client)
	if err != nil {
		rlog.Errorf("could not setup http server [%s]", err.Error())
		return
	}
}

// struct used to unmarshal config.json
type serviceConfig struct {
	Kamailio struct {
		ServerAddr string `json:"jsonrpcs_address"`
	} `json:"kamailio"`
	HTTP struct {
		ListenAddr string `json:"listen_address"`
	} `json:"http"`
}

func loadConfigFile(configFile string) (serviceConfig, error) {
	s := serviceConfig{}
	file, err := os.Open(configFile)
	if err != nil {
		return s, err
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&s)
	if err != nil {
		return s, err
	}
	return s, nil
}
