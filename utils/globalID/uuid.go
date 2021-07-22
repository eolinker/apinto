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
	lastId int64

	start int64
)

func init() {
	s, _ := time.Parse("2020-01-02T15:04:05Z07:00", time.RFC3339)
	start = s.UnixNano()
}
func GenerateID() int64 {
	id := time.Now().UnixNano() - start
	locker.Lock()
	defer locker.Unlock()

	for id <= lastId {
		id++
	}
	lastId = id
	return id
}
func GenerateIDString() string {
	id := GenerateID()
	data := make([]byte, 8)
	binary.BigEndian.PutUint64(data, uint64(id))
	return strings.ToUpper(hex.EncodeToString(data))
}
