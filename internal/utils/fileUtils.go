package utils

import (
	"github.com/dbulyk/metrics-alerting-service/internal/stores"
	"github.com/rs/zerolog/log"
)

func SaveMetrics(mem *stores.MemStorage, filename string) error {
	listMetrics, _ := mem.ListMetrics()
	producer, err := stores.NewProducer(filename)
	if err != nil {
		return err
	}
	defer func(producer *stores.Producer) {
		err := producer.Close()
		if err != nil {
			log.Error().Msgf("ошибка закрытия файла %s", filename)
		}
	}(producer)

	err = producer.Write(listMetrics)
	if err != nil {
		return err
	}
	log.Info().Msgf("метрики сохранены в файл %s", filename)
	return nil
}

func RestoreMetrics(mem *stores.MemStorage, filename string) error {
	consumer, err := stores.NewConsumer(filename)
	if err != nil {
		return err
	}
	defer func(consumer *stores.Consumer) {
		err := consumer.Close()
		if err != nil {
			log.Error().Msgf("ошибка закрытия файла %s", filename)
		}
	}(consumer)

	metrics, err := consumer.Read()
	if err != nil {
		return err
	}

	for _, metric := range metrics {
		_, err := mem.SetMetric(metric.ID, metric.MType, metric.Value, metric.Delta)
		if err != nil {
			return err
		}
	}
	log.Info().Msgf("метрики восстановлены из файла %s", filename)
	return nil
}
