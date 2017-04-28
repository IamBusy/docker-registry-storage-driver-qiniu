package qiniu

import (
	"net/http"
	"encoding/json"
	"io"
	"strings"
	"qiniupkg.com/x/errors.v7"
)

func max(x, y int) int {
	if x > y {
		return x
	}
	return y
}

func min(x, y int) int  {
	if x > y {
		return y
	}
	return x
}

func request(method string, url string, bodyType string, body io.Reader, bodyLength int64 ) (map[string]string, error) {
	req, err := newRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", bodyType)
	req.ContentLength = int64(bodyLength)
	client := http.Client{}
	resp, err := client.Do(req)
	res := make([]byte,resp.ContentLength)
	resp.Body.Read(res)
	var resMap map[string]string
	err = json.Unmarshal(res,resMap)
	if err != nil {
		return nil, errors.New("json unmarshal error")
	}
	return resMap, nil
}


// --------------------------------------------------------------------

func newRequest(method, url1 string, body io.Reader) (req *http.Request, err error) {

	var host string

	// url1 = "-H <Host> http://<ip>[:<port>]/<path>"
	//
	if strings.HasPrefix(url1, "-H") {
		url2 := strings.TrimLeft(url1[2:], " \t")
		pos := strings.Index(url2, " ")
		if pos <= 0 {
			return nil, errors.New("invalid request url")
		}
		host = url2[:pos]
		url1 = strings.TrimLeft(url2[pos+1:], " \t")
	}

	req, err = http.NewRequest(method, url1, body)
	if err != nil {
		return
	}
	if host != "" {
		req.Host = host
	}
	return
}
