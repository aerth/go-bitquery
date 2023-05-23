// Copyright © 2023 aerth
// Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the “Software”), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:
// The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.
// THE SOFTWARE IS PROVIDED “AS IS”, WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

// package bitquery paradigm
package bitquery

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)


const (
	Endpoint1 = "https://graphql.bitquery.io"
	Endpoint2 = "https://streaming.bitquery.io/graphql"
)

// BitqueryModel interface easy to implement
type BitqueryModel interface {
	ToMap() map[string]any // input vars to map
	Query() string // the query string
	Endpoint() string // endpoint ("https://streaming.bitquery.io/graphql" or "https://graphql.bitquery.io")
}

var HttpClient *http.Client = &http.Client{Timeout: time.Second * 40}

// Do gets json bytes
func Do(apikey string, x BitqueryModel) ([]byte, error) {
	q, err := json.Marshal(map[string]interface{}{
		"query":     x.Query(),
		"variables": x.ToMap(),
	})
	if err != nil {
		return nil, err
	}
	request, err := http.NewRequest(http.MethodPost, x.Endpoint(), bytes.NewBuffer(q))
	if err != nil {
		return nil, fmt.Errorf("bad request: %v", err)
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-API-KEY", apikey)
	
	response, err := HttpClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("the HTTP request failed with error %v", err)
	}
	defer response.Body.Close()

	data, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// Unmarshal send request and decode response to pointer ptr
func Unmarshal(apikey string, x BitqueryModel, ptr any) error {
	resp, err := Do(apikey, x)
	if err != nil {
		return err
	}
	var wrapped Wrapper
	err = json.Unmarshal(resp, &wrapped)
	if err != nil {
		if err != nil && strings.Contains(err.Error(), "invalid character '<'") {
			return fmt.Errorf("resp: %v", string(resp))
		}
		return json.Unmarshal(resp, ptr)
	}
	if len(wrapped.Errors) != 0 {
		return fmt.Errorf("query error: %v", wrapped.Errors)
	}
	err = json.Unmarshal(wrapped.Data, ptr)
	if err != nil {
		return err
	}
	return nil
}

// Wrapper to extract 'data' field
type Wrapper struct {
	Data json.RawMessage `json:"data,omitempty"`
	Errors Errors `json:"errors,omitempty"`
}

// Errors
//
// example response:
//
// {"data":null,"errors":[
// {"message":"Variable \"$network\" cannot be non-input type \"EthereumNetwork!\".", "locations":[{"line":2,"column":18}]},
// {"message":"Unknown type \"EthereumNetwork\".", "locations":[{"line":2,"column":18}]},
// {"message":"Cannot query field \"ethereum\" on type \"RootQuery\".", "locations":[{"line":3,"column":2}]}
// ]}
type Errors []ErrorM

func (e ErrorM) String() string {
	return e.Message
}

type ErrorMsg struct {
	Errors Errors `json:"errors"`
}
type Locations struct {
	Line   int `json:"line"`
	Column int `json:"column"`
}
type ErrorM struct {
	Message   string      `json:"message"`
	Locations []Locations `json:"locations"`
}
