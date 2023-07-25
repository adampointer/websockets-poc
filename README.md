# WebSockets POC

## Running

### Prerequisites

Get Rancher Desktop or some other similar local k8s cluster

Get Skaffold

```shell
brew install skaffold
skaffold config set default-repo local
```

Get your local node IP

```shell
kubectl get nodes -o wide

NAME                   STATUS   ROLES                  AGE    VERSION        INTERNAL-IP    EXTERNAL-IP    OS-IMAGE             KERNEL-VERSION   CONTAINER-RUNTIME
lima-rancher-desktop   Ready    control-plane,master   103d   v1.25.6+k3s1   192.168.5.15   192.168.0.19   Alpine Linux v3.16   5.15.96-0-virt   docker://20.10.20
```

Run it

```shell
skaffold run

wscat -c 'ws://192.168.0.19:30000/ws'
Connected (press CTRL+C to quit)
> {"jsonrpc" : "2.0", "id"      : 1, "method"  : "subscribe", "params"  : [ "market:spot:tickers", { "pair": "btc_usdt", "exchange": "binance" } ]}
< {"jsonrpc": "2.0", "id": 1, "result": "123456"}
< {"jsonrpc":"2.0","id":1,"method":"subscription","params":{"subscription":"1223456","result":{"exchange":"binance","timestamp":1690297936009,"bid":0.01,"ask":0.11,"bidVolume":100,"askVolume":1000}}}
```