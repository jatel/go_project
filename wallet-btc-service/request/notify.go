package request

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/BlockABC/wallet-btc-service/common/config"
	"github.com/BlockABC/wallet-btc-service/common/log"
)

type NotifyAddress struct {
	Name       string `json:"name"`
	Language   string `json:"language"`
	Chain_type string `json:"chain_type"`
	Chain_id   string `json:"chain_id"`
	Cid        string `json:"cid"`
	Id         int32  `json:"id"`
	Platform   int    `json:"platform"`
}

func GetNotifyAddress() (error, []NotifyAddress) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", config.Cfg.BtcOpt.NotifyServerAddress+"/v1/cids", nil)
	if err != nil {
		return errors.New("http new request err:" + err.Error()), nil
	}

	body := url.Values{}
	body.Add("chain_type", "BTC")
	body.Add("chain_id", "mainnet")
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

	type addressData struct {
		Address []NotifyAddress `json:"addresses"`
	}

	type address struct {
		Errno  int32       `json:"errno"`
		Errmsg string      `json:"errmsg"`
		Data   addressData `json:"data"`
	}

	var rpcResult = address{}
	if err := json.Unmarshal(respBody, &rpcResult); err != nil {
		return errors.New("get notify address json unmarshal err:" + err.Error()), nil
	}

	if 0 != rpcResult.Errno || 0 != len(rpcResult.Errmsg) {
		return errors.New(rpcResult.Errmsg), nil
	}

	return nil, rpcResult.Data.Address
}

type NotifyInfo struct {
	Chain_type string `json:"chain_type"`
	Chain_id   string `json:"chain_id"`
	Msg_type   int    `json:"msg_type"`
	Cid        string `json:"cid"`
	Msg_id     string `json:"msg_id"`
	Language   string `json:"language"`
	Token_name string `json:"token_name"`
	Name       string `json:"name"`
	Platform   int    `json:"platform"`
}

type Push_list struct {
	List []NotifyInfo `json:"push_list"`
}

func NotifyTerminal(info Push_list) error {
	marshalledJSON, err := json.Marshal(info)
	if nil != err {
		return err
	}
	bodyReader := bytes.NewReader(marshalledJSON)
	httpRequest, err := http.NewRequest("POST", config.Cfg.BtcOpt.NotifyServerAddress+"/v1/push", bodyReader)
	if err != nil {
		return err
	}
	httpRequest.Close = true
	httpRequest.Header.Set("Content-Type", "application/json")

	httpClient := &http.Client{}
	httpResponse, err := httpClient.Do(httpRequest)
	if err != nil {
		return err
	}

	// Read the raw bytes and close the response.
	respBytes, err := ioutil.ReadAll(httpResponse.Body)
	httpResponse.Body.Close()
	if err != nil {
		err = fmt.Errorf("error reading json reply: %v", err)
		return err
	}

	// Handle unsuccessful HTTP responses
	if httpResponse.StatusCode < 200 || httpResponse.StatusCode >= 300 {
		if len(respBytes) == 0 {
			return fmt.Errorf("%d %s", httpResponse.StatusCode,
				http.StatusText(httpResponse.StatusCode))
		}
		return fmt.Errorf("%s", respBytes)
	}

	// Unmarshal the response
	type errInfo struct {
		Errno  int32  `json:"errno"`
		Errmsg string `json:"errmsg"`
	}

	var resp errInfo
	if err := json.Unmarshal(respBytes, &resp); err != nil {
		log.Log.Error("Unmarshal result fail, notify info:", string(marshalledJSON))
		return err
	}

	if 0 != resp.Errno {
		log.Log.Error("send notify fail, notify info:", string(marshalledJSON), "error info:", resp)
		return errors.New(resp.Errmsg)
	}

	log.Log.Info("send notify success, notify info:", string(marshalledJSON))
	return nil
}
