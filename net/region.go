package net

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
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
	resp, err := http.Get(urlString)
	if err != nil {
		return
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	var rs regionResp
	if err = json.Unmarshal(body, &rs); err != nil {
		return
	}
	cc = rs.CountryCode
	rc = rs.RegionName
	return
}
