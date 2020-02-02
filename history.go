package main

import (
	"github.com/qjpcpu/common/debug"
	"github.com/qjpcpu/common/json"
	"io/ioutil"
	"os"
	"path/filepath"
)

var History HistoryTracker = historyFileTracker{}

type Item struct {
	Dir    string
	Test   string
	Module string
}

type HistoryTracker interface {
	Get(limit int) []Item
	Append(Item)
}

type historyFileTracker struct{}

func (t historyFileTracker) Get(limit int) (items []Item) {
	data, err := ioutil.ReadFile(t.filename())
	if err != nil {
		return
	}
	json.MustUnmarshal(data, &items)
	if len(items) > limit {
		items = items[:limit]
	}
	return
}

func (t historyFileTracker) Append(item Item) {
	items := append([]Item{item}, t.Get(10)...)
	ioutil.WriteFile(t.filename(), json.MustMarshal(items), 0644)
}

func (historyFileTracker) filename() string {
	dir, err := os.UserHomeDir()
	debug.ShouldBeNil(err)
	return filepath.Join(dir, ".gotest_history")
}
