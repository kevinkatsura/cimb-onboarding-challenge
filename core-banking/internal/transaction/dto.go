package transaction

import (
	"time"

	"core-banking/internal/snap"
)

type OriginatorInfo struct {
	OriginatorCustomerNo   string `json:"originatorCustomerNo"`
	OriginatorCustomerName string `json:"originatorCustomerName"`
	OriginatorBankCode     string `json:"originatorBankCode"`
}

type IntrabankTransferRequest struct {
	PartnerReferenceNo   string            `json:"partnerReferenceNo"`
	Amount               snap.SNAPAmount   `json:"amount"`
	BeneficiaryAccountNo string            `json:"beneficiaryAccountNo"`
	CustomerReference    *string           `json:"customerReference,omitempty"`
	FeeType              string            `json:"feeType"`
	OriginatorInfos      *[]OriginatorInfo `json:"originatorInfos,omitempty"`
	Remark               *string           `json:"remark,omitempty"`
	SourceAccountNo      string            `json:"sourceAccountNo"`
	TransactionDate      string            `json:"transactionDate"`
	AdditionalInfo       interface{}       `json:"additionalInfo,omitempty"`
}

type IntrabankTransferResponse struct {
	ResponseCode         string           `json:"responseCode"`
	ResponseMessage      string           `json:"responseMessage"`
	Amount               snap.SNAPAmount  `json:"amount"`
	BeneficiaryAccountNo string           `json:"beneficiaryAccountNo"`
	OriginatorInfos      []OriginatorInfo `json:"originatorInfos"`
	TransactionDate      string           `json:"transactionDate"`

	ReferenceNo         *string     `json:"referenceNo,omitempty"`
	PartnerReferenceNo  *string     `json:"partnerReferenceNo,omitempty"`
	BeneficiaryBankCode *string     `json:"beneficiaryBankCode,omitempty"`
	SourceAccountNo     *string     `json:"sourceAccountNo,omitempty"`
	TraceNo             *string     `json:"traceNo,omitempty"`
	AdditionalInfo      interface{} `json:"additionalInfo,omitempty"`
	CustomerReference   *string     `json:"customerReference,omitempty"`
}

type TransferStatusInquiryRequest struct {
	OriginalPartnerReferenceNo string `json:"originalPartnerReferenceNo"`
	OriginalReferenceNo        string `json:"originalReferenceNo"`
	ServiceCode                string `json:"serviceCode"`
}

type TransferStatusInquiryResponse struct {
	ResponseCode               string          `json:"responseCode"`
	ResponseMessage            string          `json:"responseMessage"`
	OriginalPartnerReferenceNo string          `json:"originalPartnerReferenceNo"`
	OriginalReferenceNo        string          `json:"originalReferenceNo"`
	ServiceCode                string          `json:"serviceCode"`
	Amount                     snap.SNAPAmount `json:"amount"`
	LatestTransactionStatus    string          `json:"latestTransactionStatus"`
	TransactionStatusDesc      string          `json:"transactionStatusDesc"`
	ReferenceNumber            string          `json:"referenceNumber"`
}

type TransferRequest struct {
	ReferenceID string `json:"reference_id"`
	FromAccount string `json:"from_account"`
	ToAccount   string `json:"to_account"`
	Amount      int64  `json:"amount"`
	Currency    string `json:"currency"`
}

type TransferResponse struct {
	Status string `json:"status"`

	TransactionID *string `json:"transaction_id,omitempty"`

	SourceAccount      string `json:"source_account"`
	DestinationAccount string `json:"destination_account"`

	Amount int64 `json:"amount"`

	// Success fields
	SourceBalanceAfter      *int64 `json:"source_balance_after,omitempty"`
	DestinationBalanceAfter *int64 `json:"destination_balance_after,omitempty"`

	// Failure fields
	CurrentBalance *int64 `json:"current_balance,omitempty"`

	Message string `json:"message"`
}

type TransactionHistoryResponse struct {
	LedgerEntryID      string  `db:"ledger_entry_id" json:"ledger_entry_id"`
	TransactionID      string  `db:"transaction_id" json:"transaction_id"`
	PartnerReferenceNo string  `db:"partner_reference_no" json:"partnerReferenceNo"`
	ReferenceNo        *string `db:"reference_no" json:"referenceNo"`

	AccountID     string `db:"account_id" json:"accountId"`
	AccountNumber string `db:"account_number" json:"accountNumber"`

	TransactionType string `db:"transaction_type" json:"transaction_type"`
	Status          string `db:"status" json:"status"`

	JournalType *string `db:"journal_type" json:"journal_type"`
	EntryType   *string `db:"entry_type" json:"entry_type"`

	Amount       int64  `db:"amount" json:"amount"`
	Currency     string `db:"currency" json:"currency"`
	BalanceAfter *int64 `db:"balance_after" json:"balance_after"`

	Description *string `db:"description" json:"description"`

	CreatedAt   time.Time  `db:"created_at" json:"created_at"`
	CompletedAt *time.Time `db:"completed_at" json:"completed_at"`
}
