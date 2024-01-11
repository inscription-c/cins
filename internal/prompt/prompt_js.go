package prompt

import (
	"bufio"
	"fmt"

	"github.com/btcsuite/btcwallet/internal/legacy/keystore"
)

func ProvideSeed() ([]byte, error) {
	return nil, fmt.Errorf("prompt not supported in WebAssembly")
}

func ProvidePrivPassphrase() ([]byte, error) {
	return nil, fmt.Errorf("prompt not supported in WebAssembly")
}

func PrivatePass(_ *bufio.Reader, _ *keystore.Store) ([]byte, error) {
	return nil, fmt.Errorf("prompt not supported in WebAssembly")
}

func PublicPass(_ *bufio.Reader, _, _, _ []byte) ([]byte, error) {
	return nil, fmt.Errorf("prompt not supported in WebAssembly")
}

func Seed(_ *bufio.Reader) ([]byte, error) {
	return nil, fmt.Errorf("prompt not supported in WebAssembly")
}
