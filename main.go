package main

import (
	"fmt"
	"time"
)

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
	UseGbxForFee bool `json:"useGbxForFee,omitempty"`

	// TransactionSignature transaction signature. lower-case hex representation of hmac-sha512 of concatenated request parameters (name=value) delimited by “&” symbol. Note that concatenation parameters should be in a strict order.
	TransactionSignature string `json:"transactionSignature"`
}

func main() {
	payment := CreateNewPaymentRequest{
		RequestTime:          time.Now().UnixNano() / 1e6,
		Amount:               "1",
		Account:              "LT543080020000000224",
		BeneficiaryReference: fmt.Sprintf("Testing payment"),
		BeneficiaryName:      "TODO",
		BeneficiaryAccount:   "TODO",
	}

	if params.Sepa.Address != "" {
		payment.BeneficiaryAddress = fmt.Sprintf("%s", params.Sepa.Address)
	}

	message := payment.createSignatureMessage()
	payment.TransactionSignature = utils.GenerateHMACSHA512([]byte(message), []byte(c.secrets.TransactionSigningSecretKey), &utils.HMACSHA512Options{})

	path := "/api/1/eurowallet/payments"
	headers := c.createAuthHeaders(path, message)

	response, err := c.RestyClient.R().
		SetBody(payment).
		SetHeaders(headers).
		SetResult(CreateNewPaymentResponse{}).
		SetError(ErrorResult{}).
		Post(fmt.Sprintf("%s%s", c.GlobitexURL, path))
}


func (r *CreateNewPaymentRequest) createSignatureMessage() string {
	var message string
	message += fmt.Sprintf("requestTime=%d", r.RequestTime)
	message += fmt.Sprintf("&account=%s", r.Account)
	message += fmt.Sprintf("&amount=%s", r.Amount.Decimal)
	message += fmt.Sprintf("&beneficiaryName=%s", r.BeneficiaryName)
	if r.BeneficiaryAddress != "" {
		message += fmt.Sprintf("&beneficiaryAddress=%s", r.BeneficiaryAddress)
	}
	message += fmt.Sprintf("&beneficiaryAccount=%s", r.BeneficiaryAccount)
	if r.BeneficiaryReference != "" {
		message += fmt.Sprintf("&beneficiaryReference=%s", r.BeneficiaryReference)
	}
	if r.UseGbxForFee != false {
		message += fmt.Sprintf("&useGbxForFee=%t", r.UseGbxForFee)
	}

	return message
}
