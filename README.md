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