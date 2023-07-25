package event_streamer

import (
	"log"
	"sync"

	"websocket-poc/pkg/streamspb"
)

type registry struct {
	lock    sync.RWMutex
	streams map[chan *streamspb.Response]struct{}
}

func newRegistry() *registry {
	return &registry{
		lock:    sync.RWMutex{},
		streams: make(map[chan *streamspb.Response]struct{}),
	}
}

func (r *registry) add(c chan *streamspb.Response) {
	log.Println("adding channel")
	r.lock.Lock()
	defer r.lock.Unlock()
	r.streams[c] = struct{}{}
}

func (r *registry) remove(c chan *streamspb.Response) {
	log.Println("removing channel")
	r.lock.Lock()
	defer r.lock.Unlock()
	delete(r.streams, c)
}

func (r *registry) broadcast(res *streamspb.Response) {
	r.lock.RLock()
	defer r.lock.RUnlock()

	for c := range r.streams {
		c := c
		go func() {
			c <- res
		}()
	}
}
