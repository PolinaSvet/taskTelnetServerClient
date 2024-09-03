package proverbs

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_loadProverbs_Success(t *testing.T) {
	mockHTML := `<html><body><h3>Proverb 1</h3><h3>Proverb 2</h3></body></html>`
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(mockHTML))
	}
	server := httptest.NewServer(http.HandlerFunc(handler))
	defer server.Close()

	proverbsList = nil

	err := loadProverbs(server.URL)
	assert.NoError(t, err)
	assert.ElementsMatch(t, []string{"Proverb 1", "Proverb 2"}, proverbsList)
}

func Test_loadProverbs_Failure(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound) // Set status to 404 Not Found
	}
	server := httptest.NewServer(http.HandlerFunc(handler))
	defer server.Close()

	proverbsList = nil

	err := loadProverbs(server.URL)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to fetch data: 404 Not Found")
	assert.Empty(t, proverbsList)
}

func Test_loadProverbs_InvalidHTML(t *testing.T) {
	mockHTML := `<html><body><p>Not a valid HTML for proverbs</p></body></html>`
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(mockHTML))
	}
	server := httptest.NewServer(http.HandlerFunc(handler))
	defer server.Close()

	proverbsList = nil

	err := loadProverbs(server.URL)
	assert.NoError(t, err)
	assert.Empty(t, proverbsList)
}

func TestGetRandomProverb(t *testing.T) {
	tests := []struct {
		name string
		list []string
		want string
	}{
		{
			name: "GetRandomProverb_EmptyList",
			list: nil,
			want: "No proverbs available.",
		},
		{
			name: "GetRandomProverb_SingleProverb",
			list: []string{"Test proverb"},
			want: "Test proverb",
		},
		{
			name: "GetRandomProverb_MultipleProverbs",
			list: []string{"Proverb 1", "Proverb 2", "Proverb 3"},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			proverbsList = tt.list

			if tt.name == "GetRandomProverb_MultipleProverbs" {
				got := GetRandomProverb()
				assert.Contains(t, tt.list, got)
			} else {
				got := GetRandomProverb()
				assert.Equal(t, tt.want, got)
			}
		})
	}
}
