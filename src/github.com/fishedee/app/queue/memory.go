package queue

import (
	. "github.com/fishedee/util"
	"sync"
)

type MemoryQueuePushPopStore struct {
	listener QueueListener
}

type MemoryQueueStore struct {
	mapPushPopStore map[string]MemoryQueuePushPopStore
	mutex           sync.RWMutex
}

func NewMemoryQueue(closeFunc *CloseFunc, config QueueStoreConfig) (QueueStoreInterface, error) {
	result := &MemoryQueueStore{}
	result.mapPushPopStore = map[string]MemoryQueuePushPopStore{}
	return NewBasicQueue(result), nil
}

func (this *MemoryQueueStore) Produce(topicId string, data []byte) error {
	this.mutex.RLock()
	result, ok := this.mapPushPopStore[topicId]
	this.mutex.RUnlock()
	if !ok {
		return nil
	}
	result.listener(data, nil)
	return nil
}

func (this *MemoryQueueStore) Consume(topicId string, listener QueueListener) error {
	this.mutex.Lock()
	this.mapPushPopStore[topicId] = MemoryQueuePushPopStore{
		listener: listener,
	}
	this.mutex.Unlock()
	return nil
}
