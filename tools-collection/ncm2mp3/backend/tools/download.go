package tools

import (
	"changeme/configs"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/iawia002/lux/extractors"
	"github.com/iawia002/lux/request"
	"github.com/iawia002/lux/utils"
	"io"
	"net/http"
	"os"
	"sort"
	"sync"
	"time"
)

var defaultThreadNumber = 10
var defaultRetryTimes = 3

var LinkDataMap map[string]*extractors.Data

type ExtractLinkData struct {
	Id      string `json:"id"`
	Title   string `json:"title"`
	Type    string `json:"type"`
	Url     string `json:"url"`
	Quality string `json:"quality"`
	Size    string `json:"size"`
}

type StreamInfo struct {
	Quality string `json:"quality"`
	Size    string `json:"size"`
}

func init() {
	LinkDataMap = make(map[string]*extractors.Data)
}

// ExtractLink 解析地址网页内容
func ExtractLink(link string) ([]ExtractLinkData, error) {
	data, err := extractors.Extract(link, extractors.Options{
		Playlist: false,
		Items:    "",
	})

	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	elds := make([]ExtractLinkData, 0, len(data))
	for i, item := range data {
		uid, err := uuid.NewUUID()
		if err != nil {
			fmt.Printf(err.Error())
			continue
		}

		sortStreams := GenSortedStreams(item.Streams)

		streamName := sortStreams[0].ID

		stream, ok := item.Streams[streamName]

		if !ok {
			continue
		}

		streamInfo := GetStreamInfo(stream)
		LinkDataMap[uid.String()] = data[i]
		eld := ExtractLinkData{
			Title:   item.Title,
			Url:     item.URL,
			Type:    string(item.Type),
			Id:      uid.String(),
			Size:    streamInfo.Size,
			Quality: streamInfo.Quality,
		}

		elds = append(elds, eld)
	}

	return elds, err
}

type DownloadOptions struct {
	Data         *extractors.Data
	DownloadPath string
}

func Download(id string) error {
	data, ok := LinkDataMap[id]
	if !ok {
		return errors.New("数据未找到")
	}
	// 获取配置
	config := configs.GetConfig()
	// 下载路径校验
	err := config.CheckDownloadPath()
	if err != nil {
		return err
	}

	options := DownloadOptions{
		Data:         data,
		DownloadPath: config.Download.Path,
	}

	err = download(options)

	return err
}

func download(options DownloadOptions) error {
	data := options.Data
	if len(data.Streams) == 0 {
		return errors.New(fmt.Sprintf("no streams in title %s", data.Title))
	}

	sortStreams := GenSortedStreams(data.Streams)

	title := data.Title

	streamName := sortStreams[0].ID

	stream, ok := data.Streams[streamName]

	if !ok {
		return errors.New(fmt.Sprintf("no stream named %s", streamName))
	}

	streamInfo := GetStreamInfo(stream)
	fmt.Printf("%v\n", streamInfo)

	if data.Captions != nil {
		for k, v := range data.Captions {
			if v != nil {
				fmt.Printf("Downloading %s ...\n", k)
				Caption(v.URL, title, v.Ext, v.Transform, options)
			}
		}
	}

	mergedFilePath, err := utils.FilePath(title, stream.Ext, 0, options.DownloadPath, false)

	if err != nil {
		return err
	}

	_, mergedFileExists, err := utils.FileSize(mergedFilePath)
	if err != nil {
		return err
	}
	// After the merge, the file size has changed, so we do not check whether the size matches
	if mergedFileExists {
		fmt.Printf("%s: file already exists, skipping\n", mergedFilePath)
		return nil
	}
	//if len(stream.Parts) == 1 {
	//
	//}

	wgp := utils.NewWaitGroupPool(defaultThreadNumber)
	errs := make([]error, 0)
	lock := sync.Mutex{}
	parts := make([]string, len(stream.Parts))

	for index, part := range stream.Parts {
		if len(errs) > 0 {
			break
		}
		partFileName := fmt.Sprintf("%s[%d]", title, index)
		partFilePath, err := utils.FilePath(partFileName, part.Ext, 0, options.DownloadPath, false)
		if err != nil {
			return err
		}
		parts[index] = partFilePath

		wgp.Add()

		go func(part *extractors.Part, fileName string) {
			defer wgp.Done()
			//	var err error
			//if downloader.option.MultiThread {
			//	err = downloader.multiThreadSave(part, data.URL, fileName)
			//} else {
			//	err = downloader.save(part, data.URL, fileName)
			//}
			err = save(part, data.URL, fileName, options)
			//err = save(part, data.URL, partFileName)
			if err != nil {
				lock.Lock()
				errs = append(errs, err)
				lock.Unlock()
			}
		}(part, partFileName)

	}

	wgp.Wait()

	if stream.Ext != "mp4" || stream.NeedMux {
		return utils.MergeFilesWithSameExtension(parts, mergedFilePath)
	}

	return utils.MergeToMP4(parts, mergedFilePath, title)
}

