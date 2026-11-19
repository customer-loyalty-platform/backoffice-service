package model

import "time"

type VoucherBatch struct {
	ID              string    `json:"id"`
	BatchCode       string    `json:"batch_code"`
	Title           string    `json:"title"`
	Description     string    `json:"description,omitempty"`
	PartnerBrand    string    `json:"partner_brand,omitempty"`
	MonetaryValue   float64   `json:"monetary_value,omitempty"`
	PointsCost      int64     `json:"points_cost"`
	TermsConditions string    `json:"terms_conditions,omitempty"`
	TotalStock      int64     `json:"total_stock"`
	AvailableStock  int64     `json:"available_stock"`
	ClaimedStock    int64     `json:"claimed_stock"`
	RevokedCount    int64     `json:"revoked_count"`
	Status          string    `json:"status"`
	CreatedAt       time.Time `json:"created_at"`
}

type CreateVoucherBatchRequest struct {
	BatchCode       string  `json:"batch_code" binding:"required"`
	Title           string  `json:"title" binding:"required"`
	Description     string  `json:"description"`
	PartnerBrand    string  `json:"partner_brand"`
	MonetaryValue   float64 `json:"monetary_value"`
	PointsCost      int64   `json:"points_cost" binding:"required"`
	TermsConditions string  `json:"terms_conditions"`
}

type UploadCodesResponse struct {
	BatchID           string `json:"batch_id"`
	CodesImported     int    `json:"codes_imported"`
	DuplicatesSkipped int    `json:"duplicates_skipped"`
	NewTotalStock     int64  `json:"new_total_stock"`
	NewAvailableStock int64  `json:"new_available_stock"`
}

type VoucherInventoryItem struct {
	ID          string  `json:"id"`
	BatchID     string  `json:"batch_id"`
	VoucherCode string  `json:"voucher_code"`
	Status      string  `json:"status"`
	ClaimedBy   *string `json:"claimed_by,omitempty"`
	ClaimedAt   *string `json:"claimed_at,omitempty"`
}

type RevokeVouchersRequest struct {
	VoucherIDs []string `json:"voucher_ids" binding:"required"`
	Reason     string   `json:"reason"`
}

