package conf

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"google.golang.org/grpc"

	"github.com/coreos/etcd/clientv3"
	"gitlab.xfq.com/tech-lab/watcher"

	"gitlab.xfq.com/tech-lab/dionysus/pkg/algs"
	"gitlab.xfq.com/tech-lab/dionysus/pkg/grpool"
)

type watchEtcd struct {
	etcdClient     *clientv3.Client
	configMap      sync.Map
	watchingConfig sync.Map
}

var (
	defaultWatchEtcd = &watchEtcd{}
	projectName      string
)

func NewWatchEtcd(cfg clientv3.Config) (*watchEtcd, error) {
	cfg.DialOptions = append(cfg.DialOptions, grpc.WithBlock())
	if cfg.DialTimeout == 0 {
		cfg.DialTimeout = 2 * time.Second
	}
	var etcdCli *clientv3.Client
	var err error
	for i := 0; i < 3; i++ {
		etcdCli, err = clientv3.New(cfg)
		if err == nil {
			break
		}
	}
	if err != nil {
		return nil, fmt.Errorf("new etcd client error: %v", err)
	}
	log.Infof("new etcd client: %v for watch success", cfg.Endpoints)
	return &watchEtcd{
		etcdClient: etcdCli,
		configMap:  sync.Map{},
	}, nil
}

func (w *watchEtcd) StartWatchConfig(watchPrefix string) error {
	return w.RegisterEtcdWatch(configClient{prefix: watchPrefix, cache: &w.configMap})
}

func (w *watchEtcd) RegisterEtcdWatch(rw ReceiveWatch) error {
	if w.etcdClient == nil {
		return fmt.Errorf("etcd client is not init, can not watch etcd")
	}

	etcdPath := toEtcdNotation(rw.GetPrefix())
	if _, load := w.watchingConfig.LoadOrStore(etcdPath, struct{}{}); load {
		return fmt.Errorf(" The key: %s had been binded before. ", etcdPath)
	}

	watch := watcher.New(etcdPath, true, w.etcdClient, log)

	var e error
	err := watch.GetWithCallback(func(resp *watcher.Response) {
		defer func() {
			if re := recover(); re != nil {
				e = fmt.Errorf("recover in get handler %v", re)
			}
		}()
		if pe := rw.OnPut([]byte(toDionysusNotation(string(resp.Key))), resp.Value); pe != nil {
			e = fmt.Errorf("get event handler error: %v", pe)
			return
		}
	})

	if err != nil {
		return fmt.Errorf("registerEtcdWatch error, prefix:%v, err: %v", rw.GetPrefix(), err)
	}
	if e != nil {
		return fmt.Errorf("registerEtcdWatch onput error, prefix:%v, err: %v", rw.GetPrefix(), e)
	}
	err = grpool.Submit(func() {
		watch.WatchWithCallback(func(resp *watcher.Response) {
			defer func() {
				if e := recover(); e != nil {
					log.Errorf("recover in watch handler %v", e)
				}
			}()
			if resp.Event == clientv3.EventTypeDelete.String() {
				if err := rw.OnDelete([]byte(toDionysusNotation(string(resp.Key)))); err != nil {
					log.Errorf("watch event handler error: %v", err)
				}
			} else {
				if err := rw.OnPut([]byte(toDionysusNotation(string(resp.Key))), resp.Value); err != nil {
					log.Errorf("watch event handler error: %v", err)
				}
			}
		})
	})
	if err != nil {
		return fmt.Errorf("grpool Submit etcd config watch error: %v", err)
	}
	log.Infof("start watch etcdPath: %v success", etcdPath)
	return nil
}

// type receiveWatch = func(*watcher.Response) error
type ReceiveWatch interface {
	OnDelete(key []byte) error
	OnPut(key []byte, value []byte) error
	GetPrefix() string
}

type configClient struct {
	prefix string
	cache  *sync.Map
}

func (w configClient) OnDelete(key []byte) error {
	w.cache.Delete(string(key))
	return nil
}

func (w configClient) OnPut(key []byte, value []byte) error {
	w.cache.Store(string(key), value)
	return nil
}

func (w configClient) GetPrefix() string {
	return w.prefix
}

