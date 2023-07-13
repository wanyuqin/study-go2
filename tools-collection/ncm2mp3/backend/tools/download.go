package tools

import (
	"changeme/configs"
	"changeme/logger"
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/wanyuqin/lux/extractors"
	"github.com/wanyuqin/lux/request"
	"github.com/wanyuqin/lux/utils"
	"io"
	"net/http"
	"os"
	"sort"
	"sync"
	"time"
)

var defaultThreadNumber = 10
var defaultRetryTimes = 3
var FileExistErr = errors.New("file already exists")

var LinkDataMap map[string]*extractors.Data

type ExtractLinkData struct {
	Id         string  `json:"id"`
	Title      string  `json:"title"`
	Type       string  `json:"type"`
	Url        string  `json:"url"`
	Quality    string  `json:"quality"`
	Size       string  `json:"size"`
	Byte       int64   `json:"byte"`
	Percentage float64 `json:"percentage"` // 百分比
}

type StreamInfo struct {
	Quality string `json:"quality"`
	Size    string `json:"size"`
	Byte    int64  `json:"byte"`
}

func init() {
	LinkDataMap = make(map[string]*extractors.Data)
}

// ExtractLink 解析地址网页内容
func ExtractLink(link string) ([]ExtractLinkData, error) {
	logger.Debug(fmt.Sprintf("extract link %s", link))
	data, err := extractors.Extract(link, extractors.Options{
		Playlist: false,
		Items:    "",
	})

	if err != nil {
		logger.Error(err.Error())
		return nil, err
	}

	elds := make([]ExtractLinkData, 0, len(data))
	for i, item := range data {
		uid, err := uuid.NewUUID()
		if err != nil {
			fmt.Printf(err.Error())
			continue
		}

		eld := ExtractLinkData{
			Title: item.Title,
			Url:   item.URL,
			Type:  string(item.Type),
			Id:    uid.String(),
		}

		sortStreams := GenSortedStreams(item.Streams)
		if len(sortStreams) > 0 {
			streamName := sortStreams[0].ID
			stream, ok := item.Streams[streamName]
			if !ok {
				continue

			}
			streamInfo := GetStreamInfo(stream)
			eld.Size = streamInfo.Size
			eld.Quality = streamInfo.Quality
			eld.Byte = streamInfo.Byte
		}

		LinkDataMap[uid.String()] = data[i]

		elds = append(elds, eld)
	}

	return elds, err
}

func Download(ctx context.Context, eld ExtractLinkData) error {
	data, ok := LinkDataMap[eld.Id]
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

	options := &DownloadOptions{
		Ctx:          ctx,
		Eld:          eld,
		Data:         data,
		DownloadPath: config.Download.Path,
		mux:          sync.RWMutex{},
		doneByte:     0,
	}
	// 开始下载 加入下载队列
	GetDownloadList().Push(eld.Id)
	err = download(options)
	return err
}

func download(options *DownloadOptions) error {
	data := options.Data
	if len(data.Streams) == 0 {
		return errors.New(fmt.Sprintf("no streams in title %s", data.Title))
	}

	sortStreams := GenSortedStreams(data.Streams)

	title := data.Title

	streamName := sortStreams[0].ID
	//stream 具体文件内容流
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
		return FileExistErr
	}

	wgp := utils.NewWaitGroupPool(defaultThreadNumber)
	errs := make([]error, 0)
	lock := sync.Mutex{}
	parts := make([]string, len(stream.Parts))

	// 每一个下载任务都要有一个ctx，用来控制goroutine的终止
	ctx, cancel := context.WithCancel(context.Background())
	GetDownloadPool().Add(options.Eld.Id, cancel)

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
		// 去下载每个part
		go func(ctx context.Context, part *extractors.Part, fileName string, options *DownloadOptions) {
			defer wgp.Done()
			//	var err error
			//if downloader.option.MultiThread {
			//	err = downloader.multiThreadSave(part, data.URL, fileName)
			//} else {
			//	err = downloader.save(part, data.URL, fileName)
			//}
			// 文件保存
			err = save(ctx, part, data.URL, fileName, options)
			//err = save(part, data.URL, partFileName)
			if err != nil {
				lock.Lock()
				errs = append(errs, err)
				lock.Unlock()
			}

			//// 下载完成之后计算百分比
			//options.CalculatePercent()

		}(ctx, part, partFileName, options)

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
		Byte:    stream.Size,
	}

}

func Caption(url, fileName, ext string, transform func([]byte) ([]byte, error), options *DownloadOptions) error {
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

func save(ctx context.Context, part *extractors.Part, refer, fileName string, options *DownloadOptions) error {

	select {
	case <-ctx.Done():
		logger.Debug("cancel download")
		return nil
	default:
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
			file.Close() // nolint
			if err == nil {
				os.Rename(tempFilePath, filePath) // nolint
			}
		}()
		// 下载内容大小
		temp := tempFileSize
		for i := 0; ; i++ {
			written, err := writeFile(part.URL, file, headers)
			if err == nil {
				options.AddDoneByte(written)
				options.CalculatePercent()
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

func CancelDownload(id string) {
	// 取消下载任务
	GetDownloadPool().Cancel(id)
	// 删除临时文件
	GetDownloadList().ClearTempFile(id)
}
