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
	perPage = 17
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
	pdf.AddUTF8Font("Consolas", "", "Consolas.ttf")
	pdf.AddUTF8Font("Arial", "", "Arial.ttf")
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

	for i, user := range users {
		if i%perPage == 0 {
			pdf.AddPage()
			y := pdf.GetY() // current Y position after writing
			// Set dashed pattern: dash length 1mm, gap length 1mm
			pdf.SetDashPattern([]float64{1, 1}, 0)
			pdf.SetLineWidth(0.1)
			pdf.Line(10, y, 200, y)
			pdf.Ln(2)
		}

		pdf.SetFont("Consolas", "", 10)
		pdf.SetTextColor(0, 0, 0)
		pdf.CellFormat(110, 6, fmt.Sprintf("Username: %s", user.Name), "", 0, "", false, 0, "")
		pdf.CellFormat(0, 6, fmt.Sprintf("Password: %s", user.Password), "", 1, "", false, 0, "")

		pdf.SetFont("Arial", "", 8)
		pdf.SetTextColor(50, 50, 50)
		pdf.CellFormat(110, 4, "En me connectant je reconnais avoir lu et m'engage à respecter la charte informatique de l'UCBL1 (https://www.univ-lyon1.fr/charte-informatique).", "", 1, "", false, 0, "")

		pdf.Ln(2)

		y := pdf.GetY() // current Y position after writing
		// Set dashed pattern: dash length 1mm, gap length 1mm
		pdf.SetDashPattern([]float64{1, 1}, 0)
		pdf.SetLineWidth(0.1)
		pdf.Line(10, y, 200, y)
		pdf.Ln(2)

		pdf.SetDashPattern([]float64{}, 0)
	}

	fmt.Printf("Exporting %d credentials to %s\n", len(users), fileStr)
	if err := pdf.OutputFileAndClose(fileStr); err != nil {
		log.Fatal(err)
	}
}
