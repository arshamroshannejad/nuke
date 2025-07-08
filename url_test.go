package nuke

import (
	"net/http/httptest"
	"testing"
)

func TestURLParam(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		param    string
		expected string
	}{
		{
			name:     "existing parameter",
			path:     "/users/{id}",
			param:    "id",
			expected: "123",
		},
		{
			name:     "non-existent parameter",
			path:     "/users/{id}",
			param:    "name",
			expected: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/users/123", nil)
			req.SetPathValue(tt.param, tt.expected)
			got := URLParam(req, tt.param)
			if got != tt.expected {
				t.Errorf("URLParam() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestQueryParam(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		param    string
		expected string
	}{
		{
			name:     "existing parameter",
			url:      "/search?q=golang",
			param:    "q",
			expected: "golang",
		},
		{
			name:     "non-existent parameter",
			url:      "/search?q=golang",
			param:    "page",
			expected: "",
		},
		{
			name:     "multiple values",
			url:      "/search?q=go&q=lang",
			param:    "q",
			expected: "go",
		},
		{
			name:     "empty value",
			url:      "/search?q=",
			param:    "q",
			expected: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.url, nil)
			got := QueryParam(req, tt.param)
			if got != tt.expected {
				t.Errorf("QueryParam() = %v, want %v", got, tt.expected)
			}
		})
	}
}
