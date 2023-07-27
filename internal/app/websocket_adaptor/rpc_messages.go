package websocket_adaptor

import (
	"encoding/json"
	"time"
	"websocket-poc/pkg/streamspb"

	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
)

const (
	eventTypeSpotTickers = "market:spot:tickers"
	eventTypeSpotTrades  = "market:spot:trades"
)

type ticker struct {
	Exchange  string  `json:"exchange"`
	Timestamp int64   `json:"timestamp"`
	Bid       float64 `json:"bid"`
	Ask       float64 `json:"ask"`
	BidVolume float64 `json:"bidVolume"`
	AskVolume float64 `json:"askVolume"`
}

type trade struct {
	Exchange  string  `json:"exchange"`
	Timestamp int64   `json:"timestamp"`
	Price     float64 `json:"price"`
	Quantity  float64 `json:"quantity"`
	IsBuy     bool    `json:"isBuy"`
	TradeID   string  `json:"tradeID"`
}

type subscriptionRPCRequest struct {
	Jsonrpc string        `json:"jsonrpc"`
	Id      int           `json:"id"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
}

type subscriptionRPCResponse struct {
	Jsonrpc string `json:"jsonrpc"`
	Id      int    `json:"id"`
	Method  string `json:"method"`
	Params  struct {
		Subscription string      `json:"subscription"`
		Result       interface{} `json:"result"`
	} `json:"params"`
}

func (r *subscriptionRPCRequest) getType() string {
	return r.Params[0].(string)
}

func (r *subscriptionRPCRequest) getArgs() map[string]interface{} {
	return r.Params[1].(map[string]interface{})
}

func mapEventTypeToFeature(evt string) streamspb.Feature {
	switch evt {
	case eventTypeSpotTickers:
		return streamspb.Feature_SPOT_TICKER
	case eventTypeSpotTrades:
		return streamspb.Feature_SPOT_TRADE
	default:
		return 99 // handle me
	}
}

func marshalRpcResponse(data *streamspb.Response) ([]byte, error) {
	var err error

	rpcResponse := &subscriptionRPCResponse{
		Jsonrpc: "2.0",
		Id:      1,
		Method:  "subscription",
		Params: struct {
			Subscription string      `json:"subscription"`
			Result       interface{} `json:"result"`
		}{
			Subscription: "1223456",
		},
	}

	now := time.Now().UTC().UnixMilli()
	var then time.Time

	if data.GetTicker() != nil {
		t := data.GetTicker()
		rpcResponse.Params.Result = &ticker{
			Exchange:  t.Exchange,
			Timestamp: t.Timestamp.AsTime().UnixMilli(),
			Bid:       fromProtoDecimal(t.Bid),
			Ask:       fromProtoDecimal(t.Ask),
			BidVolume: fromProtoDecimal(t.BidVolume),
			AskVolume: fromProtoDecimal(t.AskVolume),
		}
		then = t.Timestamp.AsTime()
	} else if data.GetTrade() != nil {
		t := data.GetTrade()
		rpcResponse.Params.Result = &trade{
			Exchange:  t.Exchange,
			Timestamp: t.Timestamp.AsTime().UnixMilli(),
			Price:     fromProtoDecimal(t.Price),
			Quantity:  fromProtoDecimal(t.Quantity),
			IsBuy:     t.IsBuy,
			TradeID:   t.TradeID,
		}
		then = t.Timestamp.AsTime()
	}
	if err != nil {
		return nil, errors.Wrap(err, "marshal protobuf payload")
	}

	go func() {
		latency := now - then.UnixMilli()
		latencyHistogram.Observe(float64(latency))
	}()

	return json.Marshal(rpcResponse)
}

func fromProtoDecimal(d *streamspb.Decimal) float64 {
	return decimal.New(int64(d.Value), d.Exponent).InexactFloat64()
}
