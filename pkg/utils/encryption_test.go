package utils_test

import (
	"spsyncpro_api/pkg/utils"

	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEncryptDecrypt(t *testing.T) {
	key := []byte("myverystrongpasswordo32bitlength") // replace with your key
	encryptor, err := utils.NewEncryptor(key)
	assert.NoError(t, err, "Failed to create encryptor")

	// Test case: The plaintext is "Hello, World!"
	plaintext := "Hello, World!"
	ciphertext, err := encryptor.Encrypt(plaintext)
	assert.NoError(t, err, "Failed to encrypt")

	decrypted, err := encryptor.Decrypt(ciphertext)
	assert.NoError(t, err, "Failed to decrypt")

	assert.Equal(t, plaintext, decrypted, "Decrypted text is not equal to the original text")

	encryptor2, err := utils.NewEncryptor(key)
	assert.NoError(t, err, "Failed to create encryptor")

	decrypted2, err := encryptor2.Decrypt(ciphertext)
	assert.NoError(t, err, "Failed to decrypt")

	assert.Equal(t, plaintext, decrypted2, "Decrypted text is not equal to the original text")
}
