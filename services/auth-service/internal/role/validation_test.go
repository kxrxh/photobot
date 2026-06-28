package role

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateRoleName(t *testing.T) {
	t.Run("valid role names", func(t *testing.T) {
		valid := []string{"admin", "user", "role_1", "my-role", "a1", "Role123"}
		for _, name := range valid {
			err := ValidateRoleName(name)
			assert.NoError(t, err, "ValidateRoleName(%q) should succeed", name)
		}
	})

	t.Run("too short", func(t *testing.T) {
		err := ValidateRoleName("a")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "at least 2 characters")
	})

	t.Run("empty", func(t *testing.T) {
		err := ValidateRoleName("")
		assert.Error(t, err)
	})

	t.Run("too long", func(t *testing.T) {
		long := strings.Repeat("a", 51)
		err := ValidateRoleName(long)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot exceed 50")
	})

	t.Run("invalid characters", func(t *testing.T) {
		invalid := []string{"role name", "role@admin", "role.name", "role!"}
		for _, name := range invalid {
			err := ValidateRoleName(name)
			assert.Error(t, err, "ValidateRoleName(%q) should fail", name)
			assert.Contains(t, err.Error(), "letters, numbers")
		}
	})
}
