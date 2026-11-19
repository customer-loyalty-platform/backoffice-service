package model

import "time"

type PartnerMerchant struct {
	ID                string    `json:"id"`
	MerchantCode      string    `json:"merchant_code"`
	MerchantName      string    `json:"merchant_name"`
	ContactEmail      string    `json:"contact_email"`
	ContactPhone      string    `json:"contact_phone"`
	ParValuePerPoint  float64   `json:"par_value_per_point"`
	CommissionRate    float64   `json:"commission_rate"`
	SettlementCurrency string   `json:"settlement_currency"`
	Status            string    `json:"status"`
	ContractStartDate string    `json:"contract_start_date"`
	ContractEndDate   *string   `json:"contract_end_date,omitempty"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type CreatePartnerRequest struct {
	MerchantCode      string  `json:"merchant_code" binding:"required"`
	MerchantName      string  `json:"merchant_name" binding:"required"`
	ContactEmail      string  `json:"contact_email" binding:"required"`
	ContactPhone      string  `json:"contact_phone"`
	ParValuePerPoint  float64 `json:"par_value_per_point" binding:"required"`
	CommissionRate    float64 `json:"commission_rate" binding:"required"`
	SettlementCurrency string `json:"settlement_currency" binding:"required"`
	ContractStartDate string  `json:"contract_start_date" binding:"required"`
	ContractEndDate   *string `json:"contract_end_date,omitempty"`
}

type UpdatePartnerRequest = CreatePartnerRequest

type SettlementTransaction struct {
	TransactionID  string  `json:"transaction_id"`
	MemberID       string  `json:"member_id"`
	PointsRedeemed int64   `json:"points_redeemed"`
	MonetaryValue  float64 `json:"monetary_value"`
	RedeemedAt     string  `json:"redeemed_at"`
}

type SettlementReport struct {
	Partner struct {
		ID           string `json:"id"`
		MerchantName string `json:"merchant_name"`
		MerchantCode string `json:"merchant_code"`
	} `json:"partner"`
	Period struct {
		Start string `json:"start"`
		End   string `json:"end"`
	} `json:"period"`
	Summary struct {
		TotalPointsRedeemed int64   `json:"total_points_redeemed"`
		ParValuePerPoint    float64 `json:"par_value_per_point"`
		GrossLiability      float64 `json:"gross_liability"`
		CommissionRate      float64 `json:"commission_rate"`
		PlatformCommission  float64 `json:"platform_commission"`
		NetPayableToPartner float64 `json:"net_payable_to_partner"`
	} `json:"summary"`
	Transactions []SettlementTransaction `json:"transactions"`
}

type SettlementReportParams struct {
	PartnerID   string `form:"partner_id" binding:"required"`
	PeriodStart string `form:"period_start"`
	PeriodEnd   string `form:"period_end"`
}

