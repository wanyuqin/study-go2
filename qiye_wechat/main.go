package main

import (
	"encoding/json"
	"fmt"
	"github.com/xuri/excelize/v2"
	"io"
	"log"
	"net/http"
	"time"
)

var (
	corpid = "wx3c4f70d4cb0d7969"
	// 打卡密钥
	checkIncorpsecret = "xGaIt2MU7cdHzd5QolW4kjFoFODsXfaz-Ra-ZybZEvw"
	// 通讯录密钥
	secret = "9y8ar4dHmgwC7TdKdXVV1wL4oCuI4DO1X-79A9SrFDE"

	startTime = "2023-04-01 00:00:00"
	endTime   = "2023-09-16 00:00:00"
)

func main() {
	accessToken, err := refreshToken(checkIncorpsecret)
	if err != nil {
		log.Fatal(err)
	}

	dayDayDataRequest := DayDataRequest{
		Starttime:  1688140800,
		Endtime:    1690819199,
		Useridlist: []string{"yuqin.wan@wowjoy.cn"},
	}
	dayDataResponse, err := getDayData(accessToken, dayDayDataRequest)
	if err != nil {
		log.Fatal(err)
	}

	hour, minute := getHourAndMinute(calRegularWorkSec(dayDataResponse))
	fmt.Printf("%d 小时 %d 分钟", hour, minute)

}

func refreshToken(secret string) (string, error) {
	url := fmt.Sprintf("https://qyapi.weixin.qq.com/cgi-bin/gettoken?corpid=%s&corpsecret=%s", corpid, secret)

	result := struct {
		Errcode     int    `json:"errcode"`
		Errmsg      string `json:"errmsg"`
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}{}
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	err = json.Unmarshal(data, &result)

	return result.AccessToken, nil

}

func getHourAndMinute(seconds int) (int, int) {
	duration := time.Duration(seconds) * time.Second
	hours := int(duration.Hours())
	minutes := int(duration.Minutes()) % 60

	return hours, minutes

}

type IamUser struct {
	Account  string
	Position string
	DeptId   string
}

func getUserInfo() ([]IamUser, error) {
	iamUsers := make([]IamUser, 0)

	excelFileName := "./userinfo.xlsx"
	f, err := excelize.OpenFile(excelFileName)
	if err != nil {
		return iamUsers, err
	}

	rows, err := f.GetRows("Sheet1")
	if err != nil {
		return iamUsers, err
	}
	for i := 1; i < len(rows); i++ {
		iamUsers = append(iamUsers, IamUser{
			Account:  rows[i][0],
			Position: rows[i][1],
			DeptId:   rows[i][2],
		})
	}
	return iamUsers, nil
}
