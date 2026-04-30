package transport

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"pethealth/internal/domain/models"
	metrics_proto "pethealth/internal/infrastructure/transport/proto"
	"pethealth/internal/infrastructure/validator"
	"pethealth/internal/usecase"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

type MetricHandler struct {
	useCase   *usecase.MetricUseCase
	validator *validator.CustomValidator
	logger    *zap.Logger
}

func NewMetricHandler(uc *usecase.MetricUseCase, v *validator.CustomValidator, l *zap.Logger) *MetricHandler {
	return &MetricHandler{
		useCase:   uc,
		validator: v,
		logger:    l,
	}
}

func (h *MetricHandler) ProcessMetric(ctx context.Context, data []byte) error {

	metric, err := h.parseProtobuf(bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("handler: parse protobuf: %w", err)
	}

	if err := h.validator.Validate(metric); err != nil {
		return fmt.Errorf("handler: validate: %w", err)
	}

	return h.useCase.ProcessMetric(ctx, &metric)
}

func (h *MetricHandler) ReceiveMetric(c *gin.Context) {
	requestID, _ := c.Get("RequestID")
	var metric models.HealthMetric
	var err error

	// Determine content type and parse accordingly
	contentType := c.GetHeader("Content-Type")
	if contentType == "application/x-protobuf" {
		metric, err = h.parseProtobuf(c.Request.Body)
	} else {
		metric, err = h.parseJSON(c)
	}

	// Handle parsing errors
	if err != nil {
		h.logger.Warn("Failed to parse metric",
			zap.Error(err),
			zap.Any("rid", requestID))
		c.JSON(http.StatusBadRequest,
			gin.H{"error": "Invalid data format",
				"details": err.Error(),
				"rid":     requestID})
		return
	}

	// Validation
	if err := h.validator.Validate(metric); err != nil {
		c.JSON(http.StatusUnprocessableEntity,
			gin.H{"error": "Validation failed",
				"details": err.Error(),
				"rid":     requestID})
		return
	}

	// Pass request ID in context for use case processing
	ctx := context.WithValue(c.Request.Context(), "RequestID", requestID)
	if err := h.useCase.ProcessMetric(ctx, &metric); err != nil {
		h.logger.Error("UseCase execution failed",
			zap.Error(err),
			zap.Any("rid", requestID))
		c.JSON(http.StatusInternalServerError,
			gin.H{"error": "Internal processing error",
				"rid": requestID})
		return
	}

	c.JSON(http.StatusOK,
		gin.H{"status": "success",
			"rid": requestID})
}

func (h *MetricHandler) parseProtobuf(body io.Reader) (models.HealthMetric, error) {
	rawData, err := io.ReadAll(body)
	if err != nil {
		return models.HealthMetric{}, err
	}

	var protoReq metrics_proto.HealthMetricRequest
	if err := proto.Unmarshal(rawData, &protoReq); err != nil {
		return models.HealthMetric{}, err
	}

	return models.HealthMetric{
		ExternalID: protoReq.GetExternalId(),
		PetID:      protoReq.GetPetId(),
		Type:       protoReq.GetType(),
		Value:      protoReq.GetValue(),
		Timestamp:  time.Now(),
	}, nil
}

func (h *MetricHandler) parseJSON(c *gin.Context) (models.HealthMetric, error) {
	var jsonReq struct {
		ExternalID string  `json:"external_id"`
		PetID      uint64  `json:"pet_id" binding:"required"`
		Type       string  `json:"type" binding:"required"`
		Value      float64 `json:"value" binding:"required"`
	}

	if err := c.ShouldBindJSON(&jsonReq); err != nil {
		return models.HealthMetric{}, err
	}

	return models.HealthMetric{
		ExternalID: jsonReq.ExternalID,
		PetID:      jsonReq.PetID,
		Type:       jsonReq.Type,
		Value:      jsonReq.Value,
		Timestamp:  time.Now(),
	}, nil
}
