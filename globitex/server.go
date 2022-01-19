package globitex

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type ResponseError struct {
	code     int
	message  string
	httpCode int
}

func (re *ResponseError) Error() string {
	return re.message
}

func queryToParams(q string) ([]Param, error) {
	params := make([]Param, 0)
	if len(q) == 0 {
		return params, nil
	}
	for _, s := range strings.Split(q, "&") {
		v := strings.Split(s, "=")
		if len(v) != 2 {
			return nil, errors.New("Fail to split params correctly")
		}
		k, err := url.QueryUnescape(v[0])
		if err != nil {
			return nil, errors.New("Cannot unescape key")
		}
		vl, err := url.QueryUnescape(v[1])
		if err != nil {
			return nil, errors.New("Cannot unescape value")
		}
		params = append(params,
			Param{
				Key:   k,
				Value: vl})
	}
	return params, nil
}

func ValidateRequest(req *http.Request) error {
	v := req.Header.Get("X-Signature")
	if v == "" {
		return &ResponseError{
			code:     30,
			message:  "Missing signature",
			httpCode: 403,
		}
	}
	v = req.Header.Get("X-API-Key")
	if v == "" {
		return &ResponseError{
			code:     10,
			message:  "Missing API key",
			httpCode: 403,
		}
	}
	v = req.Header.Get("X-Nonce")
	if v == "" {
		return &ResponseError{
			code:     10,
			message:  "Missing nonce",
			httpCode: 403,
		}
	}
	if req.Header.Get("X-API-Key") != getEnvOrFail("GLOBITEX_CLIENT_API_KEY") {
		return &ResponseError{
			code:     40,
			message:  "Invalid API key",
			httpCode: 403,
		}

	}
	_, err := strconv.Atoi(req.Header.Get("X-Nonce"))
	if err != nil {
		return &ResponseError{
			code:     60,
			message:  "Nonce is not valid",
			httpCode: 403,
		}
	}

	params := make([]Param, 0)
	if req.Method == "POST" {
		b, err := ioutil.ReadAll(req.Body)
		if err != nil {
			log.Println("Error reading request body")
		}
		defer req.Body.Close()
		params, err = queryToParams(string(b))
		if err != nil {
			log.Println("Error converting POST request to params")
		}
	} else {
		params, err = queryToParams(req.URL.RawQuery)
		if err != nil {
			log.Println("Error error converting GET request to params")
		}
	}

	if req.Header.Get("X-Signature") != HeaderSignature(
		getEnvOrFail("GLOBITEX_CLIENT_API_KEY"),
		getEnvOrFail("GLOBITEX_CLIENT_MESSAGE_SECRET"),
		req.URL.Path,
		req.Header.Get("X-Nonce"),
		params,
	) {
		return &ResponseError{
			code:     70,
			message:  "Wrong signature",
			httpCode: 403,
		}
	}
	si := -1
	for i, v := range params {
		if v.Key == "transactionSignature" {
			si = i
		}
	}
	if req.Method == "POST" {
		if si < 0 {
			return &ResponseError{
				code:     200,
				message:  "Mandatory parameter missing",
				httpCode: 400,
			}
		}
		if params[si].Value != TransactionSignature(getEnvOrFail("GLOBITEX_CLIENT_TRANSACTION_SECRET"), params) {
			return &ResponseError{
				code:     80,
				message:  "Invalid transactionSignature",
				httpCode: 400,
			}
		}
	}
	return nil
}

func Mux(handler http.Handler) http.Handler {
	urls := map[string]string{
		"/api/1/eurowallet/status":              "",
		"/api/1/eurowallet/deposit-details":     "",
		"/api/1/eurowallet/payments/history":    "",
		"/api/1/eurowallet/payments/commission": "",
		"/api/1/eurowallet/payments/status":     "",
		"/api/1/eurowallet/payments":            "",
	}
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		log.Printf("Got globitex handler request: %s", req.URL)
		if _, ok := urls[req.URL.Path]; ok {
			err := ValidateRequest(req)
			if err != nil {
				code := err.(*ResponseError).httpCode
				resp := fmt.Sprintf("{\"errors\": [{\"code\":%d,\"message\":\"%s\"}]}",
					err.(*ResponseError).code, err.(*ResponseError).message)
				log.Printf("Globitex handler response: (%d) %s", code, resp)
				w.WriteHeader(code)
				fmt.Fprintf(w, resp)
				return
			}
		}
		handler.ServeHTTP(w, req)
	})
}
