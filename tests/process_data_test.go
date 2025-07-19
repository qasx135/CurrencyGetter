package tests

import (
	"testing"
	"time"

	"SomeTask/internal/app"
	"SomeTask/internal/models"
)

func TestProcessData(t *testing.T) {
	now := time.Now()

	data := []models.ComputeStruct{
		{Date: now, Name: "Евро", RealCurrency: 90.0, NumCode: "978"},
		{Date: now, Name: "Доллар", RealCurrency: 75.5, NumCode: "840"},
		{Date: now, Name: "Йена", RealCurrency: 0.6, NumCode: "392"},
	}

	max, min, avg := app.ProcessData(data)

	if max.Name != "Евро" || max.RealCurrency != 90.0 {
		t.Errorf("Expected max to be Euro with 90.0, got %v", max)
	}

	if min.Name != "Йена" || min.RealCurrency != 0.6 {
		t.Errorf("Expected min to be Yen with 0.6, got %v", min)
	}

	expectedAvg := (90.0 + 75.5 + 0.6) / 3
	if avg != expectedAvg {
		t.Errorf("Expected avg to be %.2f, got %.2f", expectedAvg, avg)
	}
}
