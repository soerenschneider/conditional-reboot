package internal

import "testing"

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  *ConditionalRebootConfig
		wantErr bool
	}{
		{
			name: "only metrics addr specified",
			config: &ConditionalRebootConfig{
				MetricsListenAddr: ":127.0.0.1:9199",
			},
			wantErr: false,
		},
		{
			name: "only metrics dir specified",
			config: &ConditionalRebootConfig{
				MetricsDir: "/tmp",
			},
			wantErr: false,
		},
		{
			name: "both metrics addr and metrics dir specified",
			config: &ConditionalRebootConfig{
				MetricsListenAddr: ":127.0.0.1:9199",
				MetricsDir:        "/tmp",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ValidateConfig(tt.config); (err != nil) != tt.wantErr {
				t.Errorf("ValidateConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
