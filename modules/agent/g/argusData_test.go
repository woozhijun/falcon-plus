package g

import (
	"testing"
	"log"
	. "github.com/smartystreets/goconvey/convey"
	"time"
	"github.com/open-falcon/falcon-plus/common/model"
)

func TestReportArgusMetrics(t *testing.T) {

	Convey("Report argus metrics testing", t, func() {

		var c *SingleConnGrpcClient = &SingleConnGrpcClient{
			ServerAddr: "test-argus-data.monitor.mobike.io:9000",
			Timeout:    time.Duration(1000) * time.Millisecond,
		}
		if c.InitGrpcClient() != nil {
			log.Fatalf("init failed.")
			return
		}

		var metrics []*model.MetricValue
		metricValue := &model.MetricValue{"influx04.mobike.io","local.active", 1, 60, "Counter", "", 1542620866}
		metrics = append(metrics, metricValue)
		So(reportMetrics(c, metrics), ShouldEqual,true)
	})

}
