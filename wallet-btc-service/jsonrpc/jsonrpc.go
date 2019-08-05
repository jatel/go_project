package jsonrpc

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/BlockABC/wallet-btc-service/common/config"
)

type Request struct {
	Jsonrpc string            `json:"jsonrpc"`
	Method  string            `json:"method"`
	Params  []json.RawMessage `json:"params"`
	ID      interface{}       `json:"id"`
}

// A specific type is used to help ensure the wrong errors aren't used.
type RPCErrorCode int

// RPCError represents an error that is used as a part of a JSON-RPC Response
// object.
type RPCError struct {
	Code    RPCErrorCode `json:"code,omitempty"`
	Message string       `json:"message,omitempty"`
}

type Response struct {
	Result json.RawMessage `json:"result"`
	Error  *RPCError       `json:"error"`
	ID     *interface{}    `json:"id"`
}

func isValidIDType(id interface{}) bool {
	switch id.(type) {
	case int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64,
		float32, float64,
		string,
		nil:
		return true
	default:
		return false
	}
}

func newRequest(id interface{}, method string, params []interface{}) (*Request, error) {
	if !isValidIDType(id) {
		str := fmt.Sprintf("the id of type '%T' is invalid", id)
		return nil, errors.New(str)
	}

	rawParams := make([]json.RawMessage, 0, len(params))
	for _, param := range params {
		marshalledParam, err := json.Marshal(param)
		if err != nil {
			return nil, err
		}
		rawMessage := json.RawMessage(marshalledParam)
		rawParams = append(rawParams, rawMessage)
	}

	return &Request{
		Jsonrpc: "1.0",
		ID:      id,
		Method:  method,
		Params:  rawParams,
	}, nil
}

func sendPostRequest(marshalledJSON []byte, rpcAddress string, rpcPort int, rpcUser, rpcPassword string) ([]byte, error) {
	// Generate a request to the configured RPC server.
	protocol := "http"
	rpcSercer := fmt.Sprintf("%s:%d", rpcAddress, rpcPort)
	url := protocol + "://" + rpcSercer
	bodyReader := bytes.NewReader(marshalledJSON)
	httpRequest, err := http.NewRequest("POST", url, bodyReader)
	if err != nil {
		return nil, err
	}
	httpRequest.Close = true
	httpRequest.Header.Set("Content-Type", "application/json")

	// Configure basic access authorization.
	httpRequest.SetBasicAuth(rpcUser, rpcPassword)

	// Create the new HTTP client that is configured according to the user-
	// specified options and submit the request.
	httpClient := &http.Client{}
	httpResponse, err := httpClient.Do(httpRequest)
	if err != nil {
		return nil, err
	}

	// Read the raw bytes and close the response.
	respBytes, err := ioutil.ReadAll(httpResponse.Body)
	httpResponse.Body.Close()
	if err != nil {
		err = fmt.Errorf("error reading json reply: %v", err)
		return nil, err
	}

	// Handle unsuccessful HTTP responses
	if httpResponse.StatusCode < 200 || httpResponse.StatusCode >= 300 {
		// Generate a standard error to return if the server body is
		// empty.  This should not happen very often, but it's better
		// than showing nothing in case the target server has a poor
		// implementation.
		if len(respBytes) == 0 {
			return nil, fmt.Errorf("%d %s", httpResponse.StatusCode,
				http.StatusText(httpResponse.StatusCode))
		}
		return nil, fmt.Errorf("%s", respBytes)
	}

	// Unmarshal the response.
	var resp Response
	if err := json.Unmarshal(respBytes, &resp); err != nil {
		return nil, err
	}

	if resp.Error != nil {
		return nil, fmt.Errorf("%s", resp.Error.Message)
	}
	return resp.Result, nil
}

func Call(id interface{}, method string, params []interface{}) ([]byte, error) {
	request, err := newRequest(id, method, params)
	if nil != err {
		return nil, err
	}

	jsonRequest, err := json.Marshal(request)
	if nil != err {
		return nil, err
	}

	resultByte, err := sendPostRequest(jsonRequest, config.Cfg.BtcOpt.RpcAddress, config.Cfg.BtcOpt.RpcPort, config.Cfg.BtcOpt.RpcUser, config.Cfg.BtcOpt.RpcPassword)
	if nil != err {
		return nil, err
	}

	return resultByte, err
}

func OmniCall(id interface{}, method string, params []interface{}) ([]byte, error) {
	request, err := newRequest(id, method, params)
	if nil != err {
		return nil, err
	}

	jsonRequest, err := json.Marshal(request)
	if nil != err {
		return nil, err
	}

	resultByte, err := sendPostRequest(jsonRequest, config.Cfg.OmniOpt.RpcAddress, config.Cfg.OmniOpt.RpcPort, config.Cfg.OmniOpt.RpcUser, config.Cfg.OmniOpt.RpcPassword)
	if nil != err {
		return nil, err
	}

	return resultByte, err
}
