// internal/models/print_job.go
package models

import (
	"encoding/json"
)

type PrintJobStatus string

const (
	Queued    PrintJobStatus = "Queued"
	Running   PrintJobStatus = "Running"
	Done      PrintJobStatus = "Done"
	Cancelled PrintJobStatus = "Cancelled"
)

type PrintJob struct {
	ID                 string         `json:"id"`
	PrinterID          string         `json:"printer_id"`
	FilamentID         string         `json:"filament_id"`
	FilePath           string         `json:"filepath"`
	PrintWeightInGrams int            `json:"print_weight_in_grams"`
	Status             PrintJobStatus `json:"status"`
}

func (pj *PrintJob) Serialize() ([]byte, error) {
	return json.Marshal(pj)
}

func DeserializePrintJob(data []byte) (*PrintJob, error) {
	var printJob PrintJob
	err := json.Unmarshal(data, &printJob)
	return &printJob, err
}

func (pj *PrintJob) IsValidStatusTransition(newStatus PrintJobStatus) bool {
	switch newStatus {
	case Running:
		return pj.Status == Queued
	case Done:
		return pj.Status == Running
	case Cancelled:
		return pj.Status == Queued || pj.Status == Running
	default:
		return false
	}
}