func GenSortedStreams(streams map[string]*extractors.Stream) []*extractors.Stream {
	sortedStreams := make([]*extractors.Stream, 0, len(streams))
	for _, data := range streams {
		sortedStreams = append(sortedStreams, data)
	}
	if len(sortedStreams) > 1 {
		sort.SliceStable(
			sortedStreams, func(i, j int) bool { return sortedStreams[i].Size > sortedStreams[j].Size },
		)
	}
	return sortedStreams
}

func GetStreamInfo(stream *extractors.Stream) StreamInfo {
	return StreamInfo{
		Quality: stream.Quality,
		Size:    fmt.Sprintf("%.2f MiB", float64(stream.Size)/(1024*1024)),
	}

}

func Caption(url, fileName, ext string, transform func([]byte) ([]byte, error), options DownloadOptions) error {
	body, err := request.GetByte(url, url, nil)
	if err != nil {
		return err
	}

	if transform != nil {
		body, err = transform(body)
		if err != nil {
			return err
		}
	}

	filePath, err := utils.FilePath(fileName, ext, 0, options.DownloadPath, true)
	if err != nil {
		return err
	}

	file, fileError := os.Create(filePath)
	if fileError != nil {
		return fileError
	}

	defer file.Close()

	if _, err = file.Write(body); err != nil {
		return err
	}

	return nil

}

func save(part *extractors.Part, refer, fileName string, options DownloadOptions) error {
	filePath, err := utils.FilePath(fileName, part.Ext, 0, options.DownloadPath, false)
	if err != nil {
		return err
	}
	fileSize, exists, err := utils.FileSize(filePath)
	if exists && fileSize == part.Size {
		return nil
	}

	tempFilePath := filePath + ".download"

	tempFileSize, _, err := utils.FileSize(tempFilePath)
	if err != nil {
		return err
	}
	headers := map[string]string{
		"Referer": refer,
	}
	var (
		file      *os.File
		fileError error
	)
	if tempFileSize > 0 {
		// range start from 0, 0-1023 means the first 1024 bytes of the file
		headers["Range"] = fmt.Sprintf("bytes=%d-", tempFileSize)
		file, fileError = os.OpenFile(tempFilePath, os.O_APPEND|os.O_WRONLY, 0644)
	} else {
		file, fileError = os.Create(tempFilePath)
	}

	if fileError != nil {
		return fileError
	}
	// close and rename temp file at the end of this function
	defer func() {
		// must close the file before rename or it will cause
		// `The process cannot access the file because it is being used by another process.` error.
		file.Close() // nolint
		if err == nil {
			os.Rename(tempFilePath, filePath) // nolint
		}
	}()

	temp := tempFileSize
	for i := 0; ; i++ {
		written, err := writeFile(part.URL, file, headers)
		if err == nil {
			break
		} else if i+1 >= defaultRetryTimes {
			return err
		}
		temp += written
		headers["Range"] = fmt.Sprintf("bytes=%d-", temp)
		time.Sleep(1 * time.Second)
	}

	return nil

}

func writeFile(url string, file *os.File, headers map[string]string) (int64, error) {
	res, err := request.Request(http.MethodGet, url, nil, headers)
	if err != nil {
		return 0, err
	}
	defer res.Body.Close() // nolint

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return 0, err
	}

	written, err := file.Write(body)
	if err != nil {
		return int64(written), err
	}
	return int64(written), nil
}

// DownloadList  下载列表
type DownloadList struct {
	mux sync.RWMutex

	// 维护一个下载中的列表
	Table map[string]struct{}
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
		mux:   sync.RWMutex{},
		Table: make(map[string]struct{}),
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
	dl.mux.RLocker()
	defer dl.mux.RUnlock()

	return len(dl.Table)
}
