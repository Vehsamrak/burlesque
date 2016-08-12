package storage

import (
  "github.com/siddontang/ledisdb/ledis"
  "github.com/siddontang/ledisdb/config"
)

type (
	Storage struct {
    l     *ledis.Ledis
		db    *ledis.DB
	}
)

func New(path string) (s *Storage, err error) {
  var (
    l *ledis.Ledis
    db *ledis.DB
  )

  cfg := config.NewConfigDefault()
	cfg.DBName = "leveldb"
	cfg.DataDir = path
  cfg.LevelDB.Compression = true
  cfg.LevelDB.BlockSize = 512 * 1024 * 1024
  cfg.LevelDB.CacheSize =  64 * 1024 * 1024 * 1024
  cfg.LevelDB.WriteBufferSize = 256 * 1024 * 1024
  cfg.DBSyncCommit = 0
  cfg.ConnKeepaliveInterval = 0
  cfg.ConnReadBufferSize = 64 * 1024 * 1024
  cfg.ConnWriteBufferSize = 64 * 1024 * 1024

	if l, err = ledis.Open(cfg); err != nil {
    return
  }

  if db, err = l.Select(0); err != nil {
    return
  }

	s = &Storage{
    l: l,
		db: db,
	}
	return
}

func (s *Storage) Get(queue string) (message []byte, ok bool) {
  queue_key := []byte(queue)
  if count, _ := s.db.LLen(queue_key); count == 0 {
    return
  }

  values, err := s.db.BLPop([][]byte{queue_key}, 1)
	if values == nil || err != nil {
		return
	}

  message = values[1].([]byte)
	ok = true
	return
}

func (s *Storage) Put(queue string, message []byte) (err error) {
	_, err = s.db.RPush([]byte(queue), message)
	return
}

func (s *Storage) Flush(queue string) (messages [][]byte) {
  s.db.LClear([]byte(queue))
	return
}

func (s *Storage) QueueSizes() map[string]int64 {
  var count int64
	info := make(map[string]int64)
  members, _ := s.db.Scan(ledis.LIST, nil, 100, true, "")
  for i := range members {
    count, _ = s.db.LLen(members[i])
    info[string(members[i])] = count
  }

	return info
}

func (s *Storage) Close() (err error) {
	s.l.Close()
	return
}
