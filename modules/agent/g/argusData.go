package g

import (
	"log"
	"github.com/open-falcon/falcon-plus/common/model"
	"encoding/json"
	"reflect"
	"strconv"
	"errors"
	"strings"
	"sync"
	"math/rand"
	"time"
)

var (
	ArgusClientsLock *sync.RWMutex                    = new(sync.RWMutex)
	ArgusClients     map[string]*SingleConnGrpcClient = map[string]*SingleConnGrpcClient{}
)

func reportArgusMetrics(metrics []*model.MetricValue) {
	rand.Seed(time.Now().UnixNano())
	for _, i := range rand.Perm(len(Config().ArgusData.Addrs)) {
		addr := Config().ArgusData.Addrs[i]

		c := getArgusGrpcClient(addr)
		if c == nil {
			c = initArgusGrpcClient(addr)
		}
		if reportMetrics(c, metrics) {
			break
		}
	}
}

func reportMetrics(c *SingleConnGrpcClient, metrics []*model.MetricValue) bool {

	for _, data := range createArgusMetric(metrics) {
		tmp,_ := json.Marshal(data)
		log.Println("argus Metric parser result: " + string(tmp))

		err := c.ReportData(tmp)
		if err != nil {
			log.Fatalf(">>.report argus Data server failed.<<  %v", err)
			return false
		}
	}
	return true
}

func initArgusGrpcClient(addr string) *SingleConnGrpcClient {
	var c *SingleConnGrpcClient = &SingleConnGrpcClient{
		ServerAddr: addr,
		Timeout:    time.Duration(Config().ArgusData.Timeout) * time.Millisecond,
	}
	ArgusClientsLock.Lock()
	defer ArgusClientsLock.Unlock()
	if c.InitGrpcClient() != nil {
		log.Fatalf("init failed.")
	}
	ArgusClients[addr] = c
	return c
}

func getArgusGrpcClient(addr string) *SingleConnGrpcClient {
	ArgusClientsLock.RLock()
	defer ArgusClientsLock.RUnlock()

	if c, ok := ArgusClients[addr]; ok {
		return c
	}
	return nil
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
		if tmp.Type() != nil {
			return float64(0), errors.New("not match value for type -> " + tmp.Type().String())
		} else {
			return float64(0), errors.New("not match value for type -> nil")
		}
	}
}

func parseTags(tagStr string) (map[string]string) {
	if tagStr == "" {
		return make(map[string]string)
	}
	tags := make(map[string]string)
	tvs := strings.Split(tagStr, ",")
	for _, tks := range tvs {
		tk := strings.Split(tks, "=")
		tags[tk[0]] = tk[1]
	}
	return tags
}

func createArgusMetric(metrics []*model.MetricValue) (map[string]*model.ArgusMetric) {
	if metrics == nil {
		return make(map[string]*model.ArgusMetric)
	}

	resultData := make(map[string]*model.ArgusMetric)
	for _, metric := range metrics {

		data, err := parseFloat64(metric.Value)
		if err != nil {
			log.Printf("metricValue parseFloat error: %v, Metric: %v", err.Error(), metric)
			continue
		}
		endpointKey := metric.Endpoint + "&"+ metric.Tags
		metricsResult := resultData[endpointKey]
		if metricsResult == nil {
			argusMetric := &model.ArgusMetric{}
			argusMetric.Group = "open-falcon"
			argusMetric.Service = "falcon-agent"
			argusMetric.MetricType = "falcon"
			argusMetric.Step = metric.Step
			argusMetric.MeterType = metric.Type
			argusMetric.Timestamp = metric.Timestamp * 1000

			tags := parseTags(metric.Tags)
			tags["endpoint"] = metric.Endpoint
			argusMetric.Tags = tags

			initMetrics := make(map[string]float64)
			initMetrics[metric.Metric] = data
			argusMetric.Metrics = initMetrics
			resultData[endpointKey] = argusMetric
		} else {
			metricsResult.Metrics[metric.Metric] = data
		}
	}
	return resultData
}