package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"

	"tiket-in-backend/helper"
)

func AuthRequired(jwtMgr *helper.JWTManager) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return helper.Unauthorized("token tidak ditemukan")
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			return helper.Unauthorized("format token tidak valid, gunakan: Bearer <token>")
		}

		claims, err := jwtMgr.VerifyToken(parts[1])
		if err != nil {
			return helper.Unauthorized("token tidak valid atau sudah kedaluwarsa")
		}

		c.Locals("userID", claims.UserID)
		c.Locals("userRole", claims.Role)

		return c.Next()
	}
}
