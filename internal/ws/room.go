package ws

import (
	"fmt"
	"log"
	"sync"
	"time"

	jsonpatch "github.com/evanphx/json-patch/v5"
)

// Room代表一个协同编辑的房间
// 核心职责: 维护最新的页面状态快照
type Room struct {
	ID           string
	CurrentState []byte // 内存中的最新状态
	Version      int64  // 乐观锁版本号
	Clients      map[*Client]bool
	mu           sync.RWMutex
	LastActive   time.Time

	// 定时刷盘机制
	lastPersistedVersion int64         // 上次持久化的版本
	dirtyPatchCount      int           // 脏数据计数器
	flushTicker          *time.Ticker  // 定时刷盘
	stopFlush            chan struct{} // 停止信号
	pageService          PageService   // 数据库服务
}

// 刷盘配置
const (
	FlushInterval  = 30 * time.Second //每30s刷一次
	FlushThreshold = 50               // 每50个Patch刷一次
)

// 创建一个新房间 + 启动定时刷盘
func NewRoom(id string, initialState []byte, pageService PageService) *Room {
	r := &Room{
		ID:           id,
		CurrentState: initialState,
		Version:      1,
		Clients:      make(map[*Client]bool),
		LastActive:   time.Now(),
		flushTicker:  time.NewTicker(FlushInterval),
		stopFlush:    make(chan struct{}),
		pageService:  pageService,
	}

	go r.flushLoop()

	return r
}

// flushLoop启动定时刷盘循环
func (r *Room) flushLoop() {
	for {
		select {
		case <-r.flushTicker.C:
			r.flushToDB("定时")
		case <-r.stopFlush:
			r.flushToDB("销毁前")
		}
	}
}

// flushToDB 将当前状态刷写到数据库
func (r *Room) flushToDB(reason string) {
	// 读写保护下，快速复制数据
	r.mu.RLock()
	if r.Version == r.lastPersistedVersion {
		r.mu.RUnlock()
		return
	}

	snapshot := make([]byte, len(r.CurrentState))
	copy(snapshot, r.CurrentState)
	version := r.Version
	r.mu.RUnlock()

	// 无锁模式下，慢速写入数据库
	if err := r.pageService.SavePageState(r.ID, snapshot, version); err != nil {
		log.Printf("[Room %s] ⚠️ %s刷盘失败: %v", r.ID, reason, err)
		return
	}

	r.mu.Lock()
	r.lastPersistedVersion = version
	r.dirtyPatchCount = 0
	r.mu.Unlock()
	log.Printf("[Room %s] ✅ %s刷盘, 版本: %d", r.ID, reason, version)
}

// Stop 停止定时刷盘 (房间销毁时调用)
func (r *Room) Stop() {
	r.flushTicker.Stop()
	close(r.stopFlush)
}

// ApplyPatch 应用 Patch 并更新内存状态
func (r *Room) ApplyPatch(patchBytes []byte) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	patch, err := jsonpatch.DecodePatch(patchBytes)
	if err != nil {
		return fmt.Errorf("patch 解析失败: %w", err)
	}

	modified, err := patch.Apply(r.CurrentState)
	if err != nil {
		return fmt.Errorf("patch 应用失败: %w", err)
	}

	r.CurrentState = modified
	r.Version++
	r.LastActive = time.Now()
	r.dirtyPatchCount++

	// 超过阈值立即触发异步刷盘
	if r.dirtyPatchCount >= FlushThreshold {
		go r.flushToDB("阈值触发")
	}

	return nil
}

// GetSnapshot 获取当前快照（用于新用户加入）
func (r *Room) GetSnapshot() ([]byte, int64) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	snapshot := make([]byte, len(r.CurrentState))
	copy(snapshot, r.CurrentState)

	return snapshot, r.Version
}

// TODO：好像没用到Clients?
