package config

import (
	"context"
	"log/slog"
)

// WatchCallback 配置变更回调函数
type WatchCallback func(cfg *Config)

// Watcher 配置变更监听器
type Watcher struct {
	source   Source
	callback WatchCallback
}

// NewWatcher 创建配置变更监听器
func NewWatcher(source Source, callback WatchCallback) *Watcher {
	return &Watcher{source: source, callback: callback}
}

// Start 启动配置监听
func (w *Watcher) Start(ctx context.Context) error {
	ch, err := w.source.Watch(ctx)
	if err != nil {
		return err
	}
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case data, ok := <-ch:
				if !ok {
					return
				}
				cfg, err := parseConfig(data)
				if err != nil {
					slog.Error("解析配置变更失败，忽略本次更新", "error", err)
					continue
				}
				slog.Info("检测到配置变更，正在应用")
				GlobalConfig = cfg
				if w.callback != nil {
					w.callback(cfg)
				}
			}
		}
	}()
	return nil
}
