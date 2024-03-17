# cins
cins is an index, block explorer, and command-line wallet. It is experimental software with no warranty. 
See LICENSE for more details.

While the Ordinals Inscription is an excellent evolution, it is not flawless and does have some issues.

- It led to UTXO storage expansion on Bitcoin; the Block size has obviously increased since last year.
- It incurs high transaction fees and may cause congestion across the entire blockchain.
- It has a longer transaction confirmation time, which is not user-friendly for small retail investors.
- It lacks composability due to the absence of "Turing completeness" smart contract.

See the [docs](https://docs.c-ins.com/) for documentation and guides.

# Installation

go version >= 1.21 is required.
```bash
go install github.com/inscription-c/cins@latest
```

# Wallet

cins wallet is modifications made based on btcd, So you can manage your wallet through [btcctl](https://github.com/btcsuite/btcd/tree/master/cmd/btcctl).

Install btcctl with:
```bash
go install  github.com/btcsuite/btcd/cmd/btcctl@v0.24.0
```

Start wallet service with:
```bash
cins wallet --wallet_pass root --chain_url <bitcoin_rpc_connect>
```

## example

Get the default wallet balance:
```bash
btcctl getbalance default --wallet --notls --rpcuser root --rpcpass root
```

Inscribe inscriptions:
```bash
cins inscribe -f <inscription_file_path> --c_ins_description <c_ins_description_file_path> --dest <dest_owner_address> --indexer_url <cins_indexer_url>
```

inscribe flags:
```bash
Usage:
  cins inscribe [flags]

Flags:
      --c_brc_20                   is c-brc-20 protocol, add this flag will auto check protocol content effectiveness
      --c_ins_description string   cins protocol description.
      --cbor_metadata string       Include CBOR in file at <METADATA> as inscription metadata
      --compress                   Compress inscription content with brotli.
      --dest string                Send inscription to <DESTINATION> address.
      --dry_run                    Don't sign or broadcast transactions.
  -f, --filepath string            inscription file path
  -h, --help                       help for inscribe
      --indexer_url string         the URL of indexer server (default http://localhost:8335, testnet: http://localhost:18335) (default "http://localhost:8335")
      --json_metadata string       Include JSON in file at <METADATA> converted to CBOR as inscription metadata  
      --no_backup                  Do not back up recovery key.
  -p, --postage uint               Amount of postage to include in the inscription. (default 10000)
  -t, --testnet                    bitcoin testnet3
      --wallet_pass string         wallet password for master private key (default "root")
      --wallet_rpc_pass string     wallet rpc server password (default "root")
      --wallet_rpc_user string     wallet rpc server user (default "root")
      --wallet_url string          the URL of wallet RPC server to connect to (default http://localhost:8332, testnet: localhost:18332) (default "localhost:8332")
```

# Indexer

```bash
cins indexer -u root -P root --mysql_addr <mysql_addr> --mysql_user <mysql_user> --mysql_pass <mysql_pass> --mysql_db <mysql_db> --chain_url <bitcoin_rpc_connect>
```

or run with config file

```bash
cins indexer -c <path_to_config_file> 
```

config example
```yaml
server:
  testnet: true
  rpc_listen: ":18335"
  no_api: false
  index_sats: true
  index_spend_sats: false
chain:
  url: "http://127.0.0.1:18334"
  username: "root"
  password: "root"
db:
  mysql:
    addr: "127.0.0.1:3306"
    user: "root"
    password: "root"
    db: "cins"
sentry:
  dsn: ""
  traces_sample_rate: 1.0
origins:
  - ".*"
```