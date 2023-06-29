package tools

import (
	"changeme/logger"
	"context"
	"github.com/iawia002/lux/extractors"
	"github.com/iawia002/lux/extractors/acfun"
	"github.com/iawia002/lux/extractors/bcy"
	"github.com/iawia002/lux/extractors/bilibili"
	"github.com/iawia002/lux/extractors/douyin"
	"github.com/iawia002/lux/extractors/douyu"
	"github.com/iawia002/lux/extractors/facebook"
	"github.com/iawia002/lux/extractors/twitter"
	"github.com/iawia002/lux/extractors/youtube"
	"log"
	"testing"
)

func setUp() {
	extractors.Register("bilibili", bilibili.New())
	extractors.Register("acfun", acfun.New())
	extractors.Register("bcy", bcy.New())
	extractors.Register("douyin", douyin.New())
	extractors.Register("iesdouyin", douyin.New())
	extractors.Register("douyu", douyu.New())
	extractors.Register("facebook", facebook.New())
	extractors.Register("youtube", youtube.New())
	extractors.Register("youtu", youtube.New())
	extractors.Register("twitter", twitter.New())
	logger.InitLogger()
}

func TestExtractLink(t *testing.T) {
	setUp()
	// https://www.bilibili.com/video/BV1dM4y1E7Yu/?spm_id_from=333.1007.tianma.1-2-2.click
	u := "https://www.bilibili.com/video/BV1fM4y187e2/?spm_id_from=333.1007.tianma.1-1-1.click"
	//u := "https://www.acfun.cn/v/ac41618732"

	linkData, err := ExtractLink(u)
	if err != nil {
		log.Fatal(err)
		return
	}

	for _, data := range linkData {
		err = Download(context.Background(), data)
		if err != nil {
			log.Fatal(err)
			return
		}
	}

}
