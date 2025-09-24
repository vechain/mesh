package thor

import (
	"testing"
)

func TestConfig(t *testing.T) {
	config := Config{
		NodeID:      "test-node",
		NetworkType: "solo",
		APIAddr:     "localhost:8669",
		P2PPort:     11235,
		OnDemand:    true,
		Persist:     false,
		APICORS:     "*",
	}

	// Test that config fields are set correctly
	if config.NodeID != "test-node" {
		t.Errorf("Config.NodeID = %v, want test-node", config.NodeID)
	}
	if config.NetworkType != "solo" {
		t.Errorf("Config.NetworkType = %v, want solo", config.NetworkType)
	}
	if config.APIAddr != "localhost:8669" {
		t.Errorf("Config.APIAddr = %v, want localhost:8669", config.APIAddr)
	}
	if config.P2PPort != 11235 {
		t.Errorf("Config.P2PPort = %v, want 11235", config.P2PPort)
	}
	if !config.OnDemand {
		t.Errorf("Config.OnDemand = %v, want true", config.OnDemand)
	}
	if config.Persist {
		t.Errorf("Config.Persist = %v, want false", config.Persist)
	}
	if config.APICORS != "*" {
		t.Errorf("Config.APICORS = %v, want *", config.APICORS)
	}
}
