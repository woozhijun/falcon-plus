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
	rand.Seed(time.Now().UnixNano())
}