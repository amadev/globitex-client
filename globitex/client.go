package globitex

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

func getEnvOrFail(key string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	log.Fatalf("Environment variable '%s' has to be specified\n", key)
	return ""
}

func getEnv(key string, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}

	return defaultVal
}

func getEnvAsInt(key string, defaultVal string) int {
	v, err := strconv.Atoi(getEnv(key, defaultVal))
	if err != nil {
		panic(err)
	}
	return v
}

var timeout = time.Duration(getEnvAsInt("GLOBITEX_TOOL_CONNECT_TIMEOUT", "10000")) * time.Second

type Client struct {
	host              string
	apiKey            string
	messageSecret     string
	transactionSecret string
}

type Response struct {
	Body string
	Code int
}

type Param struct {
	Key   string
	Value string
}

func NewClient() *Client {
	return &Client{
		host:              getEnv("GLOBITEX_TOOL_HOST", "https://api.globitex.com"),
		apiKey:            getEnvOrFail("GLOBITEX_TOOL_API_KEY"),
		messageSecret:     getEnvOrFail("GLOBITEX_TOOL_MESSAGE_SECRET"),
		transactionSecret: getEnvOrFail("GLOBITEX_TOOL_TRANSACTION_SECRET"),
	}
}

func Nonce() string {
	return strconv.Itoa(int(time.Now().UnixNano() / int64(time.Millisecond)))
}

func stringify(params []Param, urlescape bool) string {
	var buf bytes.Buffer
	for _, p := range params {
		k := p.Key
		v := p.Value
		if urlescape {
			k = url.QueryEscape(k)
			v = url.QueryEscape(v)
		}
		buf.WriteString(fmt.Sprintf("%s=%s&", k, v))
	}
	s := buf.String()
	if len(s) == 0 {
		return s
	}
	return s[:len(s)-1]

}

func HeaderSignature(apiKey, secret, path, nonce string, params []Param) string {
	message := fmt.Sprintf("%s&%s%s", apiKey, nonce, path)
	if len(params) != 0 {
		message = fmt.Sprintf("%s?%s", message, stringify(params, false))
	}
	mac := hmac.New(sha512.New, []byte(secret))
	mac.Write([]byte(message))
	log.Printf("HeaderSignature %s for message %s", hex.EncodeToString(mac.Sum(nil)), message)
	return hex.EncodeToString(mac.Sum(nil))
}

func TransactionSignature(secret string, params []Param) string {
	filtered := make([]Param, 0)
	for _, v := range params {
		if v.Key == "transactionSignature" {
			continue
		}
		filtered = append(filtered, v)
	}
	message := stringify(filtered, false)
	mac := hmac.New(sha512.New, []byte(secret))
	mac.Write([]byte(message))
	return hex.EncodeToString(mac.Sum(nil))
}

func (c *Client) headerSignature(path, nonce string, params []Param) string {
	return HeaderSignature(c.apiKey, c.messageSecret, path, nonce, params)
}

func (c *Client) transactionSignature(params []Param) string {
	return TransactionSignature(c.transactionSecret, params)
}

func (c *Client) get(path string, params []Param) (Response, error) {
	client := &http.Client{
		Timeout: timeout}
	baseURL := fmt.Sprintf("%s%s", c.host, path)
	req, err := http.NewRequest("GET", baseURL, nil)
	if err != nil {
		return Response{}, err
	}
	nonce := Nonce()
	_ = nonce
	req.Header.Add("X-API-Key", c.apiKey)
	req.Header.Add("X-Nonce", nonce)
	req.Header.Add("X-Signature", c.headerSignature(
		strings.TrimPrefix(path, getEnv("GLOBITEX_TOOL_URL_PREFIX", "")), nonce, params))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.URL.RawQuery = stringify(params, true)
	resp, err := client.Do(req)
	if err != nil {
		return Response{}, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return Response{}, err
	}
	return Response{Body: string(body), Code: resp.StatusCode}, nil
}

func (c *Client) post(path string, params []Param) (Response, error) {
	client := &http.Client{
		Timeout: timeout}
	params = append(params, Param{"transactionSignature", c.transactionSignature(params)})
	baseURL := fmt.Sprintf("%s%s", c.host, path)
	req, err := http.NewRequest("POST", baseURL, strings.NewReader(stringify(params, true)))
	if err != nil {
		return Response{}, err
	}
	nonce := Nonce()
	req.Header.Add("X-API-Key", c.apiKey)
	req.Header.Add("X-Nonce", nonce)
	req.Header.Add("X-Signature", c.headerSignature(
		strings.TrimPrefix(path, getEnv("GLOBITEX_TOOL_URL_PREFIX", "")), nonce, params))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	for k, v := range req.Header {
		fmt.Printf("request header %s %s\n", k, v)
	}
	for _, v := range params {
		fmt.Printf("request param %s %s\n", v.Key, v.Value)
	}

	resp, err := client.Do(req)
	if err != nil {
		return Response{}, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return Response{}, err
	}
	return Response{Body: string(body), Code: resp.StatusCode}, nil
}

func (c *Client) GetAccountStatus() (Response, error) {
	path := "/api/1/eurowallet/status"
	params := make([]Param, 0)
	return c.get(path, params)
}

func (c *Client) GetDepositDetails() (Response, error) {
	path := "/api/1/eurowallet/deposit-details"
	params := make([]Param, 0)
	return c.get(path, params)
}

func (c *Client) GetPaymentHistory() (Response, error) {
	path := "/api/1/eurowallet/payments/history"
	params := make([]Param, 0)
	return c.get(path, params)
}

func (c *Client) GetPaymentCommissionAmount(params []Param) (Response, error) {
	path := "/api/1/eurowallet/payments/commission"
	return c.get(path, params)
}

func (c *Client) GetPaymentStatus(params []Param) (Response, error) {
	path := getEnv("GLOBITEX_TOOL_URL_PREFIX", "") + "/api/1/eurowallet/payments/status"
	return c.get(path, params)
}

func (c *Client) CreateNewPayment(params []Param) (Response, error) {
	path := getEnv("GLOBITEX_TOOL_URL_PREFIX", "") + "/api/1/eurowallet/payments"
	return c.post(path, params)
}
