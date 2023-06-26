package configs

import (
	"changeme/logger"
	"embed"
	"errors"
	"fmt"
	"github.com/duke-git/lancet/v2/strutil"
	"gopkg.in/yaml.v2"
	"os"
)

//go:embed settings.yaml
var Settings embed.FS

var (
	DownloadPathBlankErr  = errors.New("download path is blank")
	DownloadPathNotDirErr = errors.New("download path is not dir")
)

type Config struct {
	Download DownloadConfig `json:"download" yaml:"download"`
}

type DownloadConfig struct {
	Path string `json:"path" yaml:"path"`
}

func GetConfig() Config {
	return LoadConfig()
}

func LoadConfig() Config {
	config := Config{}

	body, err := Settings.ReadFile("settings.yaml")
	if err != nil {
		logger.Error(err.Error())
		return config
	}

	err = yaml.Unmarshal(body, &config)
	if err != nil {
		logger.Error(err.Error())
		return config
	}
	logger.Debug(fmt.Sprintf("%v\n", config))
	return config
}

func SaveDownloadSettings(downloadConfig DownloadConfig) error {
	cfg := LoadConfig()
	cfg.Download = downloadConfig

	body, err := yaml.Marshal(cfg)
	if err != nil {
		logger.Error(err.Error())
		return err
	}
	file, err := os.OpenFile("./settings.yaml", os.O_TRUNC|os.O_WRONLY, 0666)
	defer file.Close()

	if err != nil {
		logger.Error(err.Error())
		return err
	}

	_, err = file.Write(body)
	if err != nil {
		logger.Error(err.Error())
		return err
	}

	return nil

}

func (c *Config) CheckDownloadPath() error {
	if strutil.IsBlank(c.Download.Path) {
		return DownloadPathBlankErr
	}

	stat, err := os.Stat(c.Download.Path)
	if err != nil {
		return err
	}

	if !stat.IsDir() {
		return DownloadPathNotDirErr
	}
	return nil
}
