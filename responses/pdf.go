package responses

import (
	"fmt"
	"log"
	"net/http"
)

func WritePDFBytes(w http.ResponseWriter, filename string, PDFBytes []byte) {
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf("inline; filename=%q", filename))
	w.WriteHeader(http.StatusOK) // Response Header Sent & Frozen
	_, err := w.Write(PDFBytes)
	if err != nil {
		log.Printf("[ERROR] writing PDF to response: %v", err)
	}
}
