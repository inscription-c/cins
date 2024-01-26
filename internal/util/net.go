package util

import (
	"github.com/inscription-c/insc/internal/cfgutil"
	"net"
)

func DisableTls(rpcConnect, port string) (bool, error) {
	rpcConnect, err := cfgutil.NormalizeAddress(rpcConnect, port)
	if err != nil {
		return false, nil
	}
	RPCHost, _, err := net.SplitHostPort(rpcConnect)
	if err != nil {
		return false, nil
	}
	localhostListeners := map[string]struct{}{
		"localhost": {},
		"127.0.0.1": {},
		"::1":       {},
	}

	disableTls := false
	if _, ok := localhostListeners[RPCHost]; ok {
		disableTls = true
	}
	return disableTls, nil
}
