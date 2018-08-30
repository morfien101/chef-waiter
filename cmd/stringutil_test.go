package cmd

import (
	"testing"
)

func TestChomp(t *testing.T) {
	tests := []struct {
		name          string
		testString    string
		expectedValue string
	}{
		{
			name:          "no chomp",
			testString:    "testValue",
			expectedValue: "testValue",
		},
		{
			name:          "yes chomp",
			testString:    "testValue\n",
			expectedValue: "testValue",
		},
		{
			name:          "yes many chomp",
			testString:    "testValue\n\n\n",
			expectedValue: "testValue",
		},
	}

	for _, test := range tests {
		if Chomp(test.testString) != test.expectedValue {
			t.Errorf("Test %s: Chomp failed. Got %s - expected: %s", test.name, test.testString, test.expectedValue)
		}
	}
}
