# HTTP API Reference

[[toc]]

## String Format Rules

- `0xFFFFFF...i0` , all fields represent the **inscription ID** will start with 32 bytes of hex string and then concat an `i` and a number, which the 32 bytes of hex string is the transaction hash and the number is the index of the output. In a simple word, it is the outpoint of the inscription.
- `0xFFFFFF...:0` , all fields represent the **satpoint** or **outpoint** are the same thing. For compatibility reasons, the format of satpoint remains unchanged from the ord node and appears similar to the inscription ID, with  `i` replaced by `:` .
- `dst_chain` , this field represents the **destination chain** or **circulating chain** of the inscription, which is the layer two network where the inscription circulates. It follows the [SLIP-0044](https://github.com/satoshilabs/slips/blob/master/slip-0044.md) standard and uses the value from the `Coin type` column
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
  "genesis_fee": 123456789,
  "genesis_height": 123456,
  "output_value": 123456789,
  "satpoint": "0xFFF...:0",
  "timestamp": 1234567890123,
  "dst_chain": "ckb",
  "content_protocol": "BRC-20-C",
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
- `genesis_fee: number`, the transaction fee paid when the inscription was created.
- `genesis_height: number`, the height of the block where the inscription was created.
- `output_value: number`, **optional** , the value of the UTXO containing the inscription.
- `satpoint: string`, **optional** , the outpoint of the UTXO containing the inscription.
- `timestamp: number`, the timestamp of the block where the inscription belongs.
- `dst_chain: string`, the target layer two network which the inscription will circulate on.
- `content_protocol: string`, the application protocol which the inscription content follows.
- `parent` , **optional** , the ID of the parent inscription, not used in the **ins-c** protocol for now.
- `children` , **optional** , a list of IDs of the children inscription, not used in the **ins-c** protocol for now.
- `rune` , **optional** , rune protocol field, not used in the **ins-c** protocol for now.
- `sat` , **optional** , the ordinal of the UTXO containing the inscription, not used in the **ins-c** protocol for now.


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

## BRC-20-C

These APIs are used to query BRC-20-C token objects.

### GET /brc20c/token/:ticker_id

Query the attributes of a specific BRC-20-C token object.

#### Parameters

- `ticker_id: string` , **required**, the ID of the BRC-20-C token.

#### Response

```json
{
    "ticker_id": "0xFFFFFF...i0",
    "ticker": "ins-c",
    "chain": "ckb",
    "total_supply": "21000000",
    "circulating_supply": "10000000",
    "holders": 1000
}
```

- `ticker_id: string` , the ID of the token object, it is come from the inscription ID which `deploy` the token.
- `ticker: string` , the name of the token.
- `dst_chain: string` , the target layer two network which the token will circulate on.
- `total_supply: string` , the total supply of the token.
- `circulating_supply: string` , the circulating supply of the token.
- `holders: string` , the number of token holders.

### GET /brc20c/tokens/:ticker/:page

Query batch of BRC-20-C token objects, default 100 per page.

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
            "ticker": "ins-c",
            "chain": "ckb",
            "total_supply": "21000000",
            "circulating_supply": "10000000",
            "holders": 1000
        },
        // ... more tokens
    ]
}
```

- `page_index: number` , the current page number.
- `more: bool` , indicates if the next page exists.
- `tokens: token[]` , the list of tokens on the current page.

### GET /brc20c/mint-history/:ticker_id/:page

Query the mint history of a specific BRC-20-C token.

#### Parameters

- `ticker_id: string` , **required**, the ID of the BRC-20-C token.
- `page: number` , **optional**, the current page number, default the first page.

#### Response

```json
{
    "chain": "ckb",
    "ticker_id": "0xFFFFFF...i0",
    "page_index": 1,
    "more": true,
    "mint_history": [
        {
            "inscription_id": "0xFFFFFF...i0",
            "miner": "bc1p...",
            "amount": "1000",
            "to_address": null
        },
        // ... more minting records
    ]
}
```

- `chain: string` , the target layer two network that the token circulates on.
- `ticker_id: string` , the ID of the token object, obtained from the inscription ID used to deploy the token.
- `page_index: number` , the current page number.
- `more: bool` , indicates if the next page exists.
- `mint_history: mint_record[]` , the list of minting records on the current page.

The structure of `mint_record` :
- `inscription_id: string` , the ID of the inscription.
- `miner: string` , the address of the miner who inscribed/minted the token.
- `amount: string` , the amount minted in this record.
- `to_address: string` , the recipient address of the minted token, it is arbitrary address of different layer two networks.

### GET /brc20c/mint-history/:address/:page

Query the mint history of a specific address.

#### Parameters

- `address: string` , **required**, the address of the token holder.
- `page: number` , **optional**, the current page number, default the first page.

#### Response

```json
{
    "page_index": 1,
    "more": true,
    "records": [
        {
            "inscription_id": "0xFFFFFF...i0",
            "miner": "bc1p...",
            "amount": "1000",
            "to_address": null
        },
        // ... more records
    ]
}
```

- `page_index: number` , the current page number.
- `more: bool` , indicates if the next page exists.
- `records: mint_record[]` , the list of minting records on the current page, which has the same structure as the `mint_record` in the [GET /brc20c/mint-history/:ticker_id/:page](#get-brc20c-mint-history-ticker-id-page) .

### GET /brc20c/holders/:ticker_id/:page

Query the holder list of a specific BRC-20-C token.

#### Parameters

- `ticker_id: string` , **required**, the ID of the BRC-20-C token.
- `page: number` , **optional**, the current page number, default the first page.

#### Response

```json
{
    "page_index": 1,
    "more": true,
    "holders": [
        {
            "address": "ckb1...",
            "amount": "1000",
        },
        // ... more holders
    ]
}
```

- `address: string` , the address of the token holder.
- `amount: string` , the total amount minted by the holder.


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