func InitEtcdClient(addrs []string, pn string) error {
	etcdConf := clientv3.Config{
		Endpoints:   addrs,
		DialTimeout: 2 * time.Second,
		DialOptions: []grpc.DialOption{grpc.WithBlock()},
	}

	projectName = pn

	var err error
	for i := 0; i < 3; i++ {
		if defaultWatchEtcd.etcdClient, err = clientv3.New(etcdConf); err == nil {
			break
		}
	}
	if err == nil {
		log.Infof("init etcd client: %v for success", etcdConf.Endpoints)
	}
	return err
}

// 维护数据在configMap中,conf.load
func StartWatchConfig(watchPrefix string) error {
	return defaultWatchEtcd.StartWatchConfig(watchPrefix)
}

func RegisterEtcdWatch(rw ReceiveWatch) error {
	return defaultWatchEtcd.RegisterEtcdWatch(rw)
}

func (w *watchEtcd) Load(key interface{}) (value interface{}, ok bool, err error) {
	if !algs.Comparable(key) {
		return nil, false, fmt.Errorf("key uncomparable: %v", key)
	}
	value, ok = w.configMap.Load(key)
	return value, ok, nil
}

func (w *watchEtcd) LoadBytes(key string) ([]byte, error) {
	if !algs.Comparable(key) {
		return nil, fmt.Errorf("key uncomparable: %v", key)
	}
	i, ok := w.configMap.Load(key)
	if i == nil || !ok {
		return nil, fmt.Errorf("can not get value of key: %v from etcd", key)
	}
	if data, ok := i.([]byte); ok {
		if len(data) == 0 {
			return nil, fmt.Errorf("get value of key: %v from etcd is nil", key)
		}
		return data, nil
	}
	return nil, fmt.Errorf("convert value of key: %v to byte failed", key)
}

func (w *watchEtcd) Store(key, value interface{}) error {
	if !algs.Comparable(key) {
		return fmt.Errorf("key uncomparable: %v", key)
	}

	w.configMap.Store(key, value)
	return nil
}

func (w *watchEtcd) Delete(key interface{}) error {
	if !algs.Comparable(key) {
		return fmt.Errorf("key uncomparable: %v", key)
	}

	w.configMap.Delete(key)
	return nil
}

func (w *watchEtcd) LoadOrStore(key, value interface{}) (actual interface{}, loaded bool, err error) {
	if !algs.Comparable(key) {
		return nil, false, fmt.Errorf("key uncomparable: %v", key)
	}

	actual, loaded = w.configMap.LoadOrStore(key, value)
	return actual, loaded, nil
}

func (w *watchEtcd) Range(f func(key, value interface{}) bool) {
	w.configMap.Range(f)
}

func (w *watchEtcd) LoadByteString(key interface{}) (value string, ok bool, err error) {
	if !algs.Comparable(key) {
		return "", false, fmt.Errorf("key uncomparable: %v", key)
	}
	i, ok := w.configMap.Load(key)
	if i == nil || !ok {
		return "", false, nil
	}

	if data, ok := i.(string); ok {
		return data, true, nil
	}

	if data, ok := i.([]byte); ok {
		if len(data) == 0 {
			return "", false, fmt.Errorf("get value of key: %v from etcd is nil", key)
		}
		return string(data), true, nil
	}
	return "", false, nil
}

//  foo.bar  --> /projectName/foo/bar/
func toEtcdNotation(rwPrefix string) string {
	if !strings.HasSuffix(rwPrefix, ".") {
		rwPrefix = rwPrefix + "."
	}
	return "/" + projectName + strings.ReplaceAll("."+rwPrefix, ".", "/")
}

//  /projectName/foo/bar -->  foo.bar
func toDionysusNotation(etcdPrefix string) string {
	return strings.ReplaceAll(strings.TrimPrefix(etcdPrefix, "/"+projectName+"/"), "/", ".")
}

func Load(key interface{}) (value interface{}, ok bool, err error) {
	return defaultWatchEtcd.Load(key)
}

func LoadBytes(key string) ([]byte, error) {
	return defaultWatchEtcd.LoadBytes(key)
}

func Store(key, value interface{}) error {
	return defaultWatchEtcd.Store(key, value)
}

func Delete(key interface{}) error {
	return defaultWatchEtcd.Delete(key)
}
func LoadOrStore(key, value interface{}) (actual interface{}, loaded bool, err error) {
	return defaultWatchEtcd.LoadOrStore(key, value)
}

func Range(f func(key, value interface{}) bool) {
	defaultWatchEtcd.Range(f)
}

func LoadByteString(key interface{}) (value string, ok bool, err error) {
	return defaultWatchEtcd.LoadByteString(key)
}

func GetProjectName() string {
	return projectName
}
