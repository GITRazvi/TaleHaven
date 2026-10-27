package auth

import (
	"TaleHaven-Backend/modules/database"
	"TaleHaven-Backend/modules/utils"
	"database/sql"
	"errors"
	"strings"
	"time"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrEmailAlreadyExists = errors.New("email already exists")
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrUserNotFound       = errors.New("user not found")
)

// CreateUser creates a new user with hashed password
func CreateUser(email, password string) (*database.User, error) {
	// Check if user already exists
	existingUser, err := GetUserByEmail(email)
	if err != nil && err != sql.ErrNoRows {
		utils.Logger.Error("Failed to check existing user", zap.Error(err))
		return nil, err
	}
	if existingUser != nil {
		return nil, ErrEmailAlreadyExists
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		utils.Logger.Error("Failed to hash password", zap.Error(err))
		return nil, err
	}

	// Insert user into database
	query := `
		INSERT INTO users (email, password_hash, display_name, created_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id, email, role, display_name, profile_image_path, display_theme, custom_theme_color, theme_image_path, created_at
	`

	user := &database.User{}
	displayName := strings.SplitN(email, "@", 2)[0]
	err = database.DB.QueryRow(query, email, string(hashedPassword), displayName, time.Now()).Scan(
		&user.ID,
		&user.Email,
		&user.Role,
		&user.DisplayName,
		&user.ProfileImagePath,
		&user.DisplayTheme,
		&user.CustomThemeColor,
		&user.ThemeImagePath,
		&user.CreatedAt,
	)

	if err != nil {
		utils.Logger.Error("Failed to create user", zap.Error(err), zap.String("email", email))
		return nil, err
	}

	utils.Logger.Info("User created successfully", zap.Int("user_id", user.ID), zap.String("email", email))
	return user, nil
}

// GetUserByEmail retrieves a user by email
func GetUserByEmail(email string) (*database.User, error) {
	query := `
		SELECT id, email, role, display_name, profile_image_path, display_theme, custom_theme_color, theme_image_path, password_hash, created_at
		FROM users
		WHERE email = $1
	`

	user := &database.User{}
	err := database.DB.QueryRow(query, email).Scan(
		&user.ID,
		&user.Email,
		&user.Role,
		&user.DisplayName,
		&user.ProfileImagePath,
		&user.DisplayTheme,
		&user.CustomThemeColor,
		&user.ThemeImagePath,
		&user.PasswordHash,
		&user.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		utils.Logger.Error("Failed to get user by email", zap.Error(err), zap.String("email", email))
		return nil, err
	}

	return user, nil
}

// GetUserByID retrieves a user by ID
func GetUserByID(userID int) (*database.User, error) {
	query := `
		SELECT id, email, role, display_name, profile_image_path, display_theme, custom_theme_color, theme_image_path, password_hash, created_at
		FROM users
		WHERE id = $1
	`

	user := &database.User{}
	err := database.DB.QueryRow(query, userID).Scan(
		&user.ID,
		&user.Email,
		&user.Role,
		&user.DisplayName,
		&user.ProfileImagePath,
		&user.DisplayTheme,
		&user.CustomThemeColor,
		&user.ThemeImagePath,
		&user.PasswordHash,
		&user.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}
	if err != nil {
		utils.Logger.Error("Failed to get user by ID", zap.Error(err), zap.Int("user_id", userID))
		return nil, err
	}

	return user, nil
}

