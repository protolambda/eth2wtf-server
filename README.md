# ETH2.WTF server

Serves the ETH2.WTF client; providing network data over websockets.

Work in progress project.


## websocket message spec

### `Client -> Server`:

```
<1 byte: msg type><remaining: msg>

0x00: Default, not used.

TODO
```

### `Server -> Client`

```
<1 byte: msg type><remaining: msg>

TODO

```

