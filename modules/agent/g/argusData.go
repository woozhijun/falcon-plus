package g

import (
	"log"
	"github.com/open-falcon/falcon-plus/common/model"

	"encoding/json"
	"reflect"
	"strconv"
	"errors"
)

func reportArgusMetrics(metrics []*model.MetricValue, resp *model.TransferResponse) {

	result, err := json.Marshal(metrics)
	if err != nil {
		log.Println("json paser failed. ", err)
	}
	log.Println("result: " + string(result))

	for endpoint, data := range createArgusMetric(metrics) {
		tmp,_ := json.Marshal(data)
		log.Println("endpoint: " + endpoint)
		log.Println("argus Metric: " + string(tmp))
	}
}

func parseFloat64(value interface{}) (float64, error) {
	tmp := reflect.ValueOf(value)
	switch value.(type) {
	case string:
		return strconv.ParseFloat(tmp.String(), 64)
	case int64, int32, int16, int8, int:
		return float64(tmp.Int()), nil
	case uint64, uint32, uint16, uint8, uint:
		return float64(tmp.Uint()), nil
	case float64, float32:
		return tmp.Float(), nil
	default:
		return float64(0), errors.New("not match value for type -> " + tmp.Type().String())
	}
}

func createArgusMetric(metrics []*model.MetricValue) (map[string]*model.ArgusMetric) {
	if metrics == nil {
		return make(map[string]*model.ArgusMetric)
	}

	resultData := make(map[string]*model.ArgusMetric)
	for _, metric := range metrics {

		data, err := parseFloat64(metric.Value)
		if err != nil {
			log.Println("metricValue parseFloat error." + err.Error())
			continue
		}
		metricsResult := resultData[metric.Endpoint]
		if metricsResult == nil {
			argusMetric := &model.ArgusMetric{}
			argusMetric.Group = "open-falcon"
			argusMetric.Service = "falcon-agent"
			argusMetric.MetricType = "falcon"
			argusMetric.Step = metric.Step
			argusMetric.MeterType = metric.Metric
			argusMetric.Timestamp = metric.Timestamp * 1000

			tags := make(map[string]string)
			tags["endpoint"] = metric.Endpoint
			argusMetric.Tags = tags

			initMetrics := make(map[string]float64)
			initMetrics[metric.Metric] = data
			argusMetric.Metrics = initMetrics
			resultData[metric.Endpoint] = argusMetric
		} else {
			metricsResult.Metrics[metric.Metric] = data
		}
	}
	return resultData
}