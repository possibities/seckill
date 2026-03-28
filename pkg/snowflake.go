package pkg

import (
	"fmt"
	"sync"
	"time"
)

const (
	workerIDBits  = 10
	sequenceBits  = 12
	maxWorkerID   = (1 << workerIDBits) - 1
	maxSequence   = (1 << sequenceBits) - 1
	workerIDShift = sequenceBits
	timeShift     = workerIDBits + sequenceBits
)

var defaultEpoch = time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC).UnixMilli()

type Snowflake struct {
	mu            sync.Mutex
	lastTimestamp int64
	sequence      int64
	workerID      int64
	epoch         int64
}

func NewSnowflake(workerID int64) (*Snowflake, error) {
	if workerID < 0 || workerID > maxWorkerID {
		return nil, fmt.Errorf("workerID out of range: %d", workerID)
	}

	return &Snowflake{
		workerID: workerID,
		epoch:    defaultEpoch,
	}, nil
}

func (s *Snowflake) NextID() (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UnixMilli()
	if now < s.lastTimestamp {
		return 0, fmt.Errorf("clock moved backwards: now=%d last=%d", now, s.lastTimestamp)
	}

	if now == s.lastTimestamp {
		s.sequence = (s.sequence + 1) & maxSequence
		if s.sequence == 0 {
			now = waitUntilNextMillis(s.lastTimestamp)
		}
	} else {
		s.sequence = 0
	}

	s.lastTimestamp = now
	id := ((now - s.epoch) << timeShift) | (s.workerID << workerIDShift) | s.sequence
	return id, nil
}

func waitUntilNextMillis(last int64) int64 {
	now := time.Now().UnixMilli()
	for now <= last {
		now = time.Now().UnixMilli()
	}
	return now
}
