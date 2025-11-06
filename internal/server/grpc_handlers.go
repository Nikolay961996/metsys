package server

import (
	"context"
	"errors"
	"github.com/Nikolay961996/metsys/internal/server/repositories"
	"github.com/Nikolay961996/metsys/internal/server/router"

	"github.com/Nikolay961996/metsys/models"
	"github.com/Nikolay961996/metsys/proto"
)

type MetricsServiceServer struct {
	proto.UnimplementedMetricsServiceServer
	Storage repositories.Storage
}

func (s *MetricsServiceServer) GetMetric(ctx context.Context, req *proto.MetricRequest) (*proto.MetricResponse, error) {
	metric := &models.Metrics{
		ID:    req.Id,
		MType: req.Type,
	}

	actualMetric, err := router.GetActualMetrics(s.Storage, metric)
	if err != nil {
		return nil, err
	}

	response := &proto.MetricResponse{
		Id:    actualMetric.ID,
		Type:  actualMetric.MType,
		Value: 0,
		Delta: 0,
	}

	if actualMetric.Value != nil {
		response.Value = *actualMetric.Value
	}
	if actualMetric.Delta != nil {
		response.Delta = *actualMetric.Delta
	}

	return response, nil
}

func (s *MetricsServiceServer) UpdateMetric(ctx context.Context, req *proto.MetricUpdateRequest) (*proto.MetricResponse, error) {
	metric := &models.Metrics{
		ID:    req.Id,
		MType: req.Type,
		Value: &req.Value,
		Delta: &req.Delta,
	}

	if metric.MType == models.Gauge {
		s.Storage.SetGauge(metric.ID, *metric.Value)
	} else if metric.MType == models.Counter {
		s.Storage.AddCounter(metric.ID, *metric.Delta)
	} else {
		return nil, errors.New("undefined metric type")
	}

	return &proto.MetricResponse{
		Id:    metric.ID,
		Type:  metric.MType,
		Value: req.Value,
		Delta: req.Delta,
	}, nil
}

func (s *MetricsServiceServer) BatchUpdateMetrics(ctx context.Context, req *proto.BatchMetricUpdateRequest) (*proto.BatchMetricUpdateResponse, error) {
	err := s.Storage.StartTransaction(ctx)
	if err != nil {
		return nil, err
	}

	responses := []*proto.MetricResponse{}
	for _, metricReq := range req.Metrics {
		metric := &models.Metrics{
			ID:    metricReq.Id,
			MType: metricReq.Type,
			Value: &metricReq.Value,
			Delta: &metricReq.Delta,
		}

		if metric.MType == models.Gauge {
			s.Storage.SetGauge(metric.ID, *metric.Value)
		} else if metric.MType == models.Counter {
			s.Storage.AddCounter(metric.ID, *metric.Delta)
		} else {
			return nil, errors.New("undefined metric type")
		}

		responses = append(responses, &proto.MetricResponse{
			Id:    metric.ID,
			Type:  metric.MType,
			Value: metricReq.Value,
			Delta: metricReq.Delta,
		})
	}

	err = s.Storage.CommitTransaction()
	if err != nil {
		return nil, err
	}

	return &proto.BatchMetricUpdateResponse{Metrics: responses}, nil
}
