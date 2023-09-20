package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

var (
	dayURL = "https://qyapi.weixin.qq.com/cgi-bin/checkin/getcheckin_daydata?access_token="
)

type DayDataRequest struct {
	Starttime  int      `json:"starttime"`
	Endtime    int      `json:"endtime"`
	Useridlist []string `json:"useridlist"`
}

type DayDataResponse struct {
	Errcode int    `json:"errcode"`
	Errmsg  string `json:"errmsg"`
	Datas   []struct {
		BaseInfo struct {
			Date        int    `json:"date"`
			RecordType  int    `json:"record_type"`
			Name        string `json:"name"`
			NameEx      string `json:"name_ex"`
			DepartsName string `json:"departs_name"`
			Acctid      string `json:"acctid"`
			RuleInfo    struct {
				Groupid      int    `json:"groupid"`
				Groupname    string `json:"groupname"`
				Scheduleid   int    `json:"scheduleid"`
				Schedulename string `json:"schedulename,omitempty"`
				Checkintime  []struct {
					WorkSec    int `json:"work_sec"`
					OffWorkSec int `json:"off_work_sec"`
				} `json:"checkintime"`
			} `json:"rule_info"`
			DayType int `json:"day_type"`
		} `json:"base_info"`
		SummaryInfo struct {
			CheckinCount    int `json:"checkin_count"`
			RegularWorkSec  int `json:"regular_work_sec"`
			StandardWorkSec int `json:"standard_work_sec"`
			EarliestTime    int `json:"earliest_time"`
			LastestTime     int `json:"lastest_time"`
		} `json:"summary_info"`
		HolidayInfos []struct {
			SpNumber string `json:"sp_number"`
			SpTitle  struct {
				Data []struct {
					Text string `json:"text"`
					Lang string `json:"lang"`
				} `json:"data"`
			} `json:"sp_title"`
			SpDescription struct {
				Data []struct {
					Text string `json:"text"`
					Lang string `json:"lang"`
				} `json:"data"`
			} `json:"sp_description"`
		} `json:"holiday_infos"`
		ExceptionInfos []struct {
			Exception int `json:"exception"`
			Count     int `json:"count"`
			Duration  int `json:"duration"`
		} `json:"exception_infos"`
		OtInfo struct {
			OtStatus          int           `json:"ot_status"`
			OtDuration        int           `json:"ot_duration"`
			ExceptionDuration []interface{} `json:"exception_duration"`
		} `json:"ot_info"`
		SpItems []struct {
			Type       int    `json:"type"`
			VacationId int    `json:"vacation_id"`
			Count      int    `json:"count"`
			Duration   int    `json:"duration"`
			TimeType   int    `json:"time_type"`
			Name       string `json:"name"`
		} `json:"sp_items"`
	} `json:"datas"`
}

func getDayData(accessToken string, param DayDataRequest) (DayDataResponse, error) {
	dayURL = fmt.Sprintf("%s%s", dayURL, accessToken)
	result := DayDataResponse{}
	requestData, err := json.Marshal(param)
	if err != nil {
		return result, err
	}

	resp, err := http.Post(dayURL, "application/json", bytes.NewBuffer(requestData))
	if err != nil {
		return result, err
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return result, err
	}

	err = json.Unmarshal(data, &result)

	return result, err

}

func calRegularWorkSec(dayData DayDataResponse) int {
	var (
		totalWorkSpec int
		totalWorkDay  int
	)
	for i := range dayData.Datas {
		data := dayData.Datas[i]
		if data.SummaryInfo.RegularWorkSec > 0 {
			totalWorkSpec += data.SummaryInfo.RegularWorkSec
			totalWorkDay += 1
		}
	}

	averageSec := totalWorkSpec / totalWorkDay
	return averageSec
}
