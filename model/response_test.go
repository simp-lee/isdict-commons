package model

import "testing"

func TestNewSuccessResponse(t *testing.T) {
	data := map[string]string{"key": "value"}
	resp := NewSuccessResponse(data)

	if !resp.Success {
		t.Error("Expected Success to be true")
	}
	if resp.Data == nil {
		t.Error("Expected Data to be non-nil")
	}
	if resp.Error != nil {
		t.Error("Expected Error to be nil")
	}
}

func TestNewErrorResponse(t *testing.T) {
	code := "TEST_ERROR"
	message := "Test error message"
	resp := NewErrorResponse(code, message, nil)

	if resp.Success {
		t.Error("Expected Success to be false")
	}
	if resp.Data != nil {
		t.Error("Expected Data to be nil")
	}
	if resp.Error == nil {
		t.Error("Expected Error to be non-nil")
	}
	if resp.Error.Code != code {
		t.Errorf("Expected Error.Code to be %s, got %s", code, resp.Error.Code)
	}
	if resp.Error.Message != message {
		t.Errorf("Expected Error.Message to be %s, got %s", message, resp.Error.Message)
	}
}

func TestNewSuccessResponseWithMeta(t *testing.T) {
	data := []string{"item1", "item2"}
	total := int64(100)
	limit := 20
	meta := &MetaInfo{
		Total: &total,
		Limit: &limit,
	}

	resp := NewSuccessResponseWithMeta(data, meta)

	if !resp.Success {
		t.Error("Expected Success to be true")
	}
	if resp.Meta == nil {
		t.Error("Expected Meta to be non-nil")
	}
	if resp.Meta.Total == nil || *resp.Meta.Total != total {
		t.Errorf("Expected Meta.Total to be %d", total)
	}
}
