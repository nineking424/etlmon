package models

import (
	"encoding/json"
	"testing"
)

func TestResponse_JSONMarshaling(t *testing.T) {
	response := Response{
		Data: map[string]string{"key": "value"},
		Meta: &Meta{
			Total:  100,
			Limit:  10,
			Offset: 0,
		},
	}

	// Marshal to JSON
	data, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	// Unmarshal back
	var decoded Response
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	// Verify Meta
	if decoded.Meta == nil {
		t.Fatal("Meta is nil")
	}
	if decoded.Meta.Total != response.Meta.Total {
		t.Errorf("Meta.Total: got %d, want %d", decoded.Meta.Total, response.Meta.Total)
	}
	if decoded.Meta.Limit != response.Meta.Limit {
		t.Errorf("Meta.Limit: got %d, want %d", decoded.Meta.Limit, response.Meta.Limit)
	}
	if decoded.Meta.Offset != response.Meta.Offset {
		t.Errorf("Meta.Offset: got %d, want %d", decoded.Meta.Offset, response.Meta.Offset)
	}
}

func TestResponse_WithoutMeta(t *testing.T) {
	response := Response{
		Data: []string{"item1", "item2"},
		Meta: nil,
	}

	data, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	// Verify Meta is omitted when nil
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("Failed to unmarshal to map: %v", err)
	}

	if _, exists := raw["meta"]; exists {
		t.Error("meta field should be omitted when nil")
	}

	// Should still have data field
	if _, exists := raw["data"]; !exists {
		t.Error("data field is missing")
	}
}

func TestMeta_JSONMarshaling(t *testing.T) {
	meta := Meta{
		Total:  250,
		Limit:  25,
		Offset: 50,
	}

	data, err := json.Marshal(meta)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	var decoded Meta
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if decoded.Total != meta.Total {
		t.Errorf("Total: got %d, want %d", decoded.Total, meta.Total)
	}
	if decoded.Limit != meta.Limit {
		t.Errorf("Limit: got %d, want %d", decoded.Limit, meta.Limit)
	}
	if decoded.Offset != meta.Offset {
		t.Errorf("Offset: got %d, want %d", decoded.Offset, meta.Offset)
	}
}

func TestMeta_OmitEmpty(t *testing.T) {
	meta := Meta{
		Total: 0,
		Limit: 0,
		Offset: 0,
	}

	data, err := json.Marshal(meta)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	// With omitempty, zero values should be omitted
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("Failed to unmarshal to map: %v", err)
	}

	if _, exists := raw["total"]; exists {
		t.Error("total field should be omitted when zero")
	}
	if _, exists := raw["limit"]; exists {
		t.Error("limit field should be omitted when zero")
	}
	if _, exists := raw["offset"]; exists {
		t.Error("offset field should be omitted when zero")
	}
}

func TestErrorResponse_JSONMarshaling(t *testing.T) {
	errResp := ErrorResponse{
		Error:   "resource not found",
		Code:    "NOT_FOUND",
		Details: "The requested resource ID does not exist",
	}

	data, err := json.Marshal(errResp)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	var decoded ErrorResponse
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if decoded.Error != errResp.Error {
		t.Errorf("Error: got %s, want %s", decoded.Error, errResp.Error)
	}
	if decoded.Code != errResp.Code {
		t.Errorf("Code: got %s, want %s", decoded.Code, errResp.Code)
	}
	if decoded.Details != errResp.Details {
		t.Errorf("Details: got %s, want %s", decoded.Details, errResp.Details)
	}
}

func TestErrorResponse_MinimalError(t *testing.T) {
	errResp := ErrorResponse{
		Error: "internal server error",
	}

	data, err := json.Marshal(errResp)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("Failed to unmarshal to map: %v", err)
	}

	// Error field is required
	if _, exists := raw["error"]; !exists {
		t.Error("error field is missing")
	}

	// Code and Details should be omitted when empty
	if _, exists := raw["code"]; exists {
		t.Error("code field should be omitted when empty")
	}
	if _, exists := raw["details"]; exists {
		t.Error("details field should be omitted when empty")
	}
}

func TestErrorResponse_JSONTags(t *testing.T) {
	errResp := ErrorResponse{
		Error:   "validation failed",
		Code:    "VALIDATION_ERROR",
		Details: "field 'name' is required",
	}

	data, err := json.Marshal(errResp)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("Failed to unmarshal to map: %v", err)
	}

	expectedFields := []string{"error", "code", "details"}
	for _, field := range expectedFields {
		if _, exists := raw[field]; !exists {
			t.Errorf("Expected JSON field %s not found", field)
		}
	}
}
