package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"

	"codeberg.org/go-pdf/fpdf"
)

const (
	fileStr = "auth-creds.pdf"
)

type User struct {
	Name     string `json:"name"`
	Password string `json:"password"`
}

func main() {
	b, err := io.ReadAll(os.Stdin)
	if err != nil {
		log.Fatal(err)
	}
	users := []*User{}
	if err := json.Unmarshal(b, &users); err != nil {
		log.Fatal(err)
	}

	pdf := fpdf.New(fpdf.OrientationPortrait, "mm", "A4", "")
	pdf.SetTopMargin(30)
	pdf.SetHeaderFuncMode(func() {
		pdf.Image("logo.png", 10, 6, 20, 0, false, "", 0, "")
		pdf.Image("logo-target.png", 180, 6, 20, 0, false, "", 0, "")
		pdf.SetY(5)
		pdf.SetFont("Arial", "B", 15)
		pdf.Cell(55, 0, "")
		pdf.CellFormat(80, 10, "Credentials 24h IUT 2025", "1", 0, "C", false, 0, "")
		pdf.Ln(20)
	}, true)
	pdf.SetFooterFunc(func() {
		pdf.SetY(-20)
		pdf.SetFont("Arial", "I", 8)
		pdf.CellFormat(0, 5, fmt.Sprintf("Page %d/{nb}", pdf.PageNo()), "", 1, "C", false, 0, "")

		// TLP:RED centered below the page number
		pdf.SetFont("Arial", "B", 12)
		pdf.SetTextColor(255, 0, 0)
		pdf.CellFormat(0, 5, "TLP:RED", "", 0, "C", false, 0, "")
	})
	pdf.AliasNbPages("")

	pdf.SetFont("Arial", "", 10)
	pdf.AddPage()
	for _, user := range users {
		pdf.CellFormat(110, 8, fmt.Sprintf("Username: %s", user.Name), "", 0, "", false, 0, "")
		pdf.CellFormat(0, 8, fmt.Sprintf("Password: %s", user.Password), "", 1, "", false, 0, "")

		y := pdf.GetY() // current Y position after writing
		// Set dashed pattern: dash length 1mm, gap length 1mm
		pdf.SetDashPattern([]float64{1, 1}, 0)
		pdf.SetLineWidth(0.1)
		pdf.Line(10, y, 200, y)

		// Reset to solid line (important if you later want solid lines again)
		pdf.SetDashPattern([]float64{}, 0)
	}

	fmt.Printf("Exporting %d credentials to %s\n", len(users), fileStr)
	if err := pdf.OutputFileAndClose(fileStr); err != nil {
		log.Fatal(err)
	}
}
