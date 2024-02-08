# Installation

[[toc]]


## Prerequisites

To run an C-INS node, a Bitcoin full node is required. Please download the latest version of [Bitcoin Core](https://bitcoincore.org/en/download/). Please DO NOT use the **btcd**, because it lacks certain JSON RPC APIs necessary for the node. Additionally, it does not have wallet functionality for transaction signing.

## Pre-built Releases

The pre-built binary is only available for Linux for now. It can be downloaded from [releases](https://github.com/inscription-c/insc/releases). Afterward, unzip the binary to a directory in your `$PATH` .

## Build from source

The node is written in Go, so it is easy to build from the source:

```bash
git clone https://github.com/inscription-c/insc.git
cd insc && go mod download
go build
```

Set go proxy if a connection to GitHub is not possible:

```bash
go env -w GOPROXY=https://goproxy.io,direct
```

