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
	Get(string) *Item
	Append(Item)
}

type historyFileTracker struct{}

func dirnameAbs(dir string) string {
	p, err := filepath.Abs(dir)
	debug.ShouldBeNil(err)
	return p
}

func (t historyFileTracker) Get(dir string) *Item {
	list := t.GetAll()
	dir = dirnameAbs(dir)
	for i, item := range list {
		if dirnameAbs(item.Dir) == dir {
			return &list[i]
		}
	}
	return nil
}

func (t historyFileTracker) GetAll() (items []Item) {
	data, err := ioutil.ReadFile(t.filename())
	if err != nil {
		return
	}
	json.MustUnmarshal(data, &items)
	return
}

func (t historyFileTracker) Append(item Item) {
	all := t.GetAll()
	found := false
	item.Dir = dirnameAbs(item.Dir)
	for i, v := range all {
		if dirnameAbs(v.Dir) == item.Dir {
			all[i] = item
			found = true
			break
		}
	}
	if !found {
		all = append(all, item)
	}
	ioutil.WriteFile(t.filename(), json.MustMarshal(all), 0644)
}

func (historyFileTracker) filename() string {
	dir, err := os.UserHomeDir()
	debug.ShouldBeNil(err)
	return filepath.Join(dir, ".gotest_history")
}
