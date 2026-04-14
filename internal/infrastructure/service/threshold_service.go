package service

import (
	"context"
)

type staticThresholdService struct {
	defaultLimit float64
}

func NewStaticThresholdService(limit float64) *staticThresholdService {
	return &staticThresholdService{defaultLimit: limit}
}

func (s *staticThresholdService) GetThreshold(ctx context.Context, petID uint64, metricType string) (float64, error) {
	// TODO: get data from DB/Redis
	return s.defaultLimit, nil
}