// UpdateUserDisplayName changes only the authenticated user's presentation
// name and returns safe profile fields for the response.
func UpdateUserDisplayName(userID int, displayName string) (*database.User, error) {
	user := &database.User{}
	err := database.DB.QueryRow(`
		UPDATE users SET display_name = $1 WHERE id = $2
		RETURNING id, email, role, display_name, profile_image_path, display_theme, custom_theme_color, theme_image_path, created_at
	`, displayName, userID).Scan(
		&user.ID, &user.Email, &user.Role, &user.DisplayName, &user.ProfileImagePath, &user.DisplayTheme, &user.CustomThemeColor, &user.ThemeImagePath, &user.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return user, nil
}

// UpdateUserProfileImagePath records only a server-owned image path.
func UpdateUserProfileImagePath(userID int, imagePath string) (*database.User, error) {
	user := &database.User{}
	err := database.DB.QueryRow(`
		UPDATE users SET profile_image_path = $1 WHERE id = $2
		RETURNING id, email, role, display_name, profile_image_path, display_theme, custom_theme_color, theme_image_path, created_at
	`, imagePath, userID).Scan(
		&user.ID, &user.Email, &user.Role, &user.DisplayName, &user.ProfileImagePath, &user.DisplayTheme, &user.CustomThemeColor, &user.ThemeImagePath, &user.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return user, nil
}

// UpdateUserDisplayTheme records a validated, named profile palette for the
// authenticated user's own profile presentation.
func UpdateUserDisplayTheme(userID int, displayTheme, customThemeColor string) (*database.User, error) {
	user := &database.User{}
	err := database.DB.QueryRow(`
		UPDATE users SET display_theme = $1, custom_theme_color = $2 WHERE id = $3
		RETURNING id, email, role, display_name, profile_image_path, display_theme, custom_theme_color, theme_image_path, created_at
	`, displayTheme, customThemeColor, userID).Scan(
		&user.ID, &user.Email, &user.Role, &user.DisplayName, &user.ProfileImagePath, &user.DisplayTheme, &user.CustomThemeColor, &user.ThemeImagePath, &user.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return user, nil
}

// UpdateUserThemeImagePath records only a server-generated presentation
// background path for the authenticated user's profile.
func UpdateUserThemeImagePath(userID int, imagePath string) (*database.User, error) {
	user := &database.User{}
	err := database.DB.QueryRow(`
		UPDATE users SET theme_image_path = $1 WHERE id = $2
		RETURNING id, email, role, display_name, profile_image_path, display_theme, custom_theme_color, theme_image_path, created_at
	`, imagePath, userID).Scan(
		&user.ID, &user.Email, &user.Role, &user.DisplayName, &user.ProfileImagePath, &user.DisplayTheme, &user.CustomThemeColor, &user.ThemeImagePath, &user.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return user, nil
}

// ValidatePassword checks if the provided password matches the stored hash
func ValidatePassword(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

// AuthenticateUser validates user credentials and returns the user if valid
func AuthenticateUser(email, password string) (*database.User, error) {
	user, err := GetUserByEmail(email)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrInvalidCredentials
	}

	if !ValidatePassword(user.PasswordHash, password) {
		return nil, ErrInvalidCredentials
	}

	utils.Logger.Info("User authenticated successfully", zap.Int("user_id", user.ID), zap.String("email", email))
	return user, nil
}

// UpdateUserPassword updates a user's password
func UpdateUserPassword(userID int, newPassword string) error {
	// Hash the new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		utils.Logger.Error("Failed to hash new password", zap.Error(err))
		return err
	}

	query := `UPDATE users SET password_hash = $1 WHERE id = $2`

	_, err = database.DB.Exec(query, string(hashedPassword), userID)
	if err != nil {
		utils.Logger.Error("Failed to update user password", zap.Error(err), zap.Int("user_id", userID))
		return err
	}

	utils.Logger.Info("User password updated successfully", zap.Int("user_id", userID))
	return nil
}

// DeleteUser deletes a user and all associated data (cascades to sessions)
func DeleteUser(userID int) error {
	query := `DELETE FROM users WHERE id = $1`

	result, err := database.DB.Exec(query, userID)
	if err != nil {
		utils.Logger.Error("Failed to delete user", zap.Error(err), zap.Int("user_id", userID))
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return ErrUserNotFound
	}

	utils.Logger.Info("User deleted successfully", zap.Int("user_id", userID))
	return nil
}
