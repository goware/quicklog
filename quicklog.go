package quicklog

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

// TODO: perhaps we should add a redis backend..
// if we do though, lets keep track of "instance" id so we can correlate logs
// between services

// Quicklog is a simple in-memory log service for quick at-a-glance logs
type Quicklog interface {
	Info(group string, message string, v ...any)
	Warn(group string, message string, v ...any)
	Groups() []string
	Logs(group string) []LogEntry
	ToMap(timezone string, withExactTime bool, group ...string) map[string][]string
}

type LogEntry interface {
	Level() string
	Group() string
	Message() string
	Time() time.Time
	TimeAgo(timezone ...string) string
	Count() uint32
	FormattedMessage(timezone string, withExactTime ...bool) string
}

type quicklog struct {
	logs       map[string][]logEntry
	numEntries int
	mu         sync.RWMutex
}

func NewQuicklog(numEntries ...int) Quicklog {
	n := 40
	if len(numEntries) > 0 {
		n = numEntries[0]
	}
	if n < 1 {
		n = 1
	}
	if n > 500 {
		// its intended to be small, so operations are fast and memory usage is low
		n = 500
	}
	return &quicklog{
		logs:       make(map[string][]logEntry),
		numEntries: n,
	}
}

type logEntry struct {
	group   string
	message string
	level   string
	time    time.Time
	count   uint32
}

func (w *quicklog) Info(group, message string, v ...any) {
	w.log("INFO", group, message, v...)
}

func (w *quicklog) Warn(group, message string, v ...any) {
	w.log("WARN", group, message, v...)
}

func (w *quicklog) log(level, group, message string, v ...any) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if len(group) == 0 {
		return
	}

	g, ok := w.logs[group]
	if !ok {
		g = make([]logEntry, 0, w.numEntries)
	}

	msg := fmt.Sprintf(message, v...)
	if len(msg) == 0 {
		return
	}
	if len(msg) > 1000 {
		// truncate to 1000 characters per message
		msg = msg[:1000]
	}

	// First check if the message is already in the log, and if so, increment the count
	// and update the time
	for i, entry := range g {
		if entry.message == msg && entry.level == level {
			entry.count++
			entry.time = time.Now().UTC()
			g[i] = entry
			w.logs[group] = g
			return
		}
	}

	// Add the new entry
	newLen := len(g) + 1
	if newLen > w.numEntries {
		newLen = w.numEntries
	}
	newG := make([]logEntry, newLen)
	newEntry := logEntry{
		group:   group,
		message: msg,
		level:   level,
		time:    time.Now().UTC(),
		count:   1,
	}
	newG[0] = newEntry
	if len(g) > 0 {
		copy(newG[1:], g[:newLen-1])
	}
	w.logs[group] = newG
}

func (w *quicklog) Reset(group ...string) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if len(group) > 0 {
		delete(w.logs, group[0])
	} else {
		w.logs = make(map[string][]logEntry)
	}
}

func (w *quicklog) Groups() []string {
	w.mu.RLock()
	defer w.mu.RUnlock()
	keys := make([]string, 0, len(w.logs))
	for k := range w.logs {
		keys = append(keys, k)
	}
	return keys
}

func (w *quicklog) Logs(group string) []LogEntry {
	w.mu.RLock()
	defer w.mu.RUnlock()
	out := make([]LogEntry, 0, len(w.logs[group]))
	for _, entry := range w.logs[group] {
		out = append(out, entry)
	}
	return out
}

func (w *quicklog) ToMap(timezone string, withExactTime bool, group ...string) map[string][]string {
	m := make(map[string][]string)
	for k, v := range w.logs {
		if len(group) > 0 && !strings.HasPrefix(k, group[0]) {
			continue
		}
		for _, entry := range v {
			m[k] = append(m[k], entry.FormattedMessage(timezone, withExactTime))
		}
	}
	return m
}

func (l logEntry) Group() string {
	return l.group
}

func (l logEntry) Message() string {
	return l.message
}

func (l logEntry) Level() string {
	return l.level
}

func (l logEntry) Time() time.Time {
	return l.time
}

func (l logEntry) TimeAgo(timezone ...string) string {
	var err error
	loc := time.UTC
	if len(timezone) > 0 {
		loc, err = time.LoadLocation(timezone[0])
		if err != nil {
			loc = time.UTC
		}
	}

	duration := time.Since(l.time.In(loc))

	if duration < time.Minute {
		return fmt.Sprintf("%ds ago", int(duration.Seconds()))
	} else if duration < time.Hour {
		seconds := int(duration.Seconds()) % 60
		return fmt.Sprintf("%dm %ds ago", int(duration.Minutes()), seconds)
	} else if duration < 24*time.Hour {
		hours := int(duration.Hours())
		minutes := int(duration.Minutes()) % 60
		if minutes == 0 {
			return fmt.Sprintf("%dh ago", hours)
		}
		return fmt.Sprintf("%dh %dm ago", hours, minutes)
	}

	return l.time.In(loc).Format(time.RFC822)
}

func (l logEntry) Count() uint32 {
	return l.count
}

func (l logEntry) FormattedMessage(timezone string, withExactTime ...bool) string {
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		loc = time.UTC
	}
	var out string
	if len(withExactTime) > 0 && withExactTime[0] {
		out = fmt.Sprintf("%s - [%s] %s", l.time.In(loc).Format(time.RFC822), l.level, l.message)
	} else {
		out = fmt.Sprintf("%s - [%s] %s", l.TimeAgo(timezone), l.level, l.message)
	}
	if l.count > 1 {
		return fmt.Sprintf("%s [x%d]", out, l.count)
	} else {
		return out
	}
}
