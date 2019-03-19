package internalstate

import "testing"

func TestExtractVersion(t *testing.T) {
	tests := []struct {
		testString string
		expect     string
	}{
		{
			testString: "Chef: 15.9.100\n",
			expect:     "15.9.100",
		},
		{
			testString: "chef-client - 11.1.1\r\n",
			expect:     "11.1.1",
		},
		{
			testString: "wibbleWobbled1923.321 - 12.9.41\n",
			expect:     "12.9.41",
		},
	}

	for _, test := range tests {
		if extractVersion(test.testString) != test.expect {
			t.Logf("extractVersion did not collect the expected value. Expected: %s, Got: %s", test.expect, extractVersion(test.testString))
			t.Fail()
		}
	}
}
