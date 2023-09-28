package utils

import (
	"sync"
	"time"
	//	"github.com/bwmarrin/snowflake"
)

var snowflake Snowflake

// Snowflake 结构体
type Snowflake struct {
	mu        sync.Mutex
	timestamp int64 // 时间戳
	sequence  int64 // 序列号
}

// NewSnowflake 创建一个 Snowflake 实例
func InitSnowFlake() *Snowflake {
	return &Snowflake{}
}

// NextID 生成下一个 ID
func GenSnowFlakeID() int64 {
	snowflake.mu.Lock()
	defer snowflake.mu.Unlock()

	currentTimestamp := time.Now().Unix()

	// 如果当前时间戳与上次相同，则增加序列号
	if currentTimestamp == snowflake.timestamp {
		snowflake.sequence++
	} else {
		// 如果当前时间戳变化，则重置序列号
		snowflake.timestamp = currentTimestamp
		snowflake.sequence = 0
	}

	// 合并时间戳和序列号为一个 32 位整数
	id := (snowflake.timestamp << 16) | snowflake.sequence

	return id
}
