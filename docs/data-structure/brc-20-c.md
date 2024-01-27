# BRC-20-C

[[toc]]

Nowadays, people are crazy about minting inscriptions on Bitcoin. However, most inscriptions they mint are BRC-20 tokens, which use an application protocol based on the original inscription protocol. Minting BRC-20 tokens on Bitcoin is so appealing because of its fairness and narratives. All participants must pay a fee to mint and follow the transaction ordering rules on the chain.

Unlike BRC-20, BRC-20-C currently only needs two operations: deploy and mint. The "deploy" operation means creating a new BRC-20-C token. The "mint" operation means minting a BRC-20-C token. After the inscription of the "mint" operation is revealed on Bitcoin, it will be locked and can only be transferred on the circulating chain.

## Operations

### Deploy

First, A "deploy" operation is required to deploy a new BRC-20-C token. After deploying the token on Bitcoin, it can be minted on Bitcoin.

```json
{
  "p": "brc-20-c",
  "op": "deploy",
  "tick": "ins-c",
  "max": "21000000",
  "lim": "1000",
  "dec": "...."
}
```

|  Key | Required? | Full Name  | Description                                                                                                 |
|-----:|:----------|:-----------|:------------------------------------------------------------------------------------------------------------|
|    p | Yes       | Protocol   | Identifier of the protocol.                                                                                 |
|   op | Yes       | Operation  | Types of operation.                                                                                         |
| tick | Yes       | Ticker     | The token name has no length limit and can be duplicated. Characters should be limited to `0-9a-z` and `-`. |
|  max | Yes       | Max supply | Total supply of the token.                                                                                  |
|  lim | No        | Mint limit | Output limit for each minting.                                                                              |
|  dec | No        | Decimals   | Decimals, the default value is 18.                                                                              |

### Mint

After deploying the token, it can be minted on Bitcoin with a `mint` operation.

```json
{
  "p": "brc-20-c",
  "op": "mint",
  "tkid": "0xFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF...",
  "tick": "ins-c",
  "amt": "1000",
  "to": "bc1p.../ckb1..."
}
```

|  Key | Required? | Full Name  | Description                                                                        |
|-----:|:----------|:-----------|:-----------------------------------------------------------------------------------|
|    p | Yes       | Protocol   | Same as above.                                                                     |
|   op | Yes       | Operation  | Same as above.                                                                     |
| tkid | Yes       | Ticker ID  | The unique identifier between different tickers.                                   |
| tick | Yes       | Ticker     | Same as above.                                                                     |
|  amt | Yes       | Amount     | The amount in this `mint` operation must equal or less than the `lim` .         |
|   to | No        | To Address | The recipient address can be an arbitrary string and is parsed by layer two nodes. |
