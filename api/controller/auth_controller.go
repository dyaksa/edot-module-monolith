package controller

import (
	"net/http"

	"github.com/dyaksa/warehouse/domain"
	"github.com/dyaksa/warehouse/pkg/errx"
	"github.com/dyaksa/warehouse/pkg/response/response_success"
	"github.com/gin-gonic/gin"
)

type AuthController struct {
	AuthUsecase domain.AuthUsecase
}

// Register creates a new user account
// @Summary Register a new user
// @Description Create a new user account with email, phone, and password
// @Tags Authentication
// @Accept json
// @Produce json
// @Param user body domain.AuthRegisterRequest true "User registration data"
// @Success 201 {object} map[string]interface{} "User registered successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request payload or validation failed"
// @Failure 500 {object} map[string]interface{} "Failed to register user"
// @Router /auth/register [post]
func (ac *AuthController) Register(c *gin.Context) {
	var payload domain.AuthRegisterRequest

	if err := c.ShouldBindJSON(&payload); err != nil {
		c.Error(errx.E(errx.CodeValidation, "invalid request payload", errx.Op("AuthController.Register"), err))
		return
	}

	_, err := ac.AuthUsecase.Register(c.Request.Context(), payload)
	if err != nil {
		c.Error(errx.E(errx.CodeInternal, "failed to register user", errx.Op("AuthController.Register"), err))
		return
	}

	response_success.JSON(c).Msg("User registered successfully").Status("success").Send(http.StatusCreated)
}

// Login authenticates a user and returns a JWT token
// @Summary User login
// @Description Authenticate user with email/phone and password, returns JWT access token
// @Tags Authentication
// @Accept json
// @Produce json
// @Param credentials body domain.AuthLoginRequest true "User login credentials"
// @Success 200 {object} map[string]interface{} "Login successful with access token"
// @Failure 400 {object} map[string]interface{} "Invalid request payload"
// @Failure 401 {object} map[string]interface{} "Invalid credentials"
// @Failure 500 {object} map[string]interface{} "Failed to login"
// @Router /auth/login [post]
func (ac *AuthController) Login(c *gin.Context) {
	var payload domain.AuthLoginRequest

	if err := c.ShouldBindJSON(&payload); err != nil {
		c.Error(errx.E(errx.CodeValidation, "invalid request payload", errx.Op("AuthController.Login"), err))
		return
	}

	token, err := ac.AuthUsecase.Login(c.Request.Context(), payload)
	if err != nil {
		// differentiate invalid identifier vs generic internal
		c.Error(errx.E(errx.CodeInternal, "failed to login", errx.Op("AuthController.Login"), err))
		return
	}

	response_success.JSON(c).Msg("Login successful").Status("success").Data(gin.H{"access_token": token}).Send(http.StatusOK)
}
