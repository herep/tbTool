package watcher

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/coreos/etcd/clientv3"
	v3rpc "github.com/coreos/etcd/etcdserver/api/v3rpc/rpctypes"
)

type Response struct {
	Event    string
	Key      []byte
	Value    []byte
	Revision int64
}

type ResponseChan chan *Response

// Watcher ...
type Watcher interface {
	Get() ([]*Response, error)
	GetWithCallback(Callback) error
	Watch() ResponseChan
	WatchWithCallback(fn Callback)
	Close()
}

type logger interface {
	Debugf(template string, args ...interface{})
	Warnf(template string, args ...interface{})
	Errorf(template string, args ...interface{})
}

func New(path string, prefix bool, client *clientv3.Client, wlog logger) Watcher {
	return newWatcher(path, prefix, client, wlog)
}

func newWatcher(path string, prefix bool, client *clientv3.Client, logger logger) Watcher {
	ctx, cancel := context.WithCancel(context.Background())

	return &watcher{
		path:   path,
		prefix: prefix,
		client: client,
		logger: logger,
		ctx:    ctx,
		cancel: cancel,
		exitCh: make(chan struct{}, 1),
	}
}

type watcher struct {
	mu     sync.Mutex
	client *clientv3.Client
	logger logger

	ctx    context.Context
	cancel context.CancelFunc
	ticker *time.Ticker

	closed   bool
	path     string
	prefix   bool
	revision int64
	respChan ResponseChan
	exitCh   chan struct{}
}

type Callback func(*Response)

func (w *watcher) Get() ([]*Response, error) {
	var ops []clientv3.OpOption
	if w.prefix {
		ops = append(ops, clientv3.WithPrefix())
	}

	resp, err := w.client.Get(context.Background(), w.path, ops...)
	if err != nil {
		return nil, err
	}

	if resp.Header.Revision > w.getRevision() {
		w.setRevision(resp.Header.Revision)
	}

	res := make([]*Response, resp.Count)
	for i, kv := range resp.Kvs {
		res[i] = &Response{
			Event:    clientv3.EventTypePut.String(),
			Key:      kv.Key,
			Value:    kv.Value,
			Revision: kv.ModRevision,
		}
	}

	return res, nil
}

func (w *watcher) GetWithCallback(fn Callback) error {
	ch, err := w.Get()
	if err != nil {
		return err
	}

	for _, val := range ch {
		fn(val)
	}

	return nil
}

func (w *watcher) Watch() ResponseChan {
	w.respChan = make(ResponseChan, 100)

	go w.watch()

	return w.respChan
}

func (w *watcher) WatchWithCallback(fn Callback) {
	go func() {
		ch := w.Watch()
		for resp := range ch {
			fn(resp)
		}
	}()
}

func (w *watcher) watch() {
	w.logger.Debugf("start watch")

	w.ticker = time.NewTicker(time.Second * 30)

	for {
		var ops []clientv3.OpOption
		if revision := w.getRevision(); revision == 0 {
			ops = append(ops, clientv3.WithCreatedNotify())
		} else {
			ops = append(ops, clientv3.WithRev(revision+1))
		}

		if w.prefix {
			ops = append(ops, clientv3.WithPrefix())
		}

		watcher := clientv3.NewWatcher(w.client)

		rch := watcher.Watch(w.ctx, w.path, ops...)

		for {
			var (
				restart bool
				ok      bool
				resp    clientv3.WatchResponse
			)

			select {
			case <-w.ticker.C:
				restart = true
			case <-w.exitCh:
				_ = watcher.Close()

				w.logger.Debugf("watch quit")

				return
			case resp, ok = <-rch:
			}

			if restart || !ok {
				break
			}

			if resp.CompactRevision > w.getRevision() {
				w.setRevision(resp.CompactRevision)
			}
			if err := resp.Err(); err != nil {
				//switch {
				//case wr.closeErr != nil:
				//	return v3rpc.Error(wr.closeErr)
				//case wr.CompactRevision != 0:
				//	return v3rpc.ErrCompacted
				//case wr.Canceled:
				//	if len(wr.cancelReason) != 0 {
				//		return v3rpc.Error(status.Error(codes.FailedPrecondition, wr.cancelReason))
				//	}
				//	return v3rpc.ErrFutureRev
				//}
				if errors.Is(err, v3rpc.ErrCompacted) {
					break
				}

				if strings.Contains(err.Error(), "etcdserver: mvcc: required revision has been compacted") {
					break
				}

				//todo test for chan error type
				if clientv3.IsConnCanceled(err) {
					_ = watcher.Close()

					w.logger.Warnf("client connection is closing")

					return
				}

				w.logger.Errorf(err.Error())
				break
			}

			if resp.Header.Revision > w.getRevision() {
				w.setRevision(resp.Header.Revision)
			}

			for _, event := range resp.Events {
				w.respChan <- &Response{
					Event:    event.Type.String(),
					Key:      event.Kv.Key,
					Value:    event.Kv.Value,
					Revision: event.Kv.ModRevision,
				}

				w.logger.Debugf("receive event: %v", fmt.Sprintf("event: %v, key: %v, value: %v",
					event.Type.String(),
					string(event.Kv.Key),
					string(event.Kv.Value)))
			}
		}

		if err := watcher.Close(); err != nil {
			w.logger.Errorf("watcher close err: %s", err.Error())
		}

		w.logger.Debugf("restart watcher")
	}
}

func (w *watcher) getRevision() int64 {
	return atomic.LoadInt64(&w.revision)
}

func (w *watcher) setRevision(rev int64) {
	atomic.StoreInt64(&w.revision, rev)
}

/*
func (w *watcher) log() *loggerEntry {
	revision := w.getRevision()

	return w.logger.Type("watcher").Fields(map[string]interface{}{
		"path":     w.path,
		"prefix":   w.prefix,
		"revision": revision,
	})
}
*/

// Close 关闭 watcher
func (w *watcher) Close() {
	w.mu.Lock()
	defer w.mu.Unlock()

	if !w.closed {
		w.closed = true

		w.ticker.Stop()
		w.cancel()
		w.exitCh <- struct{}{}

		close(w.respChan)
	}
}
