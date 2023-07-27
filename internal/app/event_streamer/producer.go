package event_streamer

import (
	"math/rand"
	"time"

	"websocket-poc/pkg/streamspb"

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const messageRate = time.Millisecond

var (
	symbols   = []string{"btc_usdt", "eth_usdt", "xrp_usdt", "ada_usdt", "doge_usdt", "shib_usdt", "sol_usdt", "ldo_usdt", "avax_usdt", "fil_usdt"}
	exchanges = []string{"bitmex", "binance", "bitfinex", "deribit", "kraken", "kucoin", "coinbase", "okx", "bitstamp", "huobi"}
)

type Publisher interface {
	BroadcastMessage(message *streamspb.Response)
}

func ProduceMessages(pub Publisher, feature streamspb.Feature) {
	ticker := time.NewTicker(messageRate)

	for range ticker.C {
		symbol := rand.Intn(len(symbols))
		exchange := rand.Intn(len(exchanges))
		response := &streamspb.Response{
			Subscription: &streamspb.SubscriptionID{
				Symbol:   symbols[symbol],
				Exchange: exchanges[exchange],
				Feature:  feature,
			},
		}
		switch feature {
		case streamspb.Feature_SPOT_TICKER:
			response.Payload = &streamspb.Response_Ticker{
				Ticker: &streamspb.Ticker{
					Exchange:  exchanges[exchange],
					Timestamp: timestamppb.New(time.Now().UTC()),
					Bid:       &streamspb.Decimal{Value: 1, Exponent: -2},
					Ask:       &streamspb.Decimal{Value: 11, Exponent: -1},
					BidVolume: &streamspb.Decimal{Value: 478, Exponent: 1},
					AskVolume: &streamspb.Decimal{Value: 34, Exponent: 2},
				},
			}
		case streamspb.Feature_SPOT_TRADE:
			response.Payload = &streamspb.Response_Trade{
				Trade: &streamspb.Trade{
					Exchange:  exchanges[exchange],
					Timestamp: timestamppb.New(time.Now().UTC()),
					Price:     &streamspb.Decimal{Value: 1, Exponent: -2},
					Quantity:  &streamspb.Decimal{Value: 56, Exponent: 0},
					IsBuy:     false,
					TradeID:   uuid.NewString(),
				},
			}
		}
		pub.BroadcastMessage(response)
	}
}
