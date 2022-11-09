package main

import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
)

var SECRETS = map[string]string{
	"TransactionSigningSecretKey": os.Getenv("TRANSACTION_SECRET_KEY"),
	"MessageSigningSecretKey":     os.Getenv("MESSAGE_SECRET_KEY"),
	"APIKey":                      os.Getenv("API_KEY"),
}

func main() {
	makePayment()
}

func getBalances() {

}

func makePayment() {
	nonce := time.Now().UnixNano() / 1e6

	payment := CreateNewPaymentRequest{
		RequestTime:          nonce,
		Amount:               "1",
		Account:              "LT543080020000000224",
		BeneficiaryReference: fmt.Sprintf("Testing payment"),
		BeneficiaryName:      "UAB Decentralized",
		BeneficiaryAccount:   "LT593910020000000053",
		BeneficiaryAddress:   "A. Goštauto g. 8-340, LT-01108 Vilnius, LT",
	}

	message := payment.createSignatureMessage()
	payment.TransactionSignature = strings.ToLower(GenerateHMACSHA512([]byte(message), []byte(SECRETS["TransactionSigningSecretKey"]), &HMACSHA512Options{}))

	fmt.Println(payment.TransactionSignature)
	fmt.Println(payment)

	path := "/api/1/eurowallet/payments"
	headers := createAuthHeaders(path, message, nonce)

	response, err := resty.New().R().
		SetBody(payment).
		SetHeaders(headers).
		SetResult(CreateNewPaymentResponse{}).
		SetError(ErrorResult{}).
		Post(fmt.Sprintf("%s%s", "https://api.globitex.com", path))

	if err != nil {
		fmt.Printf("failed to make a request: %v", err)
		return
	}

	if response.StatusCode() != 200 {
		errResponse := response.Error().(*ErrorResult)
		fmt.Printf("received %d status: %v", response.StatusCode(), errResponse)
		return
	}

	result := response.Result().(*CreateNewPaymentResponse)

	fmt.Printf("result: %v", result)
}

func (r *CreateNewPaymentRequest) createSignatureMessage() string {
	// Missing mandatory fields: [amount, beneficiaryName, beneficiaryAccount, beneficiaryReference, transactionSignature, requestTime]}

	var message string
	message += fmt.Sprintf("amount=%s", r.Amount)
	message += fmt.Sprintf("&beneficiaryName=%s", r.BeneficiaryName)
	message += fmt.Sprintf("&beneficiaryAccount=%s", r.BeneficiaryAccount)
	message += fmt.Sprintf("&beneficiaryReference=%s", r.BeneficiaryReference)

	//message += fmt.Sprintf("&requestTime=%d", r.RequestTime)
	//message += fmt.Sprintf("&account=%s", r.Account)


	//if r.BeneficiaryAddress != "" {
	//	message += fmt.Sprintf("&beneficiaryAddress=%s", r.BeneficiaryAddress)
	//}
	//if r.UseGbxForFee != false {
	//	message += fmt.Sprintf("&useGbxForFee=%t", r.UseGbxForFee)
	//}

	fmt.Println(message)

	return message
}

func createAuthHeaders(path string, formData string, nonce int64) map[string]string {
	contentType := "application/json"

	nonceStr := fmt.Sprintf("%d", nonce)

	message := SECRETS["APIKey"]
	message += "&"
	message += nonceStr
	message += path

	fmt.Println(formData)

	// Include TransactionSignature signature here?
	if formData != "" {
		message += "?"
		message += formData
	}

	fmt.Println("createAuthHeaders", message)

	signature := GenerateHMACSHA512([]byte(message), []byte(SECRETS["MessageSigningSecretKey"]), &HMACSHA512Options{})

	return map[string]string{
		"X-API-Key":    SECRETS["APIKey"],
		"X-Nonce":      nonceStr,
		"X-Signature":  signature,
		"Content-Type": contentType,
		"Accept":       contentType,
	}
}

func GenerateHMACSHA512(message, key []byte, options *HMACSHA512Options) string {
	mac := hmac.New(sha512.New, key)
	mac.Write(message)

	macSum := mac.Sum(nil)

	if options.Encoding == "base64" {
		return base64.StdEncoding.EncodeToString(macSum)
	} else {
		return hex.EncodeToString(macSum)
	}
}

type HMACSHA512Options struct {
	Encoding string
}

type CreateNewPaymentRequest struct {
	// RequestTime request time in Unix timestamp format. Precision - milliseconds.
	RequestTime int64 `json:"requestTime"`

	// Account number from what the funds will be transferred.
	Account string `json:"account"`

	// Amount funds amount to transfer.
	Amount string `json:"amount"`

	// BeneficiaryName beneficiary name of the specified beneficiary account.
	BeneficiaryName string `json:"beneficiaryName"`

	// BeneficiaryAddress beneficiary address.
	BeneficiaryAddress string `json:"beneficiaryAddress,omitempty"`

	// BeneficiaryAccount exchange account number for the beneficiary.
	BeneficiaryAccount string `json:"beneficiaryAccount"`

	// BeneficiaryReference reference for beneficiary.
	BeneficiaryReference string `json:"beneficiaryReference"`

	// UseGbxForFee should GBX token be used to cover transaction fee.
	//UseGbxForFee bool `json:"useGbxForFee,omitempty"`

	// TransactionSignature transaction signature. lower-case hex representation of hmac-sha512 of concatenated request parameters (name=value) delimited by “&” symbol. Note that concatenation parameters should be in a strict order.
	TransactionSignature string `json:"transactionSignature"`
}

type CreateNewPaymentResponse struct {
	ID     string `json:"paymentId"`
	Status string `json:"status"`
}

type ErrorResult struct {
	Errors []Error `json:"errors"`
}

type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    string `json:"data,omitempty"`
}
