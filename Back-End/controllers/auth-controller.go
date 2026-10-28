package controllers

import (
	"TaleHaven-Backend/middlewares"
	"TaleHaven-Backend/modules/auth"
	"TaleHaven-Backend/modules/database"
	"TaleHaven-Backend/modules/sessions"
	"TaleHaven-Backend/modules/utils"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// AuthController sets up authentication routes
func AuthController(router *gin.Engine) {
	authRoutes := router.Group("/auth")
	authRoutes.Use(middlewares.RequireTrustedOrigin())
	{
		authRoutes.POST("/register", registerHandler)
		authRoutes.POST("/login", loginHandler)
		authRoutes.POST("/logout", logoutHandler)

		// Protected routes requiring authentication
		protected := authRoutes.Group("")
		protected.Use(middlewares.AuthMiddleware())
		{
			protected.GET("/me", getCurrentUserHandler)
			protected.PUT("/password", changePasswordHandler)
			protected.DELETE("/account", deleteAccountHandler)
		}
	}
}

// RegisterRequest represents the registration request payload
type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

// LoginRequest represents the login request payload
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// ChangePasswordRequest represents the change password request payload
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

// UserResponse represents the user data returned in API responses
type UserResponse struct {
	ID                            int      `json:"id"`
	Email                         string   `json:"email"`
	Role                          string   `json:"role"`
	DisplayName                   string   `json:"displayName"`
	ProfileImagePath              string   `json:"profileImagePath"`
	DisplayTheme                  string   `json:"displayTheme"`
	CustomThemeColor              string   `json:"customThemeColor"`
	ThemeImagePath                string   `json:"themeImagePath"`
	SiteTheme                     string   `json:"siteTheme"`
	SiteCustomColor               string   `json:"siteCustomColor"`
	SiteThemeImagePaths           []string `json:"siteThemeImagePaths"`
	SiteThemeImageRotationSeconds int      `json:"siteThemeImageRotationSeconds"`
	SiteThemeImageRotationMode    string   `json:"siteThemeImageRotationMode"`
	SiteThemeImageRotationMinutes int      `json:"siteThemeImageRotationMinutes"`
	CreatedAt                     string   `json:"created_at"`
}

// toUserResponse converts a User model to UserResponse
func toUserResponse(user *database.User) UserResponse {
	siteThemeImagePaths := []string(user.SiteThemeImagePaths)
	if siteThemeImagePaths == nil {
		siteThemeImagePaths = []string{}
	}
	return UserResponse{
		ID:                            user.ID,
		Email:                         user.Email,
		Role:                          user.Role,
		DisplayName:                   user.DisplayName,
		ProfileImagePath:              user.ProfileImagePath,
		DisplayTheme:                  user.DisplayTheme,
		CustomThemeColor:              user.CustomThemeColor,
		ThemeImagePath:                user.ThemeImagePath,
		SiteTheme:                     user.SiteTheme,
		SiteCustomColor:               user.SiteCustomColor,
		SiteThemeImagePaths:           siteThemeImagePaths,
		SiteThemeImageRotationSeconds: user.SiteThemeImageRotationSeconds,
		SiteThemeImageRotationMode:    user.SiteThemeImageRotationMode,
		SiteThemeImageRotationMinutes: user.SiteThemeImageRotationMinutes,
		CreatedAt:                     user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

// registerHandler handles user registration
// POST /auth/register
func registerHandler(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Logger.Warn("Invalid registration request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request",
			"message": err.Error(),
		})
		return
	}

	// Normalize email to lowercase
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	// Create user
	user, err := auth.CreateUser(req.Email, req.Password)
	if err != nil {
		if err == auth.ErrEmailAlreadyExists {
			c.JSON(http.StatusConflict, gin.H{
				"error":   "Email already exists",
				"message": "An account with this email already exists",
			})
			return
		}

		utils.Logger.Error("Failed to create user", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Registration failed",
			"message": "Failed to create account. Please try again later",
		})
		return
	}

	// Create a revocable server-side session and set only its opaque token in
	// the HttpOnly cookie.
	token, err := sessions.CreateAuthenticatedSession(user.ID, c.ClientIP(), c.Request.UserAgent())
	if err != nil {
		utils.Logger.Error("Failed to create session", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Registration failed",
			"message": "Failed to create session. Please try again later",
		})
		return
	}

	middlewares.SetAuthCookie(c, token)

	utils.Logger.Info("User registered successfully", zap.Int("user_id", user.ID), zap.String("email", user.Email))

	c.JSON(http.StatusCreated, gin.H{
		"message": "Account created successfully",
		"user":    toUserResponse(user),
	})
}

