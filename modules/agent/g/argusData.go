package g

import (
	"log"
	"github.com/open-falcon/falcon-plus/common/model"

	"math/rand"
	"time"
	"encoding/json"
)

func reportArgusMetrics(metrics []*model.MetricValue, resp *model.TransferResponse)  {

	result, err := json.Marshal(metrics)
	if err != nil {
		log.Println("json paser failed. ", err)
	}
	log.Println("result: " + string(result))

	result1, err := json.Marshal(parseArgusMetric(metrics))
	if err != nil {
		log.Println("argus json paser failed. ", err)
	}
	log.Println("argus Metric Result:" + string(result1))
	rand.Seed(time.Now().UnixNano())
}

func parseArgusMetric(metrics []*model.MetricValue) (*model.ArgusMetric) {
	if metrics == nil {
		return nil
	}

	argusMetrics := make(map[string]float64)
	tags := make(map[string]string)
	for _, metric := range metrics {
		value,ok := metric.Value.(float64)
		if !ok {
			log.Println("metric value parse error. Value:" + metric.Value.(string))
			continue
		}
		argusMetrics[metric.Metric] = value
		tags["endpoint"] = metric.Endpoint
	}

	return &model.ArgusMetric{
		Group:"open-falcon",
		Service:"falcon-agent",
		Metrics:argusMetrics,
		Tags:tags,
		MetricType:"falcon",
		MeterType:"",
		Step:60,
		Timestamp:time.Now().UnixNano(),
	}
}