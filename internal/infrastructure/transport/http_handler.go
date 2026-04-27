package transport

import (
	"context"
	"io"
	"log"
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

func (h *MetricHandler) ReceiveMetric(c *gin.Context) {
	// getting request ID from context for logging and response
	requestID, _ := c.Get("RequestID")
	contentType := c.GetHeader("Content-Type")

	var metric models.HealthMetric

	// checking content type
	if contentType == "application/x-protobuf" {
		// c.JSON(http.StatusUnsupportedMediaType, gin.H{"error": "Use application/x-protobuf", "rid": requestID})
		// return

		// parse protobuf body
		rawData, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read body", "rid": requestID})
			return
		}

		var protoReq metrics_proto.HealthMetricRequest
		if err := proto.Unmarshal(rawData, &protoReq); err != nil {
			log.Printf("[RID: %s] Proto unmarshal error: %v", requestID, err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid protobuf data", "rid": requestID})
			return
		}

		// map to domain model
		metric = models.HealthMetric{
			ExternalID: protoReq.GetExternalId(),
			PetID:      protoReq.GetPetId(),
			Type:       protoReq.GetType(),
			Value:      protoReq.GetValue(),
			Timestamp:  time.Now(),
		}
	} else {
		var jsonReq struct {
			ExternalID string  `json:"external_id"`
			PetID      uint64  `json:"pet_id" binding:"required"`
			Type       string  `json:"type" binding:"required"`
			Value      float64 `json:"value" binding:"required"`
		}
		if err := c.ShouldBindJSON(&jsonReq); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Use JSON or Protobuf correctly", "details": err.Error(), "rid": requestID})
			return
		}

		metric = models.HealthMetric{
			ExternalID: jsonReq.ExternalID,
			PetID:      jsonReq.PetID,
			Type:       jsonReq.Type,
			Value:      jsonReq.Value,
			Timestamp:  time.Now(),
		}
	}

	// validate the metric
	if err := h.validator.Validate(metric); err != nil {
		log.Printf("[RID: %s] Validation failed: %v", requestID, err)
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"error":   "Validation failed",
			"details": err.Error(),
			"rid":     requestID,
		})
		return
	}

	// pass request ID in context for use case processing
	ctx := context.WithValue(c.Request.Context(), "RequestID", requestID)

	if err := h.useCase.ProcessMetric(ctx, &metric); err != nil {
		log.Printf("[RID: %s] UseCase error: %v", requestID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "processing failed", "rid": requestID})
		return
	}

	// success response
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Metric accepted",
		"rid":     requestID,
	})
}
