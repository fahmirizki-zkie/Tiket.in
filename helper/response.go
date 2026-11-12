package helper

import (
	"github.com/gofiber/fiber/v2"

	"tiket-in-backend/app/model"
)

// Success mengirim respon 200 OK dengan format WebResponse
func Success(c *fiber.Ctx, status int, message string, data any) error {
	return c.Status(status).JSON(model.WebResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// SuccessList mengirim respon 200 OK lengkap dengan pagination Meta
func SuccessList(c *fiber.Ctx, message string, data any, meta *model.Meta) error {
	return c.Status(fiber.StatusOK).JSON(model.WebResponse{
		Success: true,
		Message: message,
		Data:    data,
		Meta:    meta,
	})
}

// Created mengirim 201 Created sekaligus memasang header Location
func Created(c *fiber.Ctx, message string, data any, location string) error {
	if location != "" {
		c.Set("Location", location)
	}
	return c.Status(fiber.StatusCreated).JSON(model.WebResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// Fail mengirim respon gagal standar
func Fail(c *fiber.Ctx, status int, message string) error {
	return c.Status(status).JSON(model.WebResponse{
		Success: false,
		Message: message,
	})
}

// FailValidation mengirim respon 422 Unprocessable Entity
func FailValidation(c *fiber.Ctx, fields map[string]string) error {
	return c.Status(fiber.StatusUnprocessableEntity).JSON(model.WebResponse{
		Success: false,
		Message: "validasi gagal",
		Errors:  fields,
	})
}

// NoContent mengirim status 204
func NoContent(c *fiber.Ctx) error {
	return c.SendStatus(fiber.StatusNoContent)
}
