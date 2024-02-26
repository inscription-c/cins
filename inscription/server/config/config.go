package config

var SrvCfg = &SrvConfigs{}

// SrvConfigs is a struct that holds the configuration options for the server.
type SrvConfigs struct {
	Testnet        bool   `yaml:"testnet"`
	RpcListen      string `yaml:"rpc_listen"`
	Username       string `yaml:"username"`
	Password       string `yaml:"password"`
	RpcConnect     string `yaml:"rpc_connect"`
	NoApi          bool   `yaml:"no_api"`
	IndexSats      string `yaml:"index_sats"`
	IndexSpendSats string `yaml:"index_spend_sats"`
	Mysql          struct {
		Addr     string `yaml:"addr"`
		User     string `yaml:"user"`
		Password string `yaml:"password"`
		DB       string `yaml:"db"`
	} `yaml:"mysql"`
	EnablePProf bool `yaml:"pprof"`
	Sentry      struct {
		Dsn              string  `yaml:"dsn"`
		TracesSampleRate float64 `yaml:"traces_sample_rate"`
	} `yaml:"sentry"`
	Origins []string `yaml:"origins"`
}
