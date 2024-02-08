# HTTP API Reference

[[toc]]

## String Format Rules

- `0xFFFFFF...i0` , all fields represent the **inscription ID** will start with 32 bytes of hex string and then concat an `i` and a number, which the 32 bytes of hex string is the transaction hash and the number is the index of the output. In a simple word, it is the outpoint of the inscription.
- `0xFFFFFF...:0` , all fields represent the **satpoint** or **outpoint** are the same thing. For compatibility reasons, the format of satpoint remains unchanged from the ord node and appears similar to the inscription ID, with  `i` replaced by `:` .
- `chain` , this field represents the **destination chain** or **circulating chain** of the inscription, which is the layer two network where the inscription circulates. It follows the [SLIP-0044](https://github.com/satoshilabs/slips/blob/master/slip-0044.md) standard and uses the value from the `Coin type` column
- `amount` , this field represents the **amount** of kinds of tokens. Storing big integers with string is more convenient in some languages.


## Inscriptions

These APIs are used to query inscription objects.

### GET /inscription/:query

Query the attributes of a specific inscription object.

#### Parameters

- `query: string`, **required**, it can be an inscription ID, `inscription_id`, or the auto-incremented serial number in the database, `inscription_number`.

#### Response

```json
{
  "inscription_id": "0xFFF...i0",
  "inscription_number": 123,
  "next": "0xFFF...i0",
  "previous": "0xFFF...i0",
  "address": "bc1p...",
  "content_length": 12345,
  "content_type": "application/json",
  "content_protocol": "C-BRC-20",
  "genesis_fee": 123456789,
  "genesis_height": 123456,
  "output_value": 123456789,
  "satpoint": "0xFFF...:0",
  "timestamp": 1234567890123,
  "c_ins_description": {
    "type": "blockchain",
    "chain": 309,
    "contract": "0xFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF",
  },
  "parent": null,
  "children": null,
  "rune": null,
  "sat": null,
}
```

- `inscription_id: string`, the ID of the inscription object.
- `inscription_number: number`, the auto-incremented serial number for the inscription object in the database.
- `next: string`, **optional**, the ID of the next inscription, which means the inscription at  `inscription_number + 1` .
- `previous: string`, **optional** , the ID of the previous inscription, which means the inscription at `inscription_number - 1` .
- `address: string`, **optional** , the address holding the inscription.
- `content_length: number`, **optional** , the length of the inscription content.
- `content_type: string`, **optional** , the MIME type of the inscription content.
- `content_protocol: string`, the application protocol which the inscription content follows.
- `genesis_fee: number`, the transaction fee paid when the inscription was created.
- `genesis_height: number`, the height of the block where the inscription was created.
- `output_value: number`, **optional** , the value of the UTXO containing the inscription.
- `satpoint: string`, **optional** , the outpoint of the UTXO containing the inscription.
- `timestamp: number`, the timestamp of the block where the inscription belongs.
- `c_ins_description: object`, the description field for C-INS inscription, it contains multiple fields:
    - When the `type` is set to `ordinals`, no other fields need to be provided. The inscription can be handled as an ordinal inscription.
    - When the `type` is set to `blockchain`, the following fields are required:
        - A `chain` field indicates the circulating chain of the inscription, which should follow the Coin Type column in [SLIP-0044](https://github.com/inscription-c/insc/wiki/Inscription/4491faf434b9af97ad81be0c2557fe4b09a55581) proposal.
        - A `contract` field indicates the contract of the inscription on the circulating chain, which should be able to interpret the circulating chain, whether it is an address or a hash.
- `parent` , **optional** , the ID of the parent inscription, not used in the **c-ins** protocol for now.
- `children` , **optional** , a list of IDs of the children inscription, not used in the **c-ins** protocol for now.
- `rune` , **optional** , rune protocol field, not used in the **c-ins** protocol for now.
- `sat` , **optional** , the ordinal of the UTXO containing the inscription, not used in the **c-ins** protocol for now.


### GET /content/:inscription_id

Query the content of a specific inscription object.

#### Alias

- `GET /preview/:inscription_id`

#### Parameters

- `inscription_id: string` , **required**, the ID of the inscription object.

#### Response

The response is based on the inscription's `content_encoding` and `content_type`. For example, if the inscription object is a JSON object, the HTTP response will be a JSON object. Also, it can be an image or a video file.


### GET /inscriptions/:page

Query batch of inscription objects, default 100 per page.

#### Parameters

- `page: number` , **optional**, the current page number, default the first page.

#### Response

```json
{
    "page_index": 1,
    "more": true,
    "inscriptions": [
        "0xFFF...i0",
        "0xFFF...i0",
        ...
    ]
}
```

- `page_index: number`, the current page number.
- `more: bool`, indicates if the next page exists.
- `inscriptions: string[]`, the list of inscription IDs on the current page.

### GET /inscriptions/block/:block/:page

Query batch of inscriptions in a specific block, default 100 per page.

#### Parameters

- `block: number` , **required**, the block height, default the latest block.
- `page: number` , **optional**,  the current page number, default the first page.

#### Response

```json
{
    "block_height": 10000,
    "page_index": 1,
    "more": true,
    "inscriptions": [
        "0xFFF...i0",
        "0xFFF...i0",
        ...
    ]
}
```

- `block_height: number` , the block height of the inscriptions.
- `page_index: number` , the current page number.
- `more: bool` , indicates if the next page exists.
- `inscriptions: string[]` , the list of inscription IDs on the current page.

## C-BRC-20

These APIs are used to query C-BRC-20 token objects.

### GET /cbrc20/token/:ticker_id

Query the attributes of a specific C-BRC-20 token object.

#### Parameters

- `ticker_id: string` , **required**, the ID of the C-BRC-20 token.

#### Response

```json
{
    "ticker_id": "0xFFFFFF...i0",
    "ticker": "c-ins",
    "chain": 309,
    "contract": "0xFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF",
    "total_supply": "21000000"
}
```

- `ticker_id: string` , the ID of the token object, it is come from the inscription ID which `deploy` the token.
- `ticker: string` , the name of the token.
- `chain: number` , the target layer two network which the token will circulate. The specific number comes from Coin Type column in [SLIP-0044](https://github.com/inscription-c/insc/wiki/Inscription/4491faf434b9af97ad81be0c2557fe4b09a55581) proposal.
- `contract: string` , the contract of the token on the circulating chain.
- `total_supply: string` , the total supply of the token.

### GET /cbrc20/tokens/:ticker/:page

Query batch of C-BRC-20 token objects, default 100 per page.

#### Parameters

- `ticker: string` , **optional**, the name of the token.
- `page: number` , **optional**, the current page number, default the first page.

#### Response

```json
{
    "page_index": 1,
    "more": true,
    "tokens": [
        {
            "ticker_id": "0xFFFFFF...i0",
            "ticker": "c-ins",
            "chain": 309,
            "contract": "0xFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF",
            "total_supply": "21000000",
        },
        // ... more tokens
    ]
}
```

- `page_index: number` , the current page number.
- `more: bool` , indicates if the next page exists.
- `tokens: token[]` , the list of tokens on the current page.


## Other

### GET /blockhash/:height

Query the block hash of specific height, default the latest block.

#### Parameters

- `height: number`, **optional**, the block height.

#### Response

For compatibility reasons, the response is in plain text format.

```text
0xFFFFF...
```

### GET /blockheight

Query the latest block height.

#### Response

For compatibility reasons, the response is in plain text format.

```text
10000
```

### GET /clock

Query the datetime calculated from the timestamp of the latest block.

#### Response

```json
{
  "height": 10000,
  "hour": 23,
  "minute": 0,
  "second": 0,
}
```

### GET /search/:query

Redirect to the specific object query API according to a regex rule, similar to the  search function in most blockchain explorers.
