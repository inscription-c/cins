package server

import (
	"errors"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/inscription-c/insc/inscription/index"
	"github.com/inscription-c/insc/inscription/log"
	"time"
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

func (r *Runner) Start() {
	go func() {
		ticker := time.NewTicker(time.Second * 5)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				err := r.opts.idx.UpdateIndex()
				if err != nil && !errors.Is(err, index.ErrInterrupted) {
					log.Srv.Error("UpdateIndex", "err", err)
				}
				if errors.Is(err, index.ErrInterrupted) {
					return
				}
			case <-r.exist:
				return
			}
		}
	}()
}

func (r *Runner) Stop() {
	close(r.exist)
}
