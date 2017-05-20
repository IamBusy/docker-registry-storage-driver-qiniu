package qiniu

import (
	"net/http"
	"encoding/json"
	"io"
	"strings"
	"qiniupkg.com/x/errors.v7"
	"fmt"
	"io/ioutil"
	"qiniupkg.com/api.v7/kodo"
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

func request(method string, url string, bodyType string, token string, body io.Reader, bodyLength int64 ) (map[string]interface{}, error) {
	req, err := newRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", bodyType)
	if token != "" {
		req.Header.Set("Authorization", token)
	}
	req.ContentLength = int64(bodyLength)
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err;
	}
	res := make([]byte,resp.ContentLength)
	res, err = ioutil.ReadAll(resp.Body)
	fmt.Println("request.responseBody=",string(res))
	var resMap map[string]interface{}
	err = json.Unmarshal(res,&resMap)
	if err != nil {
		fmt.Println("helper.request:",err)
		return nil, errors.New("json unmarshal error")
	}
	//rtn := map[string]string{}
	////var s string
	//for k, v := range resMap {
	//	rtn[k] = v.(string)
	//}
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

func newUptoken(p kodo.Bucket, key string) string {
	policy := &kodo.PutPolicy{
		Scope:   p.Name + ":" + key,
		Expires: 3600 * 24,
		UpHosts: p.UpHosts,
	}
	token := p.Conn.MakeUptoken(policy)
	fmt.Println("newUptoken:",token)
	return token
}
