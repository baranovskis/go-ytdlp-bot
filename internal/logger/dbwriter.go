package logger

import (
	"encoding/json"
	"sync"
)

// LogEvent represents a parsed log entry for broadcasting.
type LogEvent struct {
	Level   string `json:"level"`
	Message string `json:"message"`
	Fields  string `json:"fields"`
}

// DBInserter is the interface for inserting logs into the database.
type DBInserter interface {
	InsertLog(level, message, fields string) (int64, error)
}

// DBWriter is a zerolog writer that writes log entries to SQLite and broadcasts them.
type DBWriter struct {
	db          DBInserter
	subscribers map[chan LogEvent]struct{}
	mu          sync.RWMutex
}

// NewDBWriter creates a new DBWriter.
func NewDBWriter(db DBInserter) *DBWriter {
	return &DBWriter{
		db:          db,
		subscribers: make(map[chan LogEvent]struct{}),
	}
}

// Write implements io.Writer for zerolog. It parses the JSON log line and stores it.
func (w *DBWriter) Write(p []byte) (n int, err error) {
	var raw map[string]any
	if err := json.Unmarshal(p, &raw); err != nil {
		return len(p), nil
	}

	level, _ := raw["level"].(string)
	message, _ := raw["message"].(string)

	// Remove known fields to store the rest as extra fields
	delete(raw, "level")
	delete(raw, "message")
	delete(raw, "time")

	fieldsJSON := "{}"
	if len(raw) > 0 {
		if b, err := json.Marshal(raw); err == nil {
			fieldsJSON = string(b)
		}
	}

	w.db.InsertLog(level, message, fieldsJSON)

	event := LogEvent{Level: level, Message: message, Fields: fieldsJSON}
	w.mu.RLock()
	for ch := range w.subscribers {
		select {
		case ch <- event:
		default:
		}
	}
	w.mu.RUnlock()

	return len(p), nil
}

// Subscribe returns a channel that receives new log events.
func (w *DBWriter) Subscribe() chan LogEvent {
	ch := make(chan LogEvent, 64)
	w.mu.Lock()
	w.subscribers[ch] = struct{}{}
	w.mu.Unlock()
	return ch
}

// Unsubscribe removes a subscriber channel.
func (w *DBWriter) Unsubscribe(ch chan LogEvent) {
	w.mu.Lock()
	delete(w.subscribers, ch)
	w.mu.Unlock()
	close(ch)
}
