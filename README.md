# ETH2.WTF server

Serves the ETH2.WTF client; providing network data over websockets.

Work in progress project.


## websocket message spec

### `Client -> Server`:

```
<1 byte: msg type><remaining: msg>

0x00: Default, not used.

0x01: Chunkified typed content, request: <uint32: chunk id><request body>

0x02: Update viewport. Server can push events based on client viewport
```

### `Server -> Client`

```
<1 byte: msg type><remaining: msg>

0x00: Default, not used.

0x01: Chunkified typed content, serving data: <uint32: chunk id><chunk data>

```

