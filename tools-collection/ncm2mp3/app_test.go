package main

import (
	"fmt"
	"github.com/wanyuqin/lux/downloader"
	"github.com/wanyuqin/lux/extractors"
	"github.com/wanyuqin/lux/extractors/bilibili"
	"log"
	"testing"
)

func TestFindNcmList(t *testing.T) {
	list, err := FindNcmList("/Users/ethanleo/GolandProjects/study-go2/go-ncm")
	if err != nil {
		log.Fatal(err)
	}

	for _, item := range list {
		fmt.Printf("%#v", item)
	}
}

func TestDownloader(t *testing.T) {
	//u := "https://www.bilibili.com/video/BV1dM4y1E7Yu/?spm_id_from=333.1007.tianma.1-2-2.click"
	u := "https://www.youtube.com/watch?v=MJ8QbnI3oVI&list=RDMJ8QbnI3oVI&start_radio=1&ab_channel=ScottLi"
	data, err := extractors.Extract(u, extractors.Options{
		Playlist: true,
		Items:    "",
	})

	if err != nil {
		log.Fatal(err)
	}
	extractors.Register("bilibili", bilibili.New())
	defaultDownloader := downloader.New(downloader.Options{})
	errors := make([]error, 0)
	for _, item := range data {
		if item.Err != nil {
			errors = append(errors, item.Err)
			continue
		}

		if err = defaultDownloader.Download(item); err != nil {
			errors = append(errors, err)
		}
	}

}
