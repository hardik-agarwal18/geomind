package geo

import (
	"math"
	"testing"
)

func TestCoordinateValidate(t *testing.T) {
	tests := []struct {
		name    string
		input   Coordinate
		wantErr bool
	}{
		{
			name: "valid Bengaluru",
			input: Coordinate{
				Longitude: 77.5946,
				Latitude:  12.9716,
			},
		},
		{
			name: "valid origin",
			input: Coordinate{
				Longitude: 0,
				Latitude:  0,
			},
		},
		{
			name: "valid boundary",
			input: Coordinate{
				Longitude: 180,
				Latitude:  90,
			},
		},
		{
			name: "longitude too high",
			input: Coordinate{
				Longitude: 181,
				Latitude:  12,
			},
			wantErr: true,
		},
		{
			name: "longitude too low",
			input: Coordinate{
				Longitude: -181,
				Latitude:  12,
			},
			wantErr: true,
		},
		{
			name: "latitude too high",
			input: Coordinate{
				Longitude: 77,
				Latitude:  91,
			},
			wantErr: true,
		},
		{
			name: "latitude too low",
			input: Coordinate{
				Longitude: 77,
				Latitude:  -91,
			},
			wantErr: true,
		},
		{
			name: "NaN longitude",
			input: Coordinate{
				Longitude: math.NaN(),
				Latitude:  12,
			},
			wantErr: true,
		},
		{
			name: "infinite latitude",
			input: Coordinate{
				Longitude: 77,
				Latitude:  math.Inf(1),
			},
			wantErr: true,
		},
		{
			name: "infinite longitude",
			input: Coordinate{
				Longitude: math.Inf(-1),
				Latitude:  12,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.input.Validate()

			if (err != nil) != tt.wantErr {
				t.Fatalf(
					"Validate() error = %v, wantErr %v",
					err,
					tt.wantErr,
				)
			}
		})
	}
}
