package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/lambdavault/api/internal/application/usecase"
	"github.com/lambdavault/api/internal/interfaces/http/response"
)

type GeneratorHandler struct {
	useCase usecase.GeneratorUseCase
}

func NewGeneratorHandler(useCase usecase.GeneratorUseCase) *GeneratorHandler {
	return &GeneratorHandler{useCase: useCase}
}

func (h *GeneratorHandler) Generate(c *gin.Context) {
	length := parseIntQuery(c, "length", 16)
	uppercase := parseBoolQuery(c, "uppercase", true)
	lowercase := parseBoolQuery(c, "lowercase", true)
	numbers := parseBoolQuery(c, "numbers", true)
	symbols := parseBoolQuery(c, "symbols", true)

	password, err := h.useCase.Generate(length, uppercase, lowercase, numbers, symbols)
	if err != nil {
		response.InternalServerError(c, "failed to generate password")
		return
	}

	response.OK(c, gin.H{
		"password": password,
		"length":   len(password),
		"options": gin.H{
			"uppercase": uppercase,
			"lowercase": lowercase,
			"numbers":   numbers,
			"symbols":   symbols,
		},
	})
}

func parseIntQuery(c *gin.Context, key string, defaultVal int) int {
	val := c.Query(key)
	if val == "" {
		return defaultVal
	}
	parsed, err := strconv.Atoi(val)
	if err != nil {
		return defaultVal
	}
	return parsed
}

func parseBoolQuery(c *gin.Context, key string, defaultVal bool) bool {
	val := c.Query(key)
	if val == "" {
		return defaultVal
	}
	return val == "true" || val == "1"
}
