package model

import (
	"encoding/json"
	"testing"
)

func TestValidatePayload(t *testing.T) {
	tests := []struct {
		name     string
		taskType TaskType
		payload  json.RawMessage
		wantErr  bool
	}{
		{
			name:     "valid email",
			taskType: TypeEmail,
			payload:  json.RawMessage(`{"to":"john@mail.ru", "subject":"hi", "body":"text"}`),
			wantErr:  false,
		},
		{
			name:     "empty to",
			taskType: TypeEmail,
			payload:  json.RawMessage(`{"to":"", "subject":"hi", "body":"text"}`),
			wantErr:  true,
		},
		{
			name:     "invalid json",
			taskType: TypeEmail,
			payload:  json.RawMessage(`{"invalid...}`),
			wantErr:  true,
		},
		{
			name:     "valid image",
			taskType: TypeImage,
			payload:  json.RawMessage(`{"width":120, "height": 100, "url": "localhost"}`),
			wantErr:  false,
		},
		{
			name:     "invalid width",
			taskType: TypeImage,
			payload:  json.RawMessage(`{"width":-120, "height": 100, "url": "localhost"}`),
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePayload(tt.taskType, tt.payload)

			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePayload() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
