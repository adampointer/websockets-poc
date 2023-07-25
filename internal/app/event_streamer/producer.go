package event_streamer

import (
	"math/rand"
	"time"

	"websocket-poc/pkg/streamspb"

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const messageRate = 100 * time.Millisecond

var (
	symbols   = []string{"btc_usdt", "eth_usdt", "xrp_usdt"}
	exchanges = []string{"bitmex", "binance", "bitfinex"}
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
					Timestamp: timestamppb.New(time.Now()),
					Bid:       &streamspb.Decimal{Value: 1, Exponent: -2},
					Ask:       &streamspb.Decimal{Value: 11, Exponent: -2},
					BidVolume: &streamspb.Decimal{Value: 1, Exponent: 2},
					AskVolume: &streamspb.Decimal{Value: 1, Exponent: 3},
				},
			}
		case streamspb.Feature_SPOT_TRADE:
			response.Payload = &streamspb.Response_Trade{
				Trade: &streamspb.Trade{
					Exchange:  exchanges[exchange],
					Timestamp: timestamppb.New(time.Now()),
					Price:     &streamspb.Decimal{Value: 1, Exponent: -2},
					Quantity:  &streamspb.Decimal{Value: 1, Exponent: 0},
					IsBuy:     false,
					TradeID:   uuid.NewString(),
				},
			}
		}
		pub.BroadcastMessage(response)
	}
}
