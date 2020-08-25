package net

import (
	"encoding/json"
	"sync"

	"github.com/valyala/fasthttp"
)

var (
	_client *fasthttp.Client
	once    sync.Once
)

type regionResp struct {
	Status      string `json:"status"`
	Country     string `json:"country"`
	CountryCode string `json:"countryCode"`
	Region      string `json:"region"`
	RegionName  string `json:"regionName"`
	City        string `json:"city"`
	Timezone    string `json:"timezone"`
}

func RegionCode(ip string) (cc, rc string, err error) {
	urlString := "http://ip-api.com/json/" + ip
	req := fasthttp.AcquireRequest()
	req.SetRequestURI(urlString)
	resp := fasthttp.AcquireResponse()
	defer func() {
		fasthttp.ReleaseResponse(resp)
		fasthttp.ReleaseRequest(req)
	}()

	if err = client().Do(req, resp); err != nil {
		return
	}
	var rs regionResp
	if err = json.Unmarshal(resp.Body(), &rs); err != nil {
		return
	}
	cc = rs.CountryCode
	rc = rs.RegionName
	return
}

func client() *fasthttp.Client {
	once.Do(func() {
		_client = &fasthttp.Client{}
	})
	return _client
}
