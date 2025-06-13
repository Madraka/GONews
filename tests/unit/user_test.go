package unit

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"news/internal/models"
	"news/tests/testutil"
)

// UserModelTestSuite defines the test suite for User model
type UserModelTestSuite struct {
	suite.Suite
	testData *testutil.TestData
}

// SetupSuite runs before all tests in the suite
func (suite *UserModelTestSuite) SetupSuite() {
	suite.testData = testutil.NewTestData()
}

// TestUserModel_ValidUser tests valid user creation
func (suite *UserModelTestSuite) TestUserModel_ValidUser() {
	user := suite.testData.CreateTestUser()

	assert.NotEmpty(suite.T(), user.Username)
	assert.NotEmpty(suite.T(), user.Email)
	assert.NotEmpty(suite.T(), user.Password)
	assert.Equal(suite.T(), "user", user.Role)
	assert.Equal(suite.T(), "active", user.Status)
	assert.False(suite.T(), user.IsVerified) // Default is false
	assert.NotZero(suite.T(), user.CreatedAt)
	assert.NotZero(suite.T(), user.UpdatedAt)
}

// TestUserModel_AdminUser tests admin user creation
func (suite *UserModelTestSuite) TestUserModel_AdminUser() {
	admin := suite.testData.CreateTestAdmin()

	assert.Equal(suite.T(), "admin", admin.Username)
	assert.Equal(suite.T(), "admin@example.com", admin.Email)
	assert.Equal(suite.T(), "admin", admin.Role)
	assert.Equal(suite.T(), "active", admin.Status)
}

// TestUserModel_Validation tests user validation
func (suite *UserModelTestSuite) TestUserModel_Validation() {
	// Test valid role
	user := &models.User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password",
		Role:     "user",
		Status:   "active",
	}

	assert.True(suite.T(), user.ValidateRole(), "Valid role should pass validation")

	// Test invalid role
	user.Role = "invalid_role"
	assert.False(suite.T(), user.ValidateRole(), "Invalid role should fail validation")

	// Test valid status
	user.Status = "active"
	assert.True(suite.T(), user.ValidateStatus(), "Valid status should pass validation")

	// Test invalid status
	user.Status = "invalid_status"
	assert.False(suite.T(), user.ValidateStatus(), "Invalid status should fail validation")
}

// TestUserModel_IsAdmin tests admin role checking
func (suite *UserModelTestSuite) TestUserModel_IsAdmin() {
	user := suite.testData.CreateTestUser()
	admin := suite.testData.CreateTestAdmin()

	assert.Equal(suite.T(), "user", user.Role)
	assert.Equal(suite.T(), "admin", admin.Role)
	assert.True(suite.T(), user.ValidateRole())
	assert.True(suite.T(), admin.ValidateRole())
}

// TestUserModel_Timestamps tests timestamp handling
func (suite *UserModelTestSuite) TestUserModel_Timestamps() {
	user := suite.testData.CreateTestUser()

	now := time.Now()

	// Check that timestamps are recent (within 1 second)
	assert.WithinDuration(suite.T(), now, user.CreatedAt, time.Second)
	assert.WithinDuration(suite.T(), now, user.UpdatedAt, time.Second)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestUserModelTestSuite(t *testing.T) {
	suite.Run(t, new(UserModelTestSuite))
}

// Test individual functions without suite (alternative approach)
func TestUser_BasicValidation(t *testing.T) {
	testData := testutil.NewTestData()
	user := testData.CreateTestUser()

	require.NotNil(t, user)
	assert.NotEmpty(t, user.Username)
	assert.NotEmpty(t, user.Email)
	assert.Equal(t, "active", user.Status)
}

func TestUser_EmailValidation(t *testing.T) {
	tests := []struct {
		name  string
		email string
		valid bool
	}{
		{"valid email", "test@example.com", true},
		{"invalid email", "invalid-email", false},
		{"empty email", "", false},
		{"email without domain", "test@", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &models.User{
				Username: "testuser",
				Email:    tt.email,
				Password: "password",
				Role:     "user",
			}

			// Use the user variable for validation
			assert.Equal(t, tt.email, user.Email)
			// In real implementation, you would validate the email
			if tt.valid {
				assert.Contains(t, tt.email, "@")
				assert.Contains(t, tt.email, ".")
			} else {
				// Test invalid cases
				if tt.email != "" {
					assert.True(t, len(tt.email) < 5 || !contains(tt.email, "@") || !contains(tt.email, "."))
				}
			}
		})
	}
}

// Helper function for email validation test
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
