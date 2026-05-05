package dto

type CreateCheckoutSessionRequest struct {
	PlanID     string `json:"plan_id" validate:"required,uuid"`
	SuccessURL string `json:"success_url" validate:"required,max=2000"`
	CancelURL  string `json:"cancel_url" validate:"required,max=2000"`
}

type CreatePortalSessionRequest struct {
	ReturnURL string `json:"return_url" validate:"required,max=2000"`
}

type HostedBillingSessionResponse struct {
	URL string `json:"url"`
}

type BillingInvoiceResponse struct {
	ID               string `json:"id"`
	Number           string `json:"number"`
	Status           string `json:"status"`
	AmountPaid       int64  `json:"amount_paid"`
	AmountDue        int64  `json:"amount_due"`
	Currency         string `json:"currency"`
	HostedInvoiceURL string `json:"hosted_invoice_url,omitempty"`
	InvoicePDF       string `json:"invoice_pdf,omitempty"`
	CreatedAt        string `json:"created_at"`
}
