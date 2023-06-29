package tools

import (
	"changeme/logger"
	"context"
	"fmt"
	"github.com/iawia002/lux/extractors"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"math"
	"sync"
	"time"
)

var DownloadPercentRefresh = "download.percent.refresh"

type DownloadOptions struct {
	Data         *extractors.Data
	DownloadPath string
	Ctx          context.Context

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
