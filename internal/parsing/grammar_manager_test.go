package parsing_test

import (
	"testing"

	"github.com/aaamil13/CodeIndexerMCP/internal/parsing"
	"github.com/stretchr/testify/assert"
)

func TestGrammarManager_GetLanguage(t *testing.T) {
	gm := parsing.NewGrammarManager()

	tests := []struct {
		lang     string
		expected bool
	}{
		{"go", true},
		{"python", true},
		{"typescript", true},
		{"javascript", true},
		{"java", true},
		{"c", true},
		{"cpp", true},
		{"rust", true},
		{"unknown", false},
	}

	for _, test := range tests {
		t.Run(test.lang, func(t *testing.T) {
			lang, err := gm.GetLanguage(test.lang)
			if test.expected {
				assert.NoError(t, err)
				assert.NotNil(t, lang)
			} else {
				assert.Error(t, err)
				assert.Nil(t, lang)
			}
		})
	}
}

func TestGrammarManager_GetSupportedLanguages(t *testing.T) {
	gm := parsing.NewGrammarManager()
	supported := gm.GetSupportedLanguages()

	expectedLanguages := []string{"go", "python", "typescript", "javascript", "java", "c", "cpp", "rust"}
	assert.ElementsMatch(t, expectedLanguages, supported)
}
