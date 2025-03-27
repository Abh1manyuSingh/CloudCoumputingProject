// internal/models/printer.go
package models

import (
	"encoding/json"
)

type Printer struct {
	ID      string `json:"id"`
	Company string `json:"company"`
	Model   string `json:"model"`
}

func (p *Printer) Serialize() ([]byte, error) {
	return json.Marshal(p)
}

func DeserializePrinter(data []byte) (*Printer, error) {
	var printer Printer
	err := json.Unmarshal(data, &printer)
	return &printer, err
}
