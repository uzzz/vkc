package vkc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
)

const (
	MaximumExecuteBatchSize = 25

	baseUrl          = "https://api.vk.com/method"
	apiVersion       = "5.60"
	defaultUserAgent = "com.vk.vkclient/10 (unknown, iPhone OS 9.2, iPhone, Scale/2.000000)"
)

type VkClient struct {
	httpClient *http.Client
	Token      string
	UserAgent  string

	Users *usersService
	Audio *audioService
	Video *videoService
}

func NewVkClient(token string) *VkClient {
	client := &VkClient{
		httpClient: &http.Client{},
		Token:      token,
	}

	client.Users = &usersService{client: client}
	client.Audio = &audioService{client: client}
	client.Video = &videoService{client: client}

	return client
}

func (c *VkClient) NewRequest(method string, params map[string]string) (*http.Request, error) {
	requestUrl := fmt.Sprintf("%s/%s", baseUrl, method)

	data := url.Values{}
	data.Set("v", apiVersion)
	if c.Token != "" {
		data.Set("access_token", c.Token)
	}
	for k, v := range params {
		data.Set(k, v)
	}

	// TODO: remove buf
	var buf io.ReadWriter
	req, err := http.NewRequest("GET", requestUrl+"?"+data.Encode(), buf)
	if err != nil {
		return nil, err
	}

	var userAgent string
	if c.UserAgent != "" {
		userAgent = c.UserAgent
	} else {
		userAgent = defaultUserAgent
	}
	req.Header.Set("User-Agent", userAgent)

	return req, nil
}

type VkResponse struct {
	Response *json.RawMessage `json:"response"`
}

func (c *VkClient) Do(req *http.Request, v interface{}) (*http.Response, error) {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer func() {
		// Drain up to 512 bytes and close the body to let the Transport reuse the connection
		io.CopyN(ioutil.Discard, resp.Body, 512)
		resp.Body.Close()
	}()

	vkResponse, err := parseVkResponse(resp)
	if err != nil {
		return resp, err
	}

	// fmt.Println(req.URL)
	// responseData, _ := ioutil.ReadAll(resp.Body)
	// fmt.Println(string(responseData))

	if v != nil {
		if w, ok := v.(io.Writer); ok {
			io.Copy(w, bytes.NewBuffer(*vkResponse))
		} else {
			err = json.Unmarshal(*vkResponse, v)
		}
	}

	return resp, err
}

type ErrorResponse struct {
	Response *http.Response // HTTP response that caused this error
}

func (r *ErrorResponse) Error() string {
	return fmt.Sprintf("%v %v: %d",
		r.Response.Request.Method, r.Response.Request.URL, r.Response.StatusCode)
}

type vkErrorRequestParam struct {
	Key   string
	Value string
}

type vkResponseEnvelope struct {
	Error    *VkError         `json:"error"`
	Response *json.RawMessage `json:"response"`
}

func parseVkResponse(r *http.Response) (*json.RawMessage, error) {
	if c := r.StatusCode; c < 200 && c > 299 {
		return nil, &ErrorResponse{Response: r}
	}

	envelope := new(vkResponseEnvelope)
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	if data != nil {
		err = json.Unmarshal(data, envelope)
		if err != nil {
			return nil, err
		}

		if envelope.Error != nil {
			return nil, castError(envelope.Error)
		}
		return envelope.Response, nil
	}

	return nil, nil
}

func (c *VkClient) Execute(code string, v interface{}) error {
	params := map[string]string{
		"code": code,
	}
	req, err := c.NewRequest("execute", params)
	if err != nil {
		return err
	}

	_, err = c.Do(req, v)
	return err
}
