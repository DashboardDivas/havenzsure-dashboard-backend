package workorder

import (
	"bytes"
	"fmt"
	"time"

	"github.com/DashboardDivas/havenzsure-dashboard-backend/internal/workorder/dto"

	"github.com/jung-kurt/gofpdf"
)

type nopWriteCloser struct {
	*bytes.Buffer
}

func (n nopWriteCloser) Close() error { return nil }

func BuildWorkOrderPDF(wo dto.WorkOrderDetail) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "Letter", "")
	pdf.SetMargins(15, 15, 15)
	pdf.AddPage()

	// ===== Header =====
	pdf.SetFont("Arial", "B", 18)
	pdf.Cell(0, 10, "HavenzSure Work Order")
	pdf.Ln(12)

	pdf.SetFont("Arial", "", 11)
	pdf.Cell(0, 7, fmt.Sprintf("Work Order: %s", wo.Code))
	pdf.Ln(6)
	pdf.Cell(0, 7, fmt.Sprintf("Status: %s", wo.Status))
	pdf.Ln(8)

	drawLine(pdf)

	// ===== Dates =====
	sectionTitle(pdf, "Dates")
	keyValue(pdf, "Date Received", formatDate(wo.DateReceived))
	keyValue(pdf, "Last Updated", formatDate(wo.DateUpdated))

	// ===== Customer =====
	sectionTitle(pdf, "Customer Information")
	keyValue(pdf, "Name", wo.Customer.FullName)
	keyValue(pdf, "Email", wo.Customer.Email)
	keyValue(pdf, "Phone", wo.Customer.Phone)
	keyValue(pdf, "Address",
		fmt.Sprintf("%s, %s, %s %s",
			wo.Customer.Address,
			wo.Customer.City,
			wo.Customer.Province,
			wo.Customer.PostalCode,
		),
	)

	// ===== Vehicle =====
	sectionTitle(pdf, "Vehicle Information")
	keyValue(pdf, "Vehicle",
		fmt.Sprintf("%d %s %s",
			wo.Vehicle.ModelYear,
			wo.Vehicle.Make,
			wo.Vehicle.Model,
		),
	)
	keyValue(pdf, "Plate Number", wo.Vehicle.PlateNo)
	keyValue(pdf, "VIN", wo.Vehicle.VIN)
	keyValue(pdf, "Body Style", wo.Vehicle.BodyStyle)
	keyValue(pdf, "Color", wo.Vehicle.Color)

	// ===== Insurance (optional) =====
	if wo.Insurance != nil {
		sectionTitle(pdf, "Insurance Information")
		keyValue(pdf, "Company", wo.Insurance.InsuranceCompany)
		keyValue(pdf, "Policy Number", wo.Insurance.PolicyNumber)
		keyValue(pdf, "Claim Number", wo.Insurance.ClaimNumber)
		keyValue(pdf, "Agent",
			fmt.Sprintf("%s (%s)",
				wo.Insurance.AgentFullName,
				wo.Insurance.AgentPhone,
			),
		)
	}

	// ===== Shop =====
	sectionTitle(pdf, "Shop Information")
	keyValue(pdf, "Shop Name", wo.Shop.ShopName)

	var buf bytes.Buffer
	if err := pdf.OutputAndClose(nopWriteCloser{&buf}); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// Helper functions
func sectionTitle(pdf *gofpdf.Fpdf, title string) {
	pdf.Ln(6)
	pdf.SetFont("Arial", "B", 13)
	pdf.Cell(0, 8, title)
	pdf.Ln(8)
	pdf.SetFont("Arial", "", 11)
}

func keyValue(pdf *gofpdf.Fpdf, key, value string) {
	pdf.CellFormat(45, 6, key+":", "", 0, "", false, 0, "")
	pdf.MultiCell(0, 6, value, "", "", false)
}

func drawLine(pdf *gofpdf.Fpdf) {
	y := pdf.GetY()
	pdf.Line(15, y, 195, y)
	pdf.Ln(6)
}

func formatDate(t time.Time) string {
	if t.IsZero() {
		return "-"
	}
	return t.Format("2006-01-02")
}
