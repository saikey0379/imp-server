package utils

import (
	"testing"
)

func TestEncryptSymm(t *testing.T) {
	test := struct {
		key     []byte
		decrypt string
		expect  error
	}{
		key:     key_symm,
		decrypt: "d3-debuggpu-001",
		expect:  nil,
	}

	encrypt, err := EncryptSymm(test.decrypt)
	if err != test.expect {
		t.Errorf("expect requests %v actual requests %v", test.expect, err)
	}
	t.Log(encrypt)
}

func TestDecryptSymm(t *testing.T) {
	test := struct {
		key     []byte
		encrypt string
		expect  error
	}{
		key:     key_symm,
		encrypt: "ekfVQoNYTD8bVfbULcQD3llmxAg9VngSnpkhyu7vdA==",
		expect:  nil,
	}

	decrypt, err := DecryptSymm(test.encrypt)
	if err != test.expect {
		t.Errorf("expect requests %v actual requests %v", test.expect, err)
	}
	t.Log(decrypt)

}
