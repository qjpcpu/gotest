package main

import (
	"path/filepath"

	"github.com/qjpcpu/common.v2/assert"
	"github.com/qjpcpu/common.v2/cli"
)

const (
	gotestDir        = "gotest"
	gotestBucketSize = -1
)

var History HistoryTracker = historyFileTracker{holder: cli.MustNewHomeFileDB(gotestDir)}

type Item struct {
	Dir    string
	Test   string
	Module string
}

type HistoryTracker interface {
	Get(string) *Item
	Append(Item)
}

type historyFileTracker struct {
	holder *cli.FileDB
}

func dirnameAbs(dir string) string {
	p, err := filepath.Abs(dir)
	assert.ShouldBeNil(err)
	return p
}

func (t historyFileTracker) Get(dir string) *Item {
	var list []Item
	dir = dirnameAbs(dir)
	t.holder.GetItemHistoryBucket(dir, gotestBucketSize).ListItem(&list)
	if len(list) > 0 {
		return &list[0]
	}
	return nil
}

func (t historyFileTracker) Append(item Item) {
	item.Dir = dirnameAbs(item.Dir)
	t.holder.GetItemHistoryBucket(item.Dir, gotestBucketSize).InsertItem(item)
}
