package main

import (
	"changeme/backend/tools"
	"changeme/configs"
	"changeme/logger"
	"context"
	"fmt"
	"github.com/dustin/go-humanize"
	"github.com/iawia002/lux/extractors"
	"github.com/iawia002/lux/extractors/acfun"
	"github.com/iawia002/lux/extractors/bcy"
	"github.com/iawia002/lux/extractors/bilibili"
	"github.com/iawia002/lux/extractors/douyin"
	"github.com/iawia002/lux/extractors/douyu"
	"github.com/iawia002/lux/extractors/facebook"
	"github.com/iawia002/lux/extractors/twitter"
	"github.com/iawia002/lux/extractors/youtube"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"os"
	"path/filepath"
)

// App struct
type App struct {
	ctx context.Context
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	// 注册下载解释器
	extractorRegister()
	// 初始化日志
	logger.InitLogger()
	// 初始化下载列表
	tools.GetDownloadList()

}

// 关闭之前进行校验
func (a *App) beforeClose(ctx context.Context) bool {
	dl := tools.GetDownloadList()
	if dl.Length() > 0 {
		return false
	}
	return true
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

func (a *App) SelectDirectory() ([]NcmFile, error) {
	dialog, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{})
	if err != nil {
		fmt.Printf(err.Error())
		return nil, err
	}

	ncmList, err := FindNcmList(dialog)
	return ncmList, err
}

func (a *App) Transform(path []string) {
	if len(path) <= 0 {
		return
	}

	for _, p := range path {
		if isNcm(p) {
			go tools.ProcessNcmFile(p)
		}
	}
}

func (a *App) ExtractLink(link string) ([]tools.ExtractLinkData, error) {
	return tools.ExtractLink(link)
}

// Download 下载
func (a *App) Download(id string) error {
	return tools.Download(id)
}

// CancelDownload 取消下载
func (a *App) CancelDownload(id string) {

}

// GetDownloadSettings 获取下载设置
func (a *App) GetDownloadSettings() configs.DownloadConfig {
	// 加载配置文件
	config := configs.LoadConfig()
	return config.Download
}

func (a *App) SaveDownloadSettings(config configs.DownloadConfig) error {
	logger.Debug(fmt.Sprintf("%v", config))
	return configs.SaveDownloadSettings(config)
}

type NcmFile struct {
	Name    string `json:"name"`
	Path    string `json:"path"`
	ModTime string `json:"mod_time"`
	Size    string `json:"size"`
}

func FindNcmList(dirPath string) ([]NcmFile, error) {
	df, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, err
	}

	ncmFiles := make([]NcmFile, 0, 0)

	for _, file := range df {
		if !file.IsDir() && isNcm(file.Name()) {

			info, err := file.Info()
			if err != nil {
				logger.Error(err.Error())
				continue
			}

			ncmFile := NcmFile{
				Name:    file.Name(),
				Path:    filepath.Join(dirPath, file.Name()),
				ModTime: info.ModTime().Format("2006-01-02 15:04:05"),
				Size:    humanize.Bytes(uint64(info.Size())),
			}

			ncmFiles = append(ncmFiles, ncmFile)
		}
	}

	return ncmFiles, err

}

// 判断NCM
func isNcm(name string) bool {
	return filepath.Ext(name) == ".ncm"
}

// 加载下载器
func extractorRegister() {
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
}
