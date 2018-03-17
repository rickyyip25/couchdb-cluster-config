package cluster_config

import (
	"crypto/tls"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"io"
)

var (
	insecure = flag.Bool("insecure", true, "Ignore server certificate if using https")
)

type BasicAuth struct {
	Username string
	Password string
}

type CouchdbClient struct {
	BaseUri   string
	basicAuth BasicAuth
	databases []string
	client    *http.Client
}

func (c *CouchdbClient) Request(method string, uri string, body io.Reader) (respData []byte, err error) {
	req, err := http.NewRequest(method, uri, body)
	if err != nil {
		return nil, err
	}
	if body != nil {
		req.Header = http.Header{
			"Content-Type": []string{"application/json"},
		}
	}
	if len(c.basicAuth.Username) > 0 {
		req.SetBasicAuth(c.basicAuth.Username, c.basicAuth.Password)
	}

	fmt.Printf("[REQ] %v\n", req)
	resp, err := c.client.Do(req)
	if err != nil {
		fmt.Printf("[RES-ERR] %v\n", err)
		return nil, err
	}
	fmt.Printf("[RES-OK] %v\n", resp)

	respData, err = ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		if err != nil {
			respData = []byte(err.Error())
		}
		return nil, fmt.Errorf("status %s (%d): %s", resp.Status, resp.StatusCode, respData)
	}

	return respData, nil
}

func NewCouchdbClient(uri string, basicAuth BasicAuth, databases []string) *CouchdbClient {
	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: *insecure},
		},
	}

	return &CouchdbClient{
		BaseUri:   uri,
		basicAuth: basicAuth,
		databases: databases,
		client:    httpClient,
	}
}
