package globalID

import (
	"encoding/binary"
	"encoding/hex"
	"strings"
	"sync"
	"time"
)

var (
	locker = sync.Mutex{}
	lastID int64

	start int64
)

func init() {
	s, _ := time.Parse("2020-01-02T15:04:05Z07:00", time.RFC3339)
	start = s.UnixNano()
}

//GenerateID 生成id
func GenerateID() int64 {
	id := time.Now().UnixNano() - start
	locker.Lock()
	defer locker.Unlock()

	for id <= lastID {
		id++
	}
	lastID = id
	return id
}

//GenerateIDString 生成id字符串
func GenerateIDString() string {
	id := GenerateID()
	data := make([]byte, 8)
	binary.BigEndian.PutUint64(data, uint64(id))
	return strings.ToUpper(hex.EncodeToString(data))
}
