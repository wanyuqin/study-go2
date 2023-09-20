package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// StockList 股票列表
var StockList = "https://push2.eastmoney.com/api/qt/clist/get?&pz=150&pn=1&fs=b:BK1009&fields=f12,f14"

type SockListRequest struct {
	Path   string   `json:"url"`
	Pz     int      `json:"pz"`     // 页面大小
	Pn     int      `json:"pn"`     // 页码
	Fs     string   `json:"fs"`     // 股票分类
	Fields []string `json:"fields"` // 返回的列
}

func NewSockListRequest(pageSize, pageNum int, fs string, fields []string) (*SockListRequest, error) {
	request := &SockListRequest{
		Path:   "https://push2.eastmoney.com/api/qt/clist/get",
		Pz:     pageSize,
		Pn:     pageNum,
		Fields: fields,
		Fs:     fs,
	}

	u, err := url.Parse("https://push2.eastmoney.com/api/qt/clist/get")
	if err != nil {
		return request, err
	}

	if pageNum < 0 {
		pageNum = 0
	}

	if pageSize < 0 {
		pageSize = 50
	}
	query := url.Values{}
	query.Add("pz", strconv.Itoa(pageSize))
	query.Add("pn", strconv.Itoa(pageNum))
	query.Add("fs", fs)
	query.Add("fields", strings.Join(fields, ","))

	u.RawQuery = query.Encode()
	request.Path = u.String()

	return request, nil
}

func (s *SockListRequest) DoRequest() (*SockListResponse, error) {
	resp, err := http.Get(s.Path)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	response := SockListResponse{}

	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

type SockListResponse struct {
	Rc     int          `json:"rc"`
	Rt     int          `json:"rt"`
	Svr    int64        `json:"svr"`
	Lt     int          `json:"lt"`
	Full   int          `json:"full"`
	Dlmkts string       `json:"dlmkts"`
	Data   SockListData `json:"data"`
}

type SockListData struct {
	Total int                            `json:"total"`
	Diff  map[int]map[string]interface{} `json:"diff"`
}

type Sock struct {
	Code     string  `json:"code"`
	Name     string  `json:"name"`
	NewPrice float64 `json:"new_price"`
}

func (s *SockListResponse) SockList() []Sock {
	sockList := make([]Sock, 0, len(s.Data.Diff))
	if len(s.Data.Diff) <= 0 {
		return sockList
	}

	for k := range s.Data.Diff {
		value := s.Data.Diff[k]

		sock := Sock{
			Code:     value["f12"].(string),
			Name:     value["f14"].(string),
			NewPrice: value["f2"].(float64),
		}
		sockList = append(sockList, sock)

	}

	return sockList
}
