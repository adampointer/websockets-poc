package pkg

import (
	"sync"

	"websocket-poc/pkg/streamspb"
)

type key struct {
	symbol, exchange string
	feature          streamspb.Feature
}

func keyFromProto(proto *streamspb.SubscriptionID) key {
	return key{
		symbol:   proto.Symbol,
		exchange: proto.Exchange,
		feature:  proto.Feature,
	}
}

func keyToProto(k key) *streamspb.SubscriptionID {
	return &streamspb.SubscriptionID{
		Symbol:   k.symbol,
		Exchange: k.exchange,
		Feature:  k.feature,
	}
}

type Subscriptions struct {
	lock          sync.RWMutex
	subscriptions map[key]uint32
}

func NewSubscriptions() *Subscriptions {
	return &Subscriptions{
		lock:          sync.RWMutex{},
		subscriptions: make(map[key]uint32),
	}
}

func (s *Subscriptions) Add(sub *streamspb.SubscriptionID) {
	k := keyFromProto(sub)

	s.lock.Lock()
	defer s.lock.Unlock()

	if count, exists := s.subscriptions[k]; exists {
		s.subscriptions[k] = count + 1
	} else {
		s.subscriptions[k] = 1
	}
}

func (s *Subscriptions) Remove(sub *streamspb.SubscriptionID) {
	k := keyFromProto(sub)

	s.lock.Lock()
	defer s.lock.Unlock()

	if count, exists := s.subscriptions[k]; exists {
		if count == 1 {
			delete(s.subscriptions, k)
		} else {
			s.subscriptions[k] = count - 1
		}
	}
}

func (s *Subscriptions) ForEach(f func(sub *streamspb.SubscriptionID)) {
	for k, count := range s.subscriptions {
		if count < 1 {
			continue
		}
		f(keyToProto(k))
	}
}

func (s *Subscriptions) HasSubscription(sub *streamspb.SubscriptionID) bool {
	symbolWildcard := &streamspb.SubscriptionID{
		Symbol:   "all",
		Exchange: sub.Exchange,
		Feature:  sub.Feature,
	}
	exchangeWildcard := &streamspb.SubscriptionID{
		Symbol:   sub.Symbol,
		Exchange: "all",
		Feature:  sub.Feature,
	}
	_, matchesSymbol := s.subscriptions[keyFromProto(symbolWildcard)]
	_, matchesExchange := s.subscriptions[keyFromProto(exchangeWildcard)]
	_, matchesExactly := s.subscriptions[keyFromProto(sub)]

	return matchesExchange || matchesSymbol || matchesExactly
}
