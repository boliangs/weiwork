package discovery

import (
	"context"
	"google.golang.org/grpc/attributes"
	"sendswork/config"
	"time"

	"github.com/sirupsen/logrus"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc/resolver"
)

const (
	schema = "etcd"
)

// Resolver 用于 grpc 客户端
type Resolver struct {
	schema      string
	EtcdAddrs   []string
	DialTimeout int // 超时时间

	closeCh      chan struct{}
	watchCh      clientv3.WatchChan
	cli          *clientv3.Client
	keyPrifix    string
	srvAddrsList []resolver.Address

	cc     resolver.ClientConn
	logger *logrus.Logger
}

// NewResolver 基于 etcd 创建一个新的 resolver.Builder
func NewResolver(etcdAddrs []string, logger *logrus.Logger) *Resolver {
	return &Resolver{
		schema:      schema,
		EtcdAddrs:   etcdAddrs,
		DialTimeout: 3,
		logger:      logger,
	}
}

// Scheme 返回此 resolver 支持的 scheme
func (r *Resolver) Scheme() string {
	return r.schema
}

// Build 为给定的 target 创建一个新的 resolver.Resolver
func (r *Resolver) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	r.cc = cc

	r.keyPrifix = BuildPrefix(Server{Name: target.Endpoint(), Version: target.URL.Host})
	if _, err := r.start(); err != nil {
		return nil, err
	}
	return r, nil
}

// ResolveNow resolver.Resolver 接口
func (r *Resolver) ResolveNow(o resolver.ResolveNowOptions) {}

// Close resolver.Resolver 接口
func (r *Resolver) Close() {
	r.closeCh <- struct{}{}
}

// start 启动 resolver
func (r *Resolver) start() (chan<- struct{}, error) {
	var err error
	r.cli, err = clientv3.New(clientv3.Config{
		Username:    config.Conf.Etcd.Username,
		Password:    config.Conf.Etcd.Password,
		Endpoints:   r.EtcdAddrs,
		DialTimeout: time.Duration(r.DialTimeout) * time.Second,
	})
	if err != nil {
		return nil, err
	}
	resolver.Register(r)

	r.closeCh = make(chan struct{})

	if err = r.sync(); err != nil {
		return nil, err
	}

	go r.watch()

	return r.closeCh, nil
}

// watch 更新事件
func (r *Resolver) watch() {
	ticker := time.NewTicker(time.Minute)
	r.watchCh = r.cli.Watch(context.Background(), r.keyPrifix, clientv3.WithPrefix())

	for {
		select {
		case <-r.closeCh:
			return
		case res, ok := <-r.watchCh:
			if ok {
				r.update(res.Events)
			}
		case <-ticker.C:
			if err := r.sync(); err != nil {
				r.logger.Error("sync failed", err)
			}
		}
	}
}

// update 更新地址列表
func (r *Resolver) update(events []*clientv3.Event) {
	for _, ev := range events {
		var info Server
		var err error

		switch ev.Type {
		case clientv3.EventTypePut:
			info, err = ParseValue(ev.Kv.Value)
			if err != nil {
				continue
			}
			addr := resolver.Address{Addr: info.Addr, Attributes: attributes.New(info.Weight)}
			if !Exist(r.srvAddrsList, addr) {
				r.srvAddrsList = append(r.srvAddrsList, addr)
				err := r.cc.UpdateState(resolver.State{Addresses: r.srvAddrsList})
				if err != nil {
					return
				}
			}
		case clientv3.EventTypeDelete:
			info, err = SplitPath(string(ev.Kv.Key))
			if err != nil {
				continue
			}
			addr := resolver.Address{Addr: info.Addr}
			if s, ok := Remove(r.srvAddrsList, addr); ok {
				r.srvAddrsList = s
				r.cc.UpdateState(resolver.State{Addresses: r.srvAddrsList})
			}
		}
	}
}

// sync 同步获取所有地址信息
func (r *Resolver) sync() error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	res, err := r.cli.Get(ctx, r.keyPrifix, clientv3.WithPrefix())
	if err != nil {
		return err
	}
	r.srvAddrsList = []resolver.Address{}

	for _, v := range res.Kvs {
		info, err := ParseValue(v.Value)
		if err != nil {
			continue
		}
		addr := resolver.Address{Addr: info.Addr, Attributes: attributes.New(info.Weight)}
		r.srvAddrsList = append(r.srvAddrsList, addr)
	}
	rr := r.cc.UpdateState(resolver.State{Addresses: r.srvAddrsList})
	if rr != nil {
		return rr
	}
	return nil
}
