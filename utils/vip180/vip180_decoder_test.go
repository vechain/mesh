package vip180

import (
	"strings"
	"testing"
)

func TestIsVIP180TransferCallData(t *testing.T) {
	tests := []struct {
		name     string
		data     string
		expected bool
	}{
		{
			name:     "Valid transfer call data",
			data:     "0xa9059cbb000000000000000000000000f077b491b355e64048ce21e3a6fc4751eeea77fa0000000000000000000000000000000000000000000000000de0b6b3a7640000",
			expected: true,
		},
		{
			name:     "Invalid call data",
			data:     "0x1234567890abcdef",
			expected: false,
		},
		{
			name:     "Empty data",
			data:     "",
			expected: false,
		},
		{
			name:     "Non-hex data",
			data:     "invalid",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsVIP180TransferCallData(tt.data)
			if result != tt.expected {
				t.Errorf("IsVIP180TransferCallData() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestDecodeVIP180TransferCallData(t *testing.T) {
	tests := []struct {
		name        string
		data        string
		expectedTo  string
		expectedAmt string
		expectError bool
	}{
		{
			name:        "Valid transfer data",
			data:        "0xa9059cbb000000000000000000000000f077b491b355e64048ce21e3a6fc4751eeea77fa0000000000000000000000000000000000000000000000000de0b6b3a7640000",
			expectedTo:  "0xf077b491b355e64048ce21e3a6fc4751eeea77fa",
			expectedAmt: "1000000000000000000",
			expectError: false,
		},
		{
			name:        "Invalid data length",
			data:        "0xa9059cbb000000000000000000000000f077b491b355e64048ce21e3a6fc4751eeea77fa",
			expectedTo:  "",
			expectedAmt: "",
			expectError: true,
		},
		{
			name:        "Invalid hex data",
			data:        "0xinvalid",
			expectedTo:  "",
			expectedAmt: "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := DecodeVIP180TransferCallData(tt.data)

			if tt.expectError {
				if err == nil {
					t.Errorf("DecodeVIP180TransferCallData() error = nil, want error")
				}
				return
			}

			if err != nil {
				t.Errorf("DecodeVIP180TransferCallData() error = %v, want nil", err)
				return
			}

			resultTo := strings.ToLower(result.To.String())
			expectedTo := strings.ToLower(tt.expectedTo)

			if resultTo != expectedTo {
				t.Errorf("DecodeVIP180TransferCallData() To = %v, want %v", resultTo, expectedTo)
			}

			if result.Amount.String() != tt.expectedAmt {
				t.Errorf("DecodeVIP180TransferCallData() Amount = %v, want %v", result.Amount.String(), tt.expectedAmt)
			}
		})
	}
}

func TestEncodeVIP180TransferCallData(t *testing.T) {
	tests := []struct {
		name    string
		to      string
		amount  string
		wantErr bool
	}{
		{
			name:    "Valid transfer data",
			to:      "0xf077b491b355e64048ce21e3a6fc4751eeea77fa",
			amount:  "1000000000000000000",
			wantErr: false,
		},
		{
			name:    "Invalid amount",
			to:      "0xf077b491b355e64048ce21e3a6fc4751eeea77fa",
			amount:  "invalid",
			wantErr: true,
		},
		{
			name:    "Empty amount",
			to:      "0xf077b491b355e64048ce21e3a6fc4751eeea77fa",
			amount:  "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := EncodeVIP180TransferCallData(tt.to, tt.amount)
			if (err != nil) != tt.wantErr {
				t.Errorf("EncodeVIP180TransferCallData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if !strings.HasPrefix(result, "0x") {
					t.Errorf("EncodeVIP180TransferCallData() result should start with 0x, got %s", result)
				}
				if len(result) < 10 {
					t.Errorf("EncodeVIP180TransferCallData() result too short, got %s", result)
				}
			}
		})
	}
}
