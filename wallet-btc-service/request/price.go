package request

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

const URLPRICE = `https://api.coingecko.com/api/v3/simple/price`

func GetPrice(ids []string) (errInfo error, priceInfo map[string]float64) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", URLPRICE, nil)
	if err != nil {
		return errors.New("http new request err:" + err.Error()), nil
	}

	body := url.Values{}
	body.Add("ids", strings.Join(ids, ","))
	body.Add("vs_currencies", "usd")
	req.Header.Set("accept", "application/json")
	req.URL.RawQuery = body.Encode()

	resp, err := client.Do(req)
	if err != nil {
		return errors.New("client do err:" + err.Error()), nil
	}
	if resp.StatusCode != http.StatusOK {
		return errors.New("request error:" + resp.Status), nil
	}
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return errors.New("read body err:" + err.Error()), nil
	}

	type price struct {
		Price float64 `json:"usd"`
	}

	var rpcResult = map[string]price{}
	if err := json.Unmarshal(respBody, &rpcResult); err != nil {
		return errors.New("coin brief json unmarshal err:" + err.Error()), nil
	}

	var result = map[string]float64{}
	for k, v := range rpcResult {
		result[k] = v.Price
	}
	return nil, result
}
