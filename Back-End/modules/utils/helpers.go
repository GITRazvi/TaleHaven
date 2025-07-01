package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func HashText(text string) string {
	// Create a new SHA-256 hash
	hash := sha256.New()

	// Write the text as bytes to the hash
	hash.Write([]byte(text))

	// Get the final hash as a byte slice
	hashedBytes := hash.Sum(nil)

	// Convert the byte slice to a hexadecimal string
	hashedText := hex.EncodeToString(hashedBytes)

	return hashedText
}

func LoadEnv() {
	err := godotenv.Load(".env")
	if err != nil {
		Logger.Fatal("Error loading .env file")
	}
}

func AddLogsToContext(context *gin.Context, key string, value any) {
	if data, exists := context.Get("contextLogs"); exists {
		if contextData, ok := data.(map[string]any); ok {
			contextData[key] = value
		}
	}
}
