package tools

import (
	"changeme/logger"
	"sync"
	"testing"
)

func TestDownloadOptions_Process(t *testing.T) {
	logger.InitLogger()
	options := DownloadOptions{
		Eld: ExtractLinkData{
			Byte: 55259955,
		},
		mux: sync.RWMutex{},
	}

	options.Process(774144)
}
