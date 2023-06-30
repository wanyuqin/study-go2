package tools

import (
	"changeme/logger"
	"context"
	"fmt"
	"github.com/iawia002/lux/extractors"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"math"
	"os"
	"sync"
	"time"
)

var DownloadPercentRefresh = "download.percent.refresh"

// DownloadOptions 下载参数
type DownloadOptions struct {
	Data         *extractors.Data
	DownloadPath string
	// wails ctx
	Ctx context.Context

	Eld      ExtractLinkData
	mux      sync.RWMutex
	doneByte int64 // 已完成的数据大小
}

func (d *DownloadOptions) AddDoneByte(db int64) {
	d.mux.Lock()
	defer d.mux.Unlock()
	d.doneByte += db
	//atomic.AddInt64(&d.doneByte, db)
}

func (d *DownloadOptions) Process(wt int64) {
	if wt <= 0 {
		return
	}
	segment := int64(1000)
	if wt < segment {
		d.AddDoneByte(segment)
		d.CalculatePercent()
		return
	}

	n := wt / segment
	r := wt % segment

	for i := int64(0); i < n; i++ {
		d.AddDoneByte(segment)
		d.CalculatePercent()
	}

	if r > 0 {
		d.AddDoneByte(r)
		d.CalculatePercent()
	}
	return
}

// CalculatePercent 百分比计算
func (d *DownloadOptions) CalculatePercent() {
	d.mux.Lock()
	defer d.mux.Unlock()
	//d.Eld.Percentage = strconv.FormatFloat(float64(d.doneByte)/float64(d.Eld.Byte)*100, 'f', 2, 64)
	d.Eld.Percentage = math.Trunc(float64(d.doneByte) / float64(d.Eld.Byte) * 100)
	logger.Debug(fmt.Sprintf("download percent %f\n", d.Eld.Percentage))
	// 发送事件
	runtime.EventsEmit(d.Ctx, DownloadPercentRefresh, d)
}

func (d *DownloadOptions) StartWatchDownloadPercent() {
	ticker := time.NewTicker(1 * time.Second)
	for {
		select {
		case <-ticker.C:
			if d.doneByte > 0 {
				d.CalculatePercent()
			}

			if d.doneByte == d.Eld.Byte {
				ticker.Stop()
				return
			}
		}
	}
}

// DownloadList  下载列表
type DownloadList struct {
	mux sync.RWMutex

	// 维护一个下载中的列表
	Table map[string]struct{}

	// 维护下载中的临时文件地址
	DownloadTempPath map[string][]string
}

var dl *DownloadList

var once sync.Once

func GetDownloadList() *DownloadList {
	once.Do(func() {
		dl = NewDownloadList()
	})
	return dl
}

func NewDownloadList() *DownloadList {
	return &DownloadList{
		mux:              sync.RWMutex{},
		Table:            make(map[string]struct{}),
		DownloadTempPath: map[string][]string{},
	}
}

func (dl *DownloadList) Push(id string) {
	dl.mux.Lock()
	defer dl.mux.Unlock()

	dl.Table[id] = struct{}{}
}

func (dl *DownloadList) Pop(id string) {
	dl.mux.Lock()
	defer dl.mux.Unlock()
	delete(dl.Table, id)
}

func (dl *DownloadList) Length() int {
	dl.mux.Lock()
	defer dl.mux.Unlock()

	return len(dl.Table)
}

func (dl *DownloadList) PushTempPath(id string, path string) {
	dl.mux.Lock()
	defer dl.mux.Unlock()

	if _, ok := dl.DownloadTempPath[id]; !ok {
		dl.DownloadTempPath[id] = make([]string, 0)
	}

	dl.DownloadTempPath[id] = append(dl.DownloadTempPath[id], path)
}

func (dl *DownloadList) ClearTempFile(id string) {
	dl.mux.Lock()
	defer dl.mux.Unlock()

	tempFilePaths := dl.DownloadTempPath[id]
	var statErrs []error

	for _, path := range tempFilePaths {
		fi, statErr := os.Stat(path)
		if statErr != nil {
			statErrs = append(statErrs, statErr)
			continue
		}
		// 校验文件是否在被写
		for fi.Mode()&os.ModeAppend != 0 {
			logger.Debug(fmt.Sprintf("%s is still been write", fi.Name()))
		}

		os.Remove(path)
	}

}

// DownloadPool 管理下载任务
type DownloadPool struct {
	CtxMap map[string]context.CancelFunc
}

var dp *DownloadPool

var dpOnce sync.Once

func GetDownloadPool() *DownloadPool {
	dpOnce.Do(func() {
		dp = NewDownloadPool()
	})
	return dp

}

func NewDownloadPool() *DownloadPool {
	return &DownloadPool{
		CtxMap: make(map[string]context.CancelFunc),
	}
}

func (d *DownloadPool) Add(id string, cancel context.CancelFunc) {
	d.CtxMap[id] = cancel
}

func (d *DownloadPool) Cancel(id string) {
	cancelFunc := d.CtxMap[id]
	cancelFunc()
}
