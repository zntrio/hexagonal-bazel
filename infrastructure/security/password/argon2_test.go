package password

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Argon2_Encode(t *testing.T) {
	s := Argon2()

	t.Run("valid secret", func(t *testing.T) {
		passwordHash, err := s.Hash("test")
		assert.NoError(t, err)
		assert.NotEmpty(t, passwordHash)

		err = s.Verify("test", passwordHash)
		assert.NoError(t, err)
	})

	t.Run("invalid secret", func(t *testing.T) {
		passwordHash, err := s.Hash("test")
		assert.NoError(t, err)
		assert.NotEmpty(t, passwordHash)

		err = s.Verify("testify", passwordHash)
		assert.Error(t, err)
	})
}
