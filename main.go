package main

import (
	"encoding/json"
	"fmt"
	"github.com/amadev/globitex-client/globitex"
)

type Status struct {
	PaymentId string `json:"paymentId"`
	Status    string `json:"status"`
}

func main() {

	client := globitex.NewClient()

	nonce := globitex.Nonce()
	resp, err := client.CreateNewPayment(
		[]globitex.Param{
			globitex.Param{"requestTime", nonce},
			globitex.Param{"account", "LT833080020000001060"},
			globitex.Param{"amount", "1.00"},
			globitex.Param{"beneficiaryName", "Some beneficiary name"},
			globitex.Param{"beneficiaryAccount", "LT983250082405215248"},
			globitex.Param{"beneficiaryReference", "Some reference text"},
		})
	if err != nil {
		panic(err)
	}
	pr := Status{}
	err = json.Unmarshal([]byte(resp.Body), &pr)
	if err != nil {
		panic(err)
	}
	fmt.Println("CreateNewPayment ", resp)

	resp, err = client.GetPaymentStatus(
		[]globitex.Param{
			globitex.Param{"clientPaymentId", pr.PaymentId},
		})
	if err != nil {
		panic(err)
	}
	fmt.Println("GetPaymentStatus ", resp.Body)
}
