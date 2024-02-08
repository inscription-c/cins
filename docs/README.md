# Introduction

[[toc]]


## The Great Evolution

The Ordinals Inscription is an excellent evolution for Bitcoin; it is just like a call of consensus. It make the entire Bitcoin community united again after years of chaos. In the past, creating anything on Bitcoin had to follow politically correct rules, but the Ordinals Inscription has opened people's minds and aligned them towards a common goal. Nowadays, teams are working on wallets, new assets, and, most importantly, layer two solutions for Bitcoin. However, all these efforts focus on enhancing user-friendliness and improving the Inscription.


## The Issues with the Ordinals Inscription

While the Ordinals Inscription is an excellent evolution, it is not flawless and does have some issues.

- It led to UTXO storage expansion on Bitcoin; the Block size has obviously increased since last year.
- It incurs high transaction fees and may cause congestion across the entire blockchain.
- It has a longer transaction confirmation time, which is not user-friendly for small retail investors.
- It lacks composability due to the absence of "Turing completeness" smart contract.

If we examine the Ordinals Inscription protocol closely, it is easy to comprehend the issues. The Ordinal Theory assigns a rarity value to each sat on Bitcoin. However, its main objective is to solve the tracking issue of the assets in Bitcoin's UTXO model. And these assets are wrapped in inscriptions, also known as tapscript on Bitcoin. The tapscript does not provide a "Turing completeness" smart contract environment; therefore, the inscription just utilizes some "text storage feature".


## The Essence of Inscription

A few months after the release of the Ordinals Inscription, a new inscription protocol called Atomicals emerged on Bitcoin. However, it is unfortunate that it did not address the existing issues of the Ordinals Inscription. From today's perspective, solving these issues through technical means seems impossible, especially on Bitcoin. Therefore, we focus on the essence of inscription and its true nature that unifies all consensus.

When people discuss inscription, they are referring to the assets on Bitcoin. They are not concerned with how the assets are tracked but where they are stored. When inscription stores an NFT on Bitcoin, it differs significantly from other NFTs on different chains. The inscription protocol is designed to store not just text but also images, audio, and videos. Therefore, NFTs with inscription can fully store the original data, which means the assets are entirely stored on Bitcoin. This is the essence of inscription and why people find it so exciting.


## The Contractualized-Inscription Protocol

As a team in the Bitcoin ecosystem, our goal is to improve Bitcoin from every possible angle. With this key objective in mind, we propose the **Contractualized-Inscription Protocol (C-INS)** as a solution for the issues mentioned above.

### Compatibility and Upgradability

As the successor to the Ordinals Inscription protocol, C-INS insists the most valuable essence which all assets are stored on Bitcoin, and it is fully compatible with the Ordinals Inscription; any existing Ordinals Inscription can be upgraded to C-INS.

### Powerful Composability

Once an C-INS asset is created on Bitcoin, it becomes naturally locked and can only be unlocked on a layer two blockchain. With recent advancements in layer two technology, supporting a smart contract environment with "Turing completeness" and higher transaction throughput has become common. As a result, C-INS assets not only solved the issues of the Ordinals Inscription but also benefited from the progress made by various layer two networks.

### About Fair Launch

Nowadays, the concept of a "fair launch" is widely accepted in the Bitcoin community because it aligns with the nature of the Ordinals Inscription. It means that anybody or any team must participate in the launch process fairly. Each inscription carries the same minting price; everyone begins minting simultaneously, and everyone has an equal opportunity to mint.

However, in C-INS, a "fair launch" appears more as an option rather than a requirement due to the power of smart contract. Therefore, we must clarify that the recommended approach for launching a C-INS asset is still through a "fair launch". It upholds what people truly value and remains the most effective way to ensure asset value.


## A New Paradigm for A New Era

During the early days of Bitcoin, its nature was unknown to most. However, people soon discovered that it is a huge global ledger. Then, the whole Bitcoin ecosystem limited by that understanding for so many years until today. Drawing inspiration from the Ordinals Inscription, C-INS has introduced a new paradigm for asset programming. It still leverages the battle-tested global ledger at its core but goes beyond that by being programmable and composable on various layer two networks—a concept we refer to as the **"mixed ledger"** .

Within this mixed ledger paradigm, all assets are inscriptions that can only be inscribed on Bitcoin. Each inscription represents an asset with complete data, eliminating the need for any off-chain services. At the same time, people can trade these assets on layer two networks. With the ultimate potential of layer two networks, C-INS will propel assets on Bitcoin to a new era.


## Get Started

For further reading about C-INS, please read the following documents:

- [Inscription](./data-structure/inscription.md), which introduced the inscription data structure.
- [C-BRC-20](./application-protocol/c-brc-20.md), which introduced the first application protocol on C-INS.

For developers, please refer to the following documents:

- [Installation](./node-guide/installation.md), which introduced how to install the C-INS node.
- [Run C-INS Node](./node-guide/run-node.md), which introduced how to run the C-INS node.
- [Deploy Inscription](./node-guide/deploy-inscription.md), which introduced how to deploy an inscription on Bitcoin.
- [HTTP API Reference](./node-guide/http-api.md), which introduced the HTTP API of C-INS node.
