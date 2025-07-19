package tests

import (
	"SomeTask/internal/models"
	"testing"
	"time"
)

func TestComputeStruct_Validate(t *testing.T) {
	tests := []struct {
		name    string
		cs      models.ComputeStruct
		wantErr bool
	}{
		{
			name: "Valid struct",
			cs: models.ComputeStruct{
				Date:         time.Now(),
				NumCode:      "840",
				Name:         "Доллар США",
				RealCurrency: 75.5,
			},
			wantErr: false,
		},
		{
			name: "Empty NumCode",
			cs: models.ComputeStruct{
				Date:         time.Now(),
				Name:         "Доллар США",
				RealCurrency: 75.5,
			},
			wantErr: true,
		},
		{
			name: "Negative rate",
			cs: models.ComputeStruct{
				Date:         time.Now(),
				NumCode:      "840",
				Name:         "Доллар США",
				RealCurrency: -10.0,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cs.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("ComputeStruct.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
