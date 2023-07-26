# WebSockets POC

## Running

### Prerequisites

* Get [Rancher Desktop](https://docs.rancherdesktop.io/getting-started/installation/) or some other similar local k8s cluster.
* Disable Traefik in _Preferences->Kubernetes_ as we use Istio for ingress
* Get Skaffold:

```shell
brew install skaffold
skaffold config set default-repo local
```

### Build and Run

```shell
skaffold run
```

Run it! (You will get disconnected after 60s if you do not send ping frames with `/ping`)

```shell
wscat -c 'ws://localhost/ws' --slash -P
Connected (press CTRL+C to quit)
> {"jsonrpc" : "2.0", "id": 1, "method": "subscribe", "params": ["market:spot:tickers", { "pair": "btc_usdt", "exchange": "binance" }]}
< {"jsonrpc": "2.0", "id": 1, "result": "123456"}
< {"jsonrpc":"2.0","id":1,"method":"subscription","params":{"subscription":"1223456","result":{"exchange":"binance","timestamp":1690297936009,"bid":0.01,"ask":0.11,"bidVolume":100,"askVolume":1000}}}
```