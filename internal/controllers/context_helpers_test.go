package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetGroupIDFromContext(t *testing.T) {
	// Create a new gin context for each test case
	gin.SetMode(gin.TestMode)

	t.Run("Successfully retrieves valid group ID", func(t *testing.T) {
		// Create a new context and set a valid UUID
		c, _ := gin.CreateTestContext(nil)
		expectedID := uuid.New()
		c.Set("groupID", expectedID.String())

		// Call the function and check the result
		resultID, err := getGroupIDFromContext(c)

		assert.NoError(t, err)
		assert.Equal(t, expectedID, resultID)
	})

	t.Run("Returns error when group ID is missing", func(t *testing.T) {
		// Create a new context without setting any values
		c, _ := gin.CreateTestContext(nil)

		// Call the function and verify it returns the expected error
		resultID, err := getGroupIDFromContext(c)

		assert.Error(t, err)
		assert.Equal(t, "group ID not found in context", err.Error())
		assert.Equal(t, uuid.Nil, resultID)
	})

	t.Run("Returns error when group ID is invalid", func(t *testing.T) {
		// Create a new context and set an invalid UUID
		c, _ := gin.CreateTestContext(nil)
		c.Set("groupID", "invalid-uuid")

		// Call the function and verify it returns the expected error
		resultID, err := getGroupIDFromContext(c)

		assert.Error(t, err)
		assert.Equal(t, "invalid group ID", err.Error())
		assert.Equal(t, uuid.Nil, resultID)
	})
}

func TestGetUserIDFromContext(t *testing.T) {
	// We'll create a new gin context for each test case
	gin.SetMode(gin.TestMode)

	t.Run("Successfully retrieves valid user ID", func(t *testing.T) {
		// Create a new context and set a valid UUID
		c, _ := gin.CreateTestContext(nil)
		expectedID := uuid.New()
		c.Set("userID", expectedID.String())

		// Call the function and check the result
		resultID, err := getUserIDFromContext(c)

		assert.NoError(t, err)
		assert.Equal(t, expectedID, resultID)
	})

	t.Run("Returns error when user ID is missing", func(t *testing.T) {
		// Create a new context without setting any values
		c, _ := gin.CreateTestContext(nil)

		// Call the function and verify it returns the expected error
		resultID, err := getUserIDFromContext(c)

		assert.Error(t, err)
		assert.Equal(t, "user ID not found in context", err.Error())
		assert.Equal(t, uuid.Nil, resultID)
	})

	t.Run("Returns error when user ID is invalid", func(t *testing.T) {
		// Create a new context and set an invalid UUID
		c, _ := gin.CreateTestContext(nil)
		c.Set("userID", "invalid-uuid")

		// Call the function and verify it returns the expected error
		resultID, err := getUserIDFromContext(c)

		assert.Error(t, err)
		assert.Equal(t, "invalid user ID", err.Error())
		assert.Equal(t, uuid.Nil, resultID)
	})
}
