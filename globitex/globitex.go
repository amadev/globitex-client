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

var timeout = time.Duration(getEnvAsInt("GLOBITEX_CLIENT_TIMEOUT", "10000")) * time.Second

type Client struct {
	host              string
	apiKey            string
	messageSecret     string
	transactionSecret string
}

type Response string

type Param struct {
	Key   string
	Value string
}

func NewClient() *Client {
	return &Client{
		host:              getEnv("GLOBITEX_CLIENT_HOST", "https://api.globitex.com"),
		apiKey:            getEnvOrFail("GLOBITEX_CLIENT_API_KEY"),
		messageSecret:     getEnvOrFail("GLOBITEX_CLIENT_MESSAGE_SECRET"),
		transactionSecret: getEnvOrFail("GLOBITEX_CLIENT_TRANSACTION_SECRET"),
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

func (c *Client) headerSignature(path string, params []Param, nonce string) string {
	message := fmt.Sprintf("%s&%s%s", c.apiKey, nonce, path)
	if len(params) != 0 {
		message = fmt.Sprintf("%s?%s", message, stringify(params, false))
	}
	mac := hmac.New(sha512.New, []byte(c.messageSecret))
	mac.Write([]byte(message))
	return hex.EncodeToString(mac.Sum(nil))
}

func (c *Client) transactionSignature(params []Param) string {
	message := stringify(params, false)
	mac := hmac.New(sha512.New, []byte(c.transactionSecret))
	mac.Write([]byte(message))
	return hex.EncodeToString(mac.Sum(nil))

}

func (c *Client) get(path string, params []Param) (Response, error) {
	client := &http.Client{
		Timeout: timeout}
	baseURL := fmt.Sprintf("%s%s", c.host, path)
	req, err := http.NewRequest("GET", baseURL, nil)
	if err != nil {
		return "", err
	}
	nonce := Nonce()
	req.Header.Add("X-API-Key", c.apiKey)
	req.Header.Add("X-Nonce", nonce)
	req.Header.Add("X-Signature", c.headerSignature(path, params, nonce))
	req.URL.RawQuery = stringify(params, true)
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return Response(body), nil
}

func (c *Client) post(path string, params []Param) (Response, error) {
	client := &http.Client{
		Timeout: timeout}
	params = append(params, Param{"transactionSignature", c.transactionSignature(params)})
	baseURL := fmt.Sprintf("%s%s", c.host, path)
	req, err := http.NewRequest("POST", baseURL, nil)
	if err != nil {
		return "", err
	}
	nonce := Nonce()
	req.Header.Add("X-API-Key", c.apiKey)
	req.Header.Add("X-Nonce", nonce)
	req.Header.Add("X-Signature", c.headerSignature(path, params, nonce))
	req.URL.RawQuery = stringify(params, true)
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return Response(body), nil
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
	path := "/api/1/eurowallet/status"
	return c.get(path, params)
}

func (c *Client) CreateNewPayment(params []Param) (Response, error) {
	path := "/api/1/eurowallet/payments"
	return c.post(path, params)
}
