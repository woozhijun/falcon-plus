package g

import (
	"google.golang.org/grpc"
	"sync"
	"log"
	ap "github.com/open-falcon/falcon-plus/modules/agent/argus_proto"
	"time"
	"golang.org/x/net/context"
	"encoding/json"
	"fmt"
)

type SingleConnGrpcClient struct {
	sync.RWMutex
	grpcClient 			ap.MetricReportingServiceClient
	ServerAddr  		string
	Timeout   			time.Duration
}

func (this *SingleConnGrpcClient) String() string {
	return fmt.Sprintf(
		"<grpcClient=%v, ServerAddr:%s, Timeout=%v>",
		this.grpcClient,
		this.ServerAddr,
		this.Timeout,
	)
}

func (this *SingleConnGrpcClient) InitConnGrpcClient() error {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	conn, err := grpc.Dial(this.ServerAddr, opts...)
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
		return err
	}
	//defer conn.Close()
	this.grpcClient = ap.NewMetricReportingServiceClient(conn)
	return nil
}

func (this *SingleConnGrpcClient) ReportData(data []byte) (interface{}, error) {
	this.Lock()
	defer this.Unlock()
	ctx, cancel := context.WithTimeout(context.Background(), this.Timeout)
	defer cancel()
	stream, err := this.grpcClient.Report(ctx)
	if err != nil {
		log.Fatalf("%v.Report(_) = _, %v", this.grpcClient, err)
		return nil,err
	}
	var metric *ap.Metric
	if err := json.Unmarshal(data, &metric); err != nil {
		log.Fatalf("Json unmarshal(%v) = %v", data, err)
		return nil,err
	}

	if err := stream.Send(metric); err != nil {
		log.Fatalf("%v.Send(%v) = %v", stream, data, err)
		return nil,err
	}

	reply, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatalf("%v.CloseAndRecv() got error %v, want %v", stream, err, nil)
		return nil,err
	}
	defer stream.CloseAndRecv()

	log.Printf("Route summary: %v", reply)
	return reply, nil
}