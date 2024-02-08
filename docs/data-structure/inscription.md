# Inscription

[[toc]]

**Contractualized-inscription protocol**, abbreviated as **c-ins**, which is the foundation of all other application protocols:

- First, a tapscript is created to store the inscription content.
- Next, a commit transaction is created to push the tapscript to the blockchain.
- After the commit transaction is pushed, a reveal transaction is created to spend the sats containing the inscription content.

The difference is that once a sat has been inscribed in c-ins, it will be locked automatically and can be only unlocked on the layer two network where it is declared. The inscription can then be transferred on the layer two network and follow the smart contract constraints on the layer two network. In the subsequent documents, we will refer to this layer two network as the **circulating chain**.

## Tapscript Structure

The inscription content of c-ins remains fully on-chain as it is essentially a tapscript stored in the script path of a taproot address, which starts with `bc1p`. The original ordinal inscription protocol is quite extensible, and every existing inscription should be able to "upgrade" to an c-ins inscription. Therefore, we are adding more features while maintaining backward compatibility.

A text-type inscription is a tapscript like the following:

```sql
# Some more opcodes
OP_FALSE
OP_IF
  OP_PUSH "c-ins"
  OP_PUSH -1
  OP_PUSH { "type": "blockchain", "chain": 309, "contract": "0xFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF" }
  OP_PUSH 1
  OP_PUSH "text/plain;charset=utf-8"
  OP_PUSH 0
  OP_PUSH "Hello, world!"
OP_ENDIF
```

- The `OP_FALSE OP_IF … OP_ENDIF` is a particular construct that allows the interpreter to ignore all the `OP_PUSH` surrounded by it.
- `OP_PUSH "c-ins"`: The identifier for c-ins protocol.
- `OP_PUSH <number>`: The field tag; refer to the next section for the full definition of all tags.
- `OP_PUSH "text/plain;charset=utf-8"`: The field value.

## Field tags

Most tags are compatible with the [ordinal inscription protocol](https://docs.ordinals.com/inscriptions.html#fields). However, due to the layer two nature of c-ins inscription, some tags will be ignored in an on-chain environment. The following list shows all the tags and their meanings:

- `-1` means `c_ins_description`, which is stored as `0x2D31` in tapscript. This tag is essential for indicating the circulating information of the inscription. Its value is in JSON format, and further details will be explained later.
- `1` means `content-type`, which refers to the MIME type of the inscription content.
- `2` will be ignored for now.
- `3` will be ignored for now.
- `5` means `metadata`; it may optionally contain metadata in JSON format.
- `7` will be ignored for now.
- `9` means `content_encoding`, which refers to how the inscription content encoding.
- `11` will be ignored for now.
- `0` means `content`, which refers to the main content of the inscription; if the content exceeds the 520 bytes limit of tapscript, it should be split into multiple `OP_PUSH`.

### C-INS Description

The value of this field must include a required field called `type`. The remaining fields should be provided based on the different types:

- When the `type` is set to `ordinals`, no other fields need to be provided. The inscription can be handled as an ordinal inscription.
- When the `type` is set to `blockchain`, the following fields are required:
  - A `chain` field indicates the circulating chain of the inscription, which should follow the Coin Type column in [SLIP-0044](https://github.com/inscription-c/insc/wiki/Inscription/4491faf434b9af97ad81be0c2557fe4b09a55581) proposal.
  - A `contract` field indicates the contract of the inscription on the circulating chain, which should be able to interpret the circulating chain, whether it is an address or a hash.

In addition to the fields mentioned above, each project has the flexibility to include additional fields for its specific purpose. The only restriction is that the size of the script element cannot exceed 520 bytes.


## Application protocol

In the inscription content, it is possible to store arbitrary data. However, using a common protocol can enhance its functionality and interoperability. Therefore, we have predefined some **application protocols** for c-ins :

- [C-BRC-20](https://github.com/inscription-c/insc/wiki/C%E2%80%90BRC%E2%80%9020): A BRC-20 compatible token protocol.

