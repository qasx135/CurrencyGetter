package models

import (
	"encoding/xml"
	"errors"
	"time"
)

// ValCurs структура для получения всей схемы
type ValCurs struct {
	XMLName xml.Name `xml:"ValCurs"`
	Date    string   `xml:"Date,attr"`
	Name    string   `xml:"name,attr"`
	Valutes []Valute `xml:"Valute"`
}

// Valute структура, в которую складывается информация по каждой валюте
type Valute struct {
	ID        string `xml:"ID,attr"`
	NumCode   string `xml:"NumCode"`
	CharCode  string `xml:"CharCode"`
	Nominal   string `xml:"Nominal"`
	Name      string `xml:"Name"`
	Value     string `xml:"Value"`
	VunitRate string `xml:"VunitRate"`
}

// ComputeStruct структруа для сравнения и получение конечного результата, получается после обработки первичных данных
type ComputeStruct struct {
	Date         time.Time
	NumCode      string
	Name         string
	RealCurrency float64 // Value / Nominal (1 to 1)
}

func (cs *ComputeStruct) Validate() error {
	if cs.NumCode == "" {
		return errors.New("num code is required")
	}
	if cs.Name == "" {
		return errors.New("name is required")
	}
	if cs.RealCurrency <= 0 {
		return errors.New("real currency must be positive")
	}
	if cs.Date.IsZero() {
		return errors.New("date is required")
	}
	return nil
}
