package fileio

import (
	"bufio"
	"encoding/json"
	"os"

	"github.com/dbulyk/metrics-alerting-service/internal/models"
	"github.com/dbulyk/metrics-alerting-service/internal/storages"

	"github.com/rs/zerolog/log"
)

type Consumer struct {
	file    *os.File
	reader  *bufio.Scanner
	decoder *json.Decoder
}

func NewConsumer(filename string) (*Consumer, error) {
	file, err := os.OpenFile(filename, os.O_RDONLY|os.O_CREATE, os.ModePerm)
	if err != nil {
		return nil, err
	}
	return &Consumer{
		file:    file,
		reader:  bufio.NewScanner(file),
		decoder: json.NewDecoder(file),
	}, nil
}

func (c *Consumer) Read() ([]models.Metric, error) {
	metrics := make([]models.Metric, 0, 50)
	for c.reader.Scan() {
		metric := models.Metric{}
		if err := json.Unmarshal(c.reader.Bytes(), &metric); err != nil {
			return nil, err
		}
		metrics = append(metrics, metric)
	}
	return metrics, nil
}

func (c *Consumer) Close() error {
	return c.file.Close()
}

func (c *Consumer) Restore(mem storages.Repository) error {
	defer func(consumer *Consumer) {
		err := consumer.Close()
		if err != nil {
			log.Error().Msgf("ошибка закрытия файла")
		}
	}(c)

	metrics, err := c.Read()
	if err != nil {
		return err
	}

	for _, metric := range metrics {
		_, err = mem.Set(metric)
		if err != nil {
			log.Error().Err(err).Msgf("ошибка восстановления метрики %s", metric.ID)
		}
	}
	log.Info().Msgf("метрики восстановлены из файла")
	return nil
}