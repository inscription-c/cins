package config

var SrvCfg = &SrvConfigs{}

// SrvConfigs is a struct that holds the configuration options for the server.
type SrvConfigs struct {
	Server struct {
		Testnet        bool   `yaml:"testnet"`
		RpcListen      string `yaml:"rpc_listen"`
		NoApi          bool   `yaml:"no_api"`
		IndexSats      string `yaml:"index_sats"`
		IndexSpendSats string `yaml:"index_spend_sats"`
		EnablePProf    bool   `yaml:"pprof"`
		Prometheus     bool   `yaml:"prometheus"`
	} `yaml:"server"`
	Chain struct {
		Url      string `yaml:"url"`
		Username string `yaml:"username"`
		Password string `yaml:"password"`
	} `yaml:"chain"`
	DB struct {
		Mysql struct {
			Addr     string `yaml:"addr"`
			User     string `yaml:"user"`
			Password string `yaml:"password"`
			DB       string `yaml:"db"`
		} `yaml:"mysql"`
	} `yaml:"db"`
	Sentry struct {
		Dsn              string  `yaml:"dsn"`
		TracesSampleRate float64 `yaml:"traces_sample_rate"`
	} `yaml:"sentry"`
	Origins []string `yaml:"origins"`
}
