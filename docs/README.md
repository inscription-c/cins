# Introduction

[[toc]]


## The Real Inscription of All Blockchain

Each sat on Bitcoin is like a gemstone made up of binary, protected by its locking script individually. It is nearly impossible for anyone to steal multiple sats at once. Therefore, these inscribed sats genuinely deserve the name "inscription". However, what sets Bitcoin apart from other chains?

Since the Bitcoin taproot upgrade, improvements have primarily focused on signature and privacy. Even the tapscript, which enables inscription, serves these two features. If looking closely at the upcoming BIPs in proposed and draft status, it becomes evident that Bitcoin's core principles stay focused on signature and privacy. Therefore, Bitcoin always focuses on issuing and storing assets in the past, present, or future.


## The Real Needs of The Market

Recently, the Bitcoin ETFs finally received approval, a milestone event. In addition to that, we have also witnessed the rise of inscription dex on Bitcoin, which is an exciting phenomenon. These events indicate the market's demand for asset liquidity, not just storage.

However, the current inscription assets on Bitcoin are facing the following problems:

- It led to UTXO storage expansion on Bitcoin; the Block size has obviously increased since this year.
- It incurs high transaction fees and may cause congestion across the entire blockchain.
- It has a longer transaction confirmation time, which could be more user-friendly for retail investors.
- It lacks composability of inscription assets due to the absence of "Turing completeness" smart contract.

Perhaps we can anticipate that Bitcoin will solve these problems in the distant future, but why should we wait when we already have available solutions? Nowadays, layer two blockchains have become increasingly mature in solving these problems, and many of them are naturally programmable.


## The Solution Inscription-Contractualized Protocol

As mentioned above, we aim to enhance liquidity for Bitcoin assets and leverage layer two capabilities to program the assets. Therefore, we propose the **Inscription-Contractualized Protocol**, abbreviated as **INS-C**.

In INS-C, all assets are real inscriptions that can only be minted on Bitcoin. However, once minted, all assets are locked naturally and can only be unlocked on the layer two blockchain. In this way, INS-C inscriptions are not only a successor of ordinal inscriptions but also a bridge between ordinal inscriptions and layer two blockchain.

- At its core, it inherits the simplicity, purity, and fairness of ordinal inscriptions.
- In practice, it remains open and maintains forward compatibility with ordinal inscriptions by allowing existing ordinal assets to join by "promotion".


## Get Started

For further reading about INS-C, please read the following documents:

- [Inscription](./data-structure/inscription.md), which introduced the inscription data structure.
- [BRC-20-C](./application-protocol/brc-20-c.md), which introduced the first application protocol on INS-C.

For developers, please refer to the following documents:

- [Installation](./node-guide/installation.md), which introduced how to install the INS-C node.
- [Run INS-C Node](./node-guide/run-node.md), which introduced how to run the INS-C node.
- [Deploy Inscription](./node-guide/deploy-inscription.md), which introduced how to deploy an inscription on Bitcoin.
- [HTTP API Reference](./node-guide/http-api.md), which introduced the HTTP API of INS-C node.
