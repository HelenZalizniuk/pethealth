package transport

import (
	"io"
	"net/http"
	metrics_proto "pethealth/internal/infrastructure/transport/proto"
	"pethealth/internal/usecase"

	"github.com/gin-gonic/gin"
	"google.golang.org/protobuf/proto"
)

type MetricHandler struct {
	useCase *usecase.MetricUseCase
}

func NewMetricHandler(uc *usecase.MetricUseCase) *MetricHandler {
	return &MetricHandler{useCase: uc}
}

func (h *MetricHandler) ReceiveMetric(c *gin.Context) {

	if c.GetHeader("Content-Type") != "application/x-protobuf" {
		c.JSON(http.StatusUnsupportedMediaType, gin.H{"error": "Use application/x-protobuf"})
		return
	}

	rawData, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read body"})
		return
	}

	var req metrics_proto.HealthMetricRequest
	if err := proto.Unmarshal(rawData, &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid protobuf data"})
		return
	}

	// TODO: Вызов UseCase (мы его обновим на следующем шаге)

	c.JSON(http.StatusOK, metrics_proto.HealthMetricResponse{
		Status:  "success",
		Message: "Metric accepted",
	})
}
