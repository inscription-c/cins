# C‐BRC‐20

[[toc]]

Nowadays, people are crazy about minting inscriptions on Bitcoin. However, most inscriptions they mint are BRC-20 tokens, which use an application protocol based on the original inscription protocol. Minting BRC-20 tokens on Bitcoin is so appealing because of its fairness and narratives. All participants must pay a fee to mint and follow the transaction ordering rules on the chain.

Unlike BRC-20, C-BRC-20 currently only needs two operations: deploy and mint. The "deploy" operation means creating a new C-BRC-20 token. The "mint" operation means minting a C-BRC-20 token. After the inscription of the "mint" operation is revealed on Bitcoin, it will be locked and can only be transferred on the circulating chain.

## Operations

### Deploy

First, A "deploy" operation is required to deploy a new C-BRC-20 token. After deploying the token on Bitcoin, it can be minted on Bitcoin.

```json
{
  "p": "c-brc-20",
  "op": "deploy",
  "tick": "c-ins",
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
|  dec | No        | Decimals   | Decimals, the default value is 18.                                                                          |

### Mint & Transfer

These actions should be implemented on a circulating chain using a smart contract. There is no limit to the detail in the implementation, but C-INS does have some recommended practices:

- The contract should be capable of storing the Inscription ID on Bitcoin and providing an API for retrieval.
- The minting process should be as fair as possible.
