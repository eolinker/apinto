package http_context

import (
	"sync"

	"github.com/eolinker/eosc/formatter"
)

var (
	_            formatter.IEntry = (*Entry)(nil)
	proxiesChild                  = "proxies"
)

type Entry struct {
	fields  map[string]string
	childes map[string][]*Entry
	locker  sync.RWMutex
}

func NewEntry() *Entry {
	return &Entry{fields: make(map[string]string), childes: nil, locker: sync.RWMutex{}}
}

func (e *Entry) SetField(key string, value string) {
	e.locker.Lock()
	defer e.locker.Unlock()
	e.fields[key] = value
}

func (e *Entry) SetChildren(name string, fields []map[string]string) {
	e.locker.Lock()
	defer e.locker.Unlock()

	fieldLen := len(fields)
	entries := make([]*Entry, fieldLen)
	for i, field := range fields {
		entry := NewEntry()
		entry.fields = field
		entries[i] = entry
	}
	for key, value := range fields[fieldLen-1] {
		e.fields[key] = value
	}
	if name == "" {
		name = proxiesChild
	}
	e.childes[name] = entries
}

func (e *Entry) Read(pattern string) string {
	e.locker.RLock()
	defer e.locker.RUnlock()
	if v, ok := e.fields[pattern]; ok {
		return v
	}
	return ""
}

func (e *Entry) Children(child string) []formatter.IEntry {
	e.locker.RLock()
	defer e.locker.RUnlock()
	if child == "" {
		child = proxiesChild
	}
	if v, ok := e.childes[child]; ok {
		entries := make([]formatter.IEntry, len(v))
		for i, value := range v {
			entries[i] = value
		}
		return entries
	}
	return nil
}