// loginHandler handles user login
// POST /auth/login
func loginHandler(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Logger.Warn("Invalid login request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request",
			"message": err.Error(),
		})
		return
	}

	// Normalize email to lowercase
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	// Authenticate user
	user, err := auth.AuthenticateUser(req.Email, req.Password)
	if err != nil {
		if err == auth.ErrInvalidCredentials {
			utils.Logger.Info("Failed login attempt", zap.String("email", req.Email))
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Invalid credentials",
				"message": "Invalid email or password",
			})
			return
		}

		utils.Logger.Error("Authentication error", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Authentication failed",
			"message": "An error occurred during login. Please try again later",
		})
		return
	}

	token, err := sessions.CreateAuthenticatedSession(user.ID, c.ClientIP(), c.Request.UserAgent())
	if err != nil {
		utils.Logger.Error("Failed to create session", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Login failed",
			"message": "Failed to create session. Please try again later",
		})
		return
	}

	middlewares.SetAuthCookie(c, token)

	utils.Logger.Info("User logged in successfully",
		zap.Int("user_id", user.ID),
		zap.String("email", user.Email))

	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
		"user":    toUserResponse(user),
	})
}

// logoutHandler handles user logout
// POST /auth/logout
func logoutHandler(c *gin.Context) {
	if token, err := c.Cookie(middlewares.AuthCookieName); err == nil {
		if revokeErr := sessions.RevokeAuthenticatedSession(token); revokeErr != nil {
			utils.Logger.Error("Failed to revoke session during logout", zap.Error(revokeErr))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Logout failed", "message": "Unable to end the session. Please try again."})
			return
		}
	}
	middlewares.ClearAuthCookie(c)

	utils.Logger.Info("User logged out")

	c.JSON(http.StatusOK, gin.H{
		"message": "Logged out successfully",
	})
}

// getCurrentUserHandler returns the currently authenticated user's data
// GET /auth/me
func getCurrentUserHandler(c *gin.Context) {
	userData, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated", "message": "You must be logged in to access this resource"})
		return
	}

	user, ok := userData.(*database.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Invalid user data",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": toUserResponse(user),
	})
}

// changePasswordHandler handles password changes for authenticated users
// PUT /auth/password
func changePasswordHandler(c *gin.Context) {
	// Get authenticated user
	userData, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Authentication required",
			"message": "You must be logged in to change your password",
		})
		return
	}

	user, ok := userData.(*database.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Invalid user data",
		})
		return
	}

	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Logger.Warn("Invalid change password request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request",
			"message": err.Error(),
		})
		return
	}

	// Verify old password
	if !auth.ValidatePassword(user.PasswordHash, req.OldPassword) {
		utils.Logger.Info("Failed password change attempt - incorrect old password",
			zap.Int("user_id", user.ID))

		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Invalid password",
			"message": "Current password is incorrect",
		})
		return
	}

	// Update password
	err := auth.UpdateUserPassword(user.ID, req.NewPassword)
	if err != nil {
		utils.Logger.Error("Failed to update password", zap.Error(err), zap.Int("user_id", user.ID))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Password change failed",
			"message": "Failed to update password. Please try again later",
		})
		return
	}
	if err := sessions.RevokeAllAuthenticatedSessions(user.ID); err != nil {
		utils.Logger.Error("Failed to revoke sessions after password change", zap.Error(err), zap.Int("user_id", user.ID))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Password change failed", "message": "Password changed but sessions could not be secured. Please try again."})
		return
	}
	middlewares.ClearAuthCookie(c)

	utils.Logger.Info("Password changed successfully", zap.Int("user_id", user.ID))

	c.JSON(http.StatusOK, gin.H{
		"message": "Password changed successfully",
	})
}

// deleteAccountHandler handles account deletion for authenticated users
// DELETE /auth/account
func deleteAccountHandler(c *gin.Context) {
	// Get authenticated user
	userData, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Authentication required",
			"message": "You must be logged in to delete your account",
		})
		return
	}

	user, ok := userData.(*database.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Invalid user data",
		})
		return
	}

	// Delete the user (cascades to sessions)
	err := auth.DeleteUser(user.ID)
	if err != nil {
		utils.Logger.Error("Failed to delete user account", zap.Error(err), zap.Int("user_id", user.ID))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Account deletion failed",
			"message": "Failed to delete account. Please try again later",
		})
		return
	}
	removeStoredProfilePhoto(user.ProfileImagePath)
	removeStoredProfileThemeImage(user.ThemeImagePath)
	middlewares.ClearAuthCookie(c)

	utils.Logger.Info("User account deleted", zap.Int("user_id", user.ID), zap.String("email", user.Email))

	c.JSON(http.StatusOK, gin.H{
		"message": "Account deleted successfully",
	})
}
