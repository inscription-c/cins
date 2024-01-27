# Inscription

[[toc]]

**Inscription-contractualized protocol**, abbreviated as **ins-c**, which is the foundation of all other application protocols:

- First, a tapscript is created to store the inscription content.
- Next, a commit transaction is created to push the tapscript to the blockchain.
- After the commit transaction is pushed, a reveal transaction is created to spend the sats containing the inscription content.

The difference is that once a sat has been inscribed in ins-c, it will be locked automatically and can be only unlocked on the layer two network where it is declared. The inscription can then be transferred on the layer two network and follow the smart contract constraints on the layer two network. In the subsequent documents, we will refer to this layer two network as the **circulating chain**.

## Tapscript Structure

The inscription content of ins-c remains fully on-chain as it is essentially a tapscript stored in the script path of a taproot address, which starts with `bc1p`. The original ordinal inscription protocol is quite extensible, and every existing inscription should be able to "upgrade" to an ins-c inscription. Therefore, we are adding more features while maintaining backward compatibility.

A text-type inscription is a tapscript like the following:

```sql
# Some more opcodes
OP_FALSE
OP_IF
  OP_PUSH "ins-c"
  OP_PUSH 10000
  OP_PUSH { "chain_type": 309, "unlocker": "0xf1ef61b6977508d9ec56fe43399a01e576086a76cf0f7c687d1418335e8c401f" }
  OP_PUSH 1
  OP_PUSH "text/plain;charset=utf-8"
  OP_PUSH 0
  OP_PUSH "Hello, world!"
OP_ENDIF
```

- The `OP_FALSE OP_IF … OP_ENDIF` is a particular construct that allows the interpreter to ignore all the `OP_PUSH` surrounded by it.
- `OP_PUSH "ins-c"`: The identifier for ins-c protocol.
- `OP_PUSH <number>`: The field tag; refer to the next section for the full definition of all tags.
- `OP_PUSH "text/plain;charset=utf-8"`: The field value.

## Field tags

Most tags are compatible with the [ordinal inscription protocol](https://docs.ordinals.com/inscriptions.html#fields). However, due to the layer two nature of ins-c inscription, some tags will be ignored in an on-chain environment. The following list shows all the tags and their meanings:

- `10000` means `unlock_condition`, a necessary tag indicating the inscription's circulating chain. The JSON may include additional fields, but it must also include these two fields:
  - A `chain_type` field, which should follow the `Coin Type` column in [SLIP-0044]() proposal.
  - A `unlocker` field, which indicates the unlocker address, public key, or script hash, should be able to interpret the circulating chain.
- `1` means `content-type`, which refers to the MIME type of the inscription content.
- `2` will be ignored for now.
- `3` will be ignored for now.
- `5` means `metadata`; it may optionally contain metadata in JSON format.
- `7` will be ignored for now.
- `9` means `content_encoding`, which refers to how the inscription content encoding.
- `11` will be ignored for now.
- `0` means `content`, which refers to the main content of the inscription; if the content exceeds the 520 bytes limit of tapscript, it should be split into multiple `OP_PUSH`.

## Application protocol

In the inscription content, it is possible to store arbitrary data. However, using a common protocol can enhance its functionality and interoperability. Therefore, we have predefined some **application protocols** for ins-c :

- [BRC-20-C](./brc-20-c.md): A BRC-20 compatible token protocol.
