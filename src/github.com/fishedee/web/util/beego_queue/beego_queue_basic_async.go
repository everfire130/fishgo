package beego_queue

import (
	"errors"
	"sync"
)

type BasicAsyncQueuePubSubStore struct {
	listener []BeegoQueueListener
	mutex    sync.RWMutex
}

type BasicAsyncQueueStore struct {
	BeegoQueueStoreBasicInterface
	mapPubSubStore map[string]*BasicAsyncQueuePubSubStore
	mutex          sync.Mutex
}

func NewBasicAsyncQueue(target BeegoQueueStoreBasicInterface) *BasicAsyncQueueStore {
	return &BasicAsyncQueueStore{
		BeegoQueueStoreBasicInterface: target,
		mapPubSubStore:                map[string]*BasicAsyncQueuePubSubStore{},
	}
}

func (this *BasicAsyncQueueStore) Publish(topicId string, data interface{}) error {
	return this.Produce(topicId, data)
}

func (this *BasicAsyncQueueStore) subscribeInner(topicId string, single *BasicAsyncQueuePubSubStore) error {
	return this.Consume(topicId, func(argv interface{}) error {
		var lastError error
		single.mutex.RLock()
		listeners := single.listener
		single.mutex.RUnlock()

		for _, singleListener := range listeners {
			err := singleListener(argv)
			if err != nil {
				if lastError == nil {
					lastError = errors.New(err.Error())
				} else {
					lastError = errors.New(lastError.Error() + "\n" + err.Error())
				}
			}
		}
		return lastError
	})
}

func (this *BasicAsyncQueueStore) Subscribe(topicId string, listener BeegoQueueListener) error {
	this.mutex.Lock()
	result, ok := this.mapPubSubStore[topicId]
	if !ok {
		result = &BasicAsyncQueuePubSubStore{}
		result.listener = []BeegoQueueListener{listener}
	}
	this.mapPubSubStore[topicId] = result
	this.mutex.Unlock()

	if !ok {
		return this.subscribeInner(topicId, result)
	} else {
		result.mutex.Lock()
		result.listener = append(result.listener, listener)
		result.mutex.Unlock()
		return nil
	}
}
