package main

import (
	"encoding/json"
	"fmt"
	"github.com/amadev/globitex-client/globitex"
	"os"
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
			globitex.Param{"account", os.Getenv("GLOBITEX_TOOL_SRC_ACCOUNT")},
			globitex.Param{"amount", os.Getenv("GLOBITEX_TOOL_AMOUNT")},
			globitex.Param{"beneficiaryName", "Some beneficiary name"},
			globitex.Param{"beneficiaryAccount", os.Getenv("GLOBITEX_TOOL_DST_ACCOUNT")},
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
