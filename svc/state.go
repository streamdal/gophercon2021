package main

import (
	"context"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"go.etcd.io/etcd/clientv3"
)

type State struct {
	ServiceName string
	EtcdClient  *clientv3.Client
	Cache       map[string]bool
	CacheMutex  *sync.RWMutex
	Flush       map[string]bool
	FlushMutex  *sync.Mutex
	NoOp        bool // If set, State won't do anything
}

func NewState(serviceName string) (*State, error) {
	if os.Getenv("ENABLE_STATE") == "" {
		return &State{
			ServiceName: serviceName,
			NoOp:        true,
		}, nil
	}

	log.Info("state tracking is enabled")

	client, err := clientv3.New(clientv3.Config{
		Endpoints: []string{"localhost:2379"},
	})
	if err != nil {
		return nil, errors.Wrap(err, "unable to setup etcd")
	}

	s := &State{
		ServiceName: serviceName,
		EtcdClient:  client,
		Cache:       make(map[string]bool),
		CacheMutex:  &sync.RWMutex{},
		Flush:       make(map[string]bool),
		FlushMutex:  &sync.Mutex{},
	}

	// Perform initial import
	if err := s.importCache(s.etcdPath()); err != nil {
		return nil, errors.Wrap(err, "unable to perform initial import")
	}

	// Start flusher
	s.startFlusher()

	return s, nil
}

func (s *State) Contains(key string) bool {
	if s.NoOp {
		return false
	}

	s.CacheMutex.RLock()
	defer s.CacheMutex.RUnlock()

	_, ok := s.Cache[key]

	return ok
}

func (s *State) Add(key string) {
	if s.NoOp {
		return
	}

	s.CacheMutex.Lock()
	defer s.CacheMutex.Unlock()

	s.Cache[key] = true

	s.FlushMutex.Lock()
	s.Flush[key] = true
	s.FlushMutex.Unlock()
}

// importCache will recursively read entries from etcd and update local cache
func (s *State) importCache(keyspace string) error {
	if s.NoOp {
		return nil
	}

	resp, err := s.EtcdClient.Get(context.TODO(), keyspace, clientv3.WithPrefix())
	if err != nil {
		return errors.Wrap(err, "unable to complete recursive fetch")
	}

	s.CacheMutex.Lock()
	defer s.CacheMutex.Unlock()

	for _, v := range resp.Kvs {
		messageID := strings.TrimPrefix(string(v.Key), s.etcdPath())

		log.Infof("caching message id '%s'", messageID)

		s.Cache[messageID] = true
	}

	return nil
}

// startFlusher will periodically dump all >new< cache entries to etcd
func (s *State) startFlusher() {
	if s.NoOp {
		return
	}

	go func() {
		ticker := time.NewTicker(time.Second)

		for {
			var numFlushed int

			<-ticker.C
			// Save cache contents to etcd
			s.FlushMutex.Lock()
			for k := range s.Flush {
				if _, err := s.EtcdClient.Put(context.TODO(), s.etcdPath()+k, "true"); err != nil {
					log.Errorf("unable to flush key '%s' to etcd: %s", k, err)
					continue
				}

				// Remove the entry from the flush "queue"
				delete(s.Flush, k)

				numFlushed += 1
			}
			s.FlushMutex.Unlock()

			if numFlushed > 0 {
				log.Infof("successfully flushed '%d' cached entries to etcd", numFlushed)
			}
		}
	}()
}

func (s *State) etcdPath() string {
	return "/service/" + s.ServiceName + "/"
}
