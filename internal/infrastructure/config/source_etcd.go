package config

import (
	"context"
	"fmt"
	"sync"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

// EtcdSource etcd 配置源实现
type EtcdSource struct {
	cfg    *EtcdSourceConfig
	mu     sync.Mutex
	client *clientv3.Client
}

// NewEtcdSource 创建 etcd 配置源
func NewEtcdSource(cfg *EtcdSourceConfig) *EtcdSource {
	return &EtcdSource{cfg: cfg}
}

func (s *EtcdSource) Type() string { return "etcd" }

func (s *EtcdSource) getClient() (*clientv3.Client, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.client != nil {
		return s.client, nil
	}
	timeout := s.cfg.Timeout
	if timeout == 0 {
		timeout = 5 * time.Second
	}
	client, err := clientv3.New(clientv3.Config{
		Endpoints:   s.cfg.Endpoints,
		DialTimeout: timeout,
		Username:    s.cfg.Username,
		Password:    s.cfg.Password,
	})
	if err != nil {
		return nil, fmt.Errorf("连接 etcd 失败: %w", err)
	}
	s.client = client
	return s.client, nil
}

// Load 从 etcd 加载完整配置
func (s *EtcdSource) Load() ([]byte, error) {
	client, err := s.getClient()
	if err != nil {
		return nil, err
	}
	timeout := s.cfg.Timeout
	if timeout == 0 {
		timeout = 5 * time.Second
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	resp, err := client.Get(ctx, s.cfg.Key)
	if err != nil {
		return nil, fmt.Errorf("从 etcd 读取配置失败: %w", err)
	}
	if len(resp.Kvs) == 0 {
		return nil, fmt.Errorf("etcd 中未找到配置 key: %s", s.cfg.Key)
	}
	return resp.Kvs[0].Value, nil
}

// Watch 监听 etcd 配置变更
func (s *EtcdSource) Watch(ctx context.Context) (<-chan []byte, error) {
	client, err := s.getClient()
	if err != nil {
		return nil, err
	}
	ch := make(chan []byte, 1)
	go func() {
		defer close(ch)
		watchCh := client.Watch(ctx, s.cfg.Key)
		for {
			select {
			case <-ctx.Done():
				return
			case resp, ok := <-watchCh:
				if !ok {
					return
				}
				for _, event := range resp.Events {
					if event.Type == clientv3.EventTypePut && event.Kv != nil {
						select {
						case ch <- event.Kv.Value:
						default:
							<-ch
							ch <- event.Kv.Value
						}
					}
				}
			}
		}
	}()
	return ch, nil
}

// Close 关闭 etcd 客户端连接
func (s *EtcdSource) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.client != nil {
		err := s.client.Close()
		s.client = nil
		return err
	}
	return nil
}
