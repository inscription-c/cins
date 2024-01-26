package server

import (
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/inscription-c/insc/inscription/index"
)

type Options struct {
	rescan bool
	idx    *index.Indexer
	cli    *rpcclient.Client
}

type Option func(*Options)

type Runner struct {
	exist chan struct{}
	opts  *Options
}

func WithClient(cli *rpcclient.Client) func(*Options) {
	return func(options *Options) {
		options.cli = cli
	}
}

func WithIndex(idx *index.Indexer) func(*Options) {
	return func(options *Options) {
		options.idx = idx
	}
}

func WithRescan(rescan bool) func(*Options) {
	return func(options *Options) {
		options.rescan = rescan
	}
}

func NewRunner(opts ...Option) *Runner {
	r := &Runner{
		exist: make(chan struct{}),
		opts:  &Options{},
	}
	for _, v := range opts {
		v(r.opts)
	}
	return r
}
