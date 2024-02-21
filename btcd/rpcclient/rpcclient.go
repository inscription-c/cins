package rpcclient

import (
	"github.com/go-playground/validator/v10"
)

// clientOptions is a struct that holds the configuration options for the client.
// It includes fields for the Host, User, Password, and Batch.
type clientOptions struct {
	Host     string `validator:"uri"` // Host is the URI of the host.
	User     string // User is the username for authentication.
	Password string // Password is the password for authentication.
	Batch    bool   // Batch indicates whether the client should operate in batch mode.
}

// ClientOption is a function type that takes a pointer to a clientOptions struct as a parameter.
// It is used to set the fields of the clientOptions struct.
type ClientOption func(*clientOptions)

// WithClientHost is a function that returns a ClientOption.
// The returned ClientOption sets the Host field of the clientOptions struct to the provided host.
func WithClientHost(host string) ClientOption {
	return func(o *clientOptions) {
		o.Host = host
	}
}

// WithClientUser is a function that returns a ClientOption.
// The returned ClientOption sets the User field of the clientOptions struct to the provided user.
func WithClientUser(user string) ClientOption {
	return func(o *clientOptions) {
		o.User = user
	}
}

// WithClientPassword is a function that returns a ClientOption.
// The returned ClientOption sets the Password field of the clientOptions struct to the provided password.
func WithClientPassword(password string) ClientOption {
	return func(o *clientOptions) {
		o.Password = password
	}
}

// WithClientBatch is a function that returns a ClientOption.
// The returned ClientOption sets the Batch field of the clientOptions struct to the provided batch.
func WithClientBatch(batch bool) ClientOption {
	return func(o *clientOptions) {
		o.Batch = batch
	}
}

// NewClient is a function that creates a new rpcclient.Client.
// It takes a variadic number of ClientOption functions as parameters, which are used to set the fields of the clientOptions struct.
// The function first applies each ClientOption to the clientOptions struct.
// It then validates the clientOptions struct using the validator package.
// If there is an error during validation, it returns nil and the error.
// The function then creates a new rpcclient.ConnConfig with the fields of the clientOptions struct.
// If the Batch field of the clientOptions struct is true, it creates a new batch client.
// Otherwise, it creates a new client.
// The function then returns the client and nil.
func NewClient(optFns ...ClientOption) (*Client, error) {
	opts := &clientOptions{}
	for _, v := range optFns {
		v(opts)
	}
	if err := validator.New().Struct(opts); err != nil {
		return nil, err
	}
	connCfg := &ConnConfig{
		Host:         opts.Host,
		User:         opts.User,
		Pass:         opts.Password,
		HTTPPostMode: true,
	}
	if opts.Batch {
		return NewBatch(connCfg)
	} else {
		return New(connCfg, nil)
	}
}
