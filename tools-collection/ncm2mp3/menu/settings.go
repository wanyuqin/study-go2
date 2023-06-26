package menu

import (
	"github.com/wailsapp/wails/v2/pkg/menu"
	"runtime"
)

func SettingsMenu() *menu.Menu {
	appMenu := menu.NewMenu()

	settingsMenu := appMenu.AddSubmenu("偏好设置")
	settingsMenu.AddText("下载器设置", nil, downloadSettings)

	if runtime.GOOS == "darwin" {
		appMenu.Append(menu.EditMenu())
	}

	return appMenu

}

func downloadSettings(data *menu.CallbackData) {

}

func GetDownloadSettings() {

}
