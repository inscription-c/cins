package wallet

import (
	"errors"
	"github.com/inscription-c/cins/pkg/wallet/rpc/legacyrpc"
	"github.com/inscription-c/cins/pkg/wallet/wallet"
	"github.com/inscription-c/cins/wallet/log"
	"net"
	"runtime"
	"strings"
)

// startRPCServices starts all RPC servers provided by the wallet.
func startRPCServers(walletLoader *wallet.Loader) (*legacyrpc.Server, error) {
	var (
		legacyServer *legacyrpc.Server
		legacyListen = net.Listen
	)

	if len(cfg.LegacyRPCListeners) != 0 {
		listeners := makeListeners(cfg.LegacyRPCListeners, legacyListen)
		if len(listeners) == 0 {
			err := errors.New("failed to create listeners for legacy RPC server")
			return nil, err
		}
		opts := legacyrpc.Options{
			Username:            cfg.Username,
			Password:            cfg.Password,
			MaxPOSTClients:      cfg.LegacyRPCMaxClients,
			MaxWebsocketClients: cfg.LegacyRPCMaxWebsockets,
		}
		legacyServer = legacyrpc.NewServer(&opts, walletLoader, listeners)
	}

	if legacyServer == nil {
		return nil, errors.New("no suitable RPC services can be started")
	}

	return legacyServer, nil
}

type listenFunc func(net string, laddr string) (net.Listener, error)

// makeListeners splits the normalized listen addresses into IPv4 and IPv6
// addresses and creates new net.Listeners for each with the passed listen func.
// Invalid addresses are logged and skipped.
func makeListeners(normalizedListenAddrs []string, listen listenFunc) []net.Listener {
	ipv4Addrs := make([]string, 0, len(normalizedListenAddrs)*2)
	ipv6Addrs := make([]string, 0, len(normalizedListenAddrs)*2)
	for _, addr := range normalizedListenAddrs {
		host, _, err := net.SplitHostPort(addr)
		if err != nil {
			log.Log.Errorf("SplitHostPort failed for listener %s: %v", addr, err)
			continue
		}

		// Empty host or host of * on plan9 is both IPv4 and IPv6.
		if host == "" || (host == "*" && runtime.GOOS == "plan9") {
			ipv4Addrs = append(ipv4Addrs, addr)
			ipv6Addrs = append(ipv6Addrs, addr)
			continue
		}

		// Remove the IPv6 zone from the host, if present.  The zone
		// prevents ParseIP from correctly parsing the IP address.
		// ResolveIPAddr is intentionally not used here due to the
		// possibility of leaking a DNS query over Tor if the host is a
		// hostname and not an IP address.
		zoneIndex := strings.Index(host, "%")
		if zoneIndex != -1 {
			host = host[:zoneIndex]
		}

		ip := net.ParseIP(host)
		switch {
		case ip == nil:
			log.Log.Warnf("`%s` is not a valid IP address", host)
		case ip.To4() == nil:
			ipv6Addrs = append(ipv6Addrs, addr)
		default:
			ipv4Addrs = append(ipv4Addrs, addr)
		}
	}
	listeners := make([]net.Listener, 0, len(ipv6Addrs)+len(ipv4Addrs))
	for _, addr := range ipv4Addrs {
		listener, err := listen("tcp4", addr)
		if err != nil {
			log.Log.Warnf("Can't listen on %s: %v", addr, err)
			continue
		}
		listeners = append(listeners, listener)
	}
	for _, addr := range ipv6Addrs {
		listener, err := listen("tcp6", addr)
		if err != nil {
			log.Log.Warnf("Can't listen on %s: %v", addr, err)
			continue
		}
		listeners = append(listeners, listener)
	}
	return listeners
}

// startWalletRPCServices associates each of the (optionally-nil) RPC servers
// with a wallet to enable remote wallet access.  For the GRPC server, this
// registers the WalletService service, and for the legacy JSON-RPC server it
// enables methods that require a loaded wallet.
func startWalletRPCServices(wallet *wallet.Wallet, legacyServer *legacyrpc.Server) {
	if legacyServer != nil {
		legacyServer.RegisterWallet(wallet)
	}
}
