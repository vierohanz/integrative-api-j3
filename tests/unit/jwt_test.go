package unit

import (
	"gofiber-starterkit/pkg/utils"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestJWT_Unit(t *testing.T) {
	userID := uuid.New()
	t.Log("==================================================")
	t.Logf(" TEST SUITE: JWT Security Utility")
	t.Logf(" USER ID   : %v", userID)
	t.Log("==================================================")

	t.Run("Generate and Validate JWT", func(t *testing.T) {
		t.Log("[STEP 1] Generating secure JWT token...")
		token, err := utils.GenerateJWT(userID)
		
		assert.NoError(t, err, "Failed to generate JWT")
		assert.NotEmpty(t, token, "Generated token is empty")
		t.Log("  >> Result: SUCCESS")

		t.Log("[STEP 2] Validating token claims...")
		claims, err := utils.ValidateJWT(token)
		
		assert.NoError(t, err, "Failed to validate JWT")
		assert.Equal(t, userID, claims.UserID, "UserID mismatch in claims")
		t.Log("  >> Result: VALIDATED")
	})

	t.Run("Validate Invalid Token", func(t *testing.T) {
		t.Log("[STEP 3] Verifying security rejection for invalid token...")
		_, err := utils.ValidateJWT("invalid.token.string")
		
		assert.Error(t, err, "Security flaw: invalid token was accepted")
		t.Log("  >> Result: REJECTED (Safe)")
	})
	
	t.Log("--------------------------------------------------")
}
