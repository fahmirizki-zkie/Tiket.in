package middleware

import (
	"github.com/gofiber/fiber/v2"

	"tiket-in-backend/helper"
)

func RequireRole(allowedRoles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		role, ok := c.Locals("userRole").(string)
		if !ok || role == "" {
			return helper.Unauthorized("sesi tidak valid")
		}

		for _, r := range allowedRoles {
			if r == role {
				return c.Next()
			}
		}

		return helper.Forbidden("anda tidak memiliki akses ke resource ini")
	}
}
