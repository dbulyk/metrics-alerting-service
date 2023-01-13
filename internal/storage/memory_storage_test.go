package storage

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestMemStorage_Collect(t *testing.T) {
	type fields struct {
		Metrics map[string]interface{}
	}
	type args struct {
		ch    chan map[string]interface{}
		count int64
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name:   "Сбор метрик",
			fields: fields{Metrics: map[string]interface{}{}},
			args: args{
				ch:    make(chan map[string]interface{}),
				count: 1,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &MemStorage{
				Metrics: tt.fields.Metrics,
			}
			go m.Collect(tt.args.ch, tt.args.count)
			m.Metrics = <-tt.args.ch
			assert.Containsf(t, m.Metrics, "Mallocs", "Метрики не собрались")
		})
	}
}

func TestMemStorage_SetMetric(t *testing.T) {
	type fields struct {
		Metrics map[string]interface{}
	}
	type args struct {
		mType  string
		mName  string
		mValue float64
	}
	type checkCounter struct {
		isCounter bool
		wantValue counter
	}
	tests := []struct {
		name         string
		fields       fields
		args         args
		wantStatus   int
		checkCounter checkCounter
	}{
		{
			name:         "Неизвестный тип",
			fields:       fields{Metrics: map[string]interface{}{}},
			args:         args{mType: "test", mName: "testUnknownType", mValue: float64(100)},
			wantStatus:   http.StatusNotImplemented,
			checkCounter: checkCounter{isCounter: false},
		},
		{
			name:         "Верные аргументы с типом gauge",
			fields:       fields{Metrics: map[string]interface{}{}},
			args:         args{mType: "gauge", mName: "testGauge", mValue: float64(100)},
			wantStatus:   http.StatusOK,
			checkCounter: checkCounter{isCounter: false},
		},
		{
			name:         "Верные аргументы с типом count и несуществующим полем",
			fields:       fields{Metrics: map[string]interface{}{}},
			args:         args{mType: "counter", mName: "testCount", mValue: float64(100)},
			wantStatus:   http.StatusOK,
			checkCounter: checkCounter{isCounter: true, wantValue: 100},
		},
		{
			name: "Верные аргументы с типом count и существующим полем",
			fields: fields{Metrics: map[string]interface{}{
				"testCount": counter(100),
			}},
			args:         args{mType: "counter", mName: "testCount", mValue: float64(200)},
			wantStatus:   http.StatusOK,
			checkCounter: checkCounter{isCounter: true, wantValue: 300},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &MemStorage{
				Metrics: tt.fields.Metrics,
			}

			gotStatus := m.Set(tt.args.mType, tt.args.mName, tt.args.mValue)
			assert.Equalf(t, tt.wantStatus, gotStatus,
				"Полученный статус код (%d) не соответствует ожидаемому (%d)", gotStatus, tt.wantStatus)

			if tt.checkCounter.isCounter == true {
				assert.Equalf(t, tt.checkCounter.wantValue, m.Metrics["testCount"],
					"Значение counter не обновилось. Ожидаемое значение: %d, полученное значение: %d",
					tt.checkCounter.wantValue, m.Metrics["testCount"])
			}
		})
	}
}
