package main

import (
	"fmt"
	"github.com/amadev/globitex-client/globitex"
	"net/http"
)

func status() {
}

func main() {

	mux := http.NewServeMux()
	mux.HandleFunc("/api/1/eurowallet/payments", func(w http.ResponseWriter, req *http.Request) {
		w.Write([]byte("ok"))
	})
	go func() {
		http.ListenAndServe(":8090", globitex.Mux(mux))
	}()

	client := globitex.NewClient()

	resp, err := client.GetAccountStatus()
	if err != nil {
		panic(err)
	}
	fmt.Println("GetAccountStatus ", resp)

	resp, err = client.GetDepositDetails()
	if err != nil {
		panic(err)
	}
	fmt.Println("GetDepositDetails ", resp)

	resp, err = client.GetPaymentHistory()
	if err != nil {
		panic(err)
	}
	fmt.Println("GetPaymentHistory ", resp)

	resp, err = client.GetPaymentCommissionAmount(
		[]globitex.Param{
			globitex.Param{"beneficiaryBankAccount", "LT983250082405215248"},
			globitex.Param{"amount", "100"},
		})
	if err != nil {
		panic(err)
	}
	fmt.Println("GetPaymentCommissionAmount ", resp)

	resp, err = client.GetPaymentStatus(
		[]globitex.Param{
			globitex.Param{"clientPaymentId", "100916"},
		})
	if err != nil {
		panic(err)
	}
	fmt.Println("GetPaymentStatus ", resp)

	nonce := globitex.Nonce()
	resp, err = client.CreateNewPayment(
		[]globitex.Param{
			globitex.Param{"requestTime", nonce},
			globitex.Param{"account", "LT833080020000001060"},
			globitex.Param{"amount", "1.00"},
			globitex.Param{"beneficiaryName", "Some beneficiary name"},
			globitex.Param{"beneficiaryAccount", "LT983250082405215248"},
			globitex.Param{"beneficiaryReference", "Some reference text"},
			globitex.Param{"externalPaymentId", "TR-" + nonce},
		})
	if err != nil {
		panic(err)
	}
	fmt.Println("GetPaymentStatus ", resp)
}
