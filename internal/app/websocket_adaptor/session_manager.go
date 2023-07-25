package websocket_adaptor

import (
	"sync"

	"websocket-poc/pkg/streamspb"

	"github.com/google/uuid"
	"github.com/pkg/errors"
)

var sessions = newSessionManager()

type sessionChannels struct {
	dataC    chan *streamspb.Response
	controlC chan *streamspb.Request
}

type sessionManager struct {
	lock     sync.RWMutex
	sessions map[uuid.UUID]*sessionChannels
}

func newSessionManager() *sessionManager {
	return &sessionManager{
		lock:     sync.RWMutex{},
		sessions: make(map[uuid.UUID]*sessionChannels),
	}
}

func (m *sessionManager) add(id uuid.UUID, dataC chan *streamspb.Response, controlC chan *streamspb.Request) {
	m.lock.Lock()
	defer m.lock.Unlock()

	m.sessions[id] = &sessionChannels{
		dataC:    dataC,
		controlC: controlC,
	}
}

func (m *sessionManager) remove(id uuid.UUID) {
	m.lock.Lock()
	defer m.lock.Unlock()

	if s, ok := m.sessions[id]; ok {
		s.controlC <- &streamspb.Request{
			Action: streamspb.Action_REMOVE,
		}
	}

	delete(m.sessions, id)
}

func (m *sessionManager) broadcast(res *streamspb.Response) {
	m.lock.RLock()
	defer m.lock.RUnlock()

	for _, c := range m.sessions {
		c := c
		go func() {
			c.dataC <- res
		}()
	}
}

func (m *sessionManager) get(id uuid.UUID) (*sessionChannels, error) {
	m.lock.RLock()
	defer m.lock.RUnlock()

	ses, ok := m.sessions[id]

	if !ok {
		return nil, errors.Errorf("no session found for %s", id.String())
	}

	return ses, nil
}
