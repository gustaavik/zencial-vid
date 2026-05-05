package mapper

import (
	"github.com/zenfulcode/zencial/internal/adapter/handler/v1/dto"
	billinguc "github.com/zenfulcode/zencial/internal/usecase/billing"
)

func BillingSessionToResponse(session *billinguc.HostedSession) dto.HostedBillingSessionResponse {
	return dto.HostedBillingSessionResponse{
		URL: session.URL,
	}
}

func BillingInvoiceToResponse(invoice *billinguc.Invoice) dto.BillingInvoiceResponse {
	return dto.BillingInvoiceResponse{
		ID:               invoice.ID,
		Number:           invoice.Number,
		Status:           invoice.Status,
		AmountPaid:       invoice.AmountPaid,
		AmountDue:        invoice.AmountDue,
		Currency:         invoice.Currency,
		HostedInvoiceURL: invoice.HostedInvoiceURL,
		InvoicePDF:       invoice.InvoicePDF,
		CreatedAt:        invoice.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

func BillingInvoicesToResponse(invoices []*billinguc.Invoice) []dto.BillingInvoiceResponse {
	result := make([]dto.BillingInvoiceResponse, len(invoices))
	for i := range invoices {
		result[i] = BillingInvoiceToResponse(invoices[i])
	}
	return result
}
