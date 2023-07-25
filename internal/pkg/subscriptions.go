package pkg

import (
	"log"
	"sync"

	"websocket-poc/pkg/streamspb"

	"google.golang.org/protobuf/encoding/protojson"
)

type Subscriptions struct {
	lock          sync.RWMutex
	subscriptions map[string]uint16
}

func NewSubscriptions() *Subscriptions {
	return &Subscriptions{
		lock:          sync.RWMutex{},
		subscriptions: make(map[string]uint16),
	}
}

func (s *Subscriptions) Add(key *streamspb.SubscriptionID) {
	keyStr := keyToString(key)

	s.lock.Lock()
	defer s.lock.Unlock()

	if count, exists := s.subscriptions[keyStr]; exists {
		s.subscriptions[keyStr] = count + 1
	} else {
		s.subscriptions[keyStr] = 1
	}
}

func (s *Subscriptions) Remove(key *streamspb.SubscriptionID) {
	keyStr := keyToString(key)

	s.lock.Lock()
	defer s.lock.Unlock()

	if count, exists := s.subscriptions[keyStr]; exists {
		if count == 1 {
			delete(s.subscriptions, keyStr)
		} else {
			s.subscriptions[keyStr] = count - 1
		}
	}
}

func (s *Subscriptions) ForEach(f func(id *streamspb.SubscriptionID)) {
	for keyStr, count := range s.subscriptions {
		if count < 1 {
			continue
		}
		key := keyFromString(keyStr)
		f(key)
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
	_, matchesSymbol := s.subscriptions[keyToString(symbolWildcard)]
	_, matchesExchange := s.subscriptions[keyToString(exchangeWildcard)]
	_, matchesExactly := s.subscriptions[keyToString(sub)]

	return matchesExchange || matchesSymbol || matchesExactly
}

func keyToString(key *streamspb.SubscriptionID) string {
	bs, err := protojson.Marshal(key)
	if err != nil {
		log.Fatal(err)
	}
	return string(bs)
}

func keyFromString(keyStr string) *streamspb.SubscriptionID {
	var out streamspb.SubscriptionID
	if err := protojson.Unmarshal([]byte(keyStr), &out); err != nil {
		log.Fatal(err)
	}
	return &out
}
