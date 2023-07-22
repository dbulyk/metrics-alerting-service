package fileio

import (
	"encoding/json"
	"os"

	"github.com/dbulyk/metrics-alerting-service/internal/models"
	"github.com/dbulyk/metrics-alerting-service/internal/storages"

	"github.com/rs/zerolog/log"
)

type Producer struct {
	file    *os.File
	encoder *json.Encoder
}

func NewProducer(filename string) (*Producer, error) {
	err := os.MkdirAll("tmp", os.ModePerm)
	if err != nil {
		return nil, err
	}

	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.ModePerm)
	if err != nil {
		return nil, err
	}
	return &Producer{
		file:    file,
		encoder: json.NewEncoder(file),
	}, nil
}

func (p *Producer) Write(metrics []*models.Metric) error {
	for i := range metrics {
		err := p.encoder.Encode(&metrics[i])
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *Producer) Close() error {
	return p.file.Close()
}

func (p *Producer) Save(mem storages.Repository, filename string) error {
	listMetrics, _ := mem.GetAll()

	defer func(p *Producer) {
		err := p.Close()
		if err != nil {
			log.Error().Msgf("ошибка закрытия файла %s", filename)
		}
	}(p)

	err := p.Write(listMetrics)
	if err != nil {
		return err
	}
	log.Info().Msgf("метрики сохранены в файл %s", filename)
	return nil
}
