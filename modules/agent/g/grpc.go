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
	"math"
	"errors"
)

type SingleConnGrpcClient struct {
	sync.RWMutex
	grpcClient 			ap.MetricReportingServiceClient
	grpcConn			*grpc.ClientConn
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

func (this *SingleConnGrpcClient)InitGrpcClient() error {
	err := this.serverGrpcConn()
	if err != nil {
		return err
	}
	this.grpcClient = ap.NewMetricReportingServiceClient(this.grpcConn)
	return nil
}

func (this *SingleConnGrpcClient) close() {
	if this.grpcClient != nil {
		this.grpcConn.Close()
		this.grpcConn = nil
		this.grpcClient = nil
	}
}

func (this *SingleConnGrpcClient) serverGrpcConn() error {
	if this.grpcConn != nil {
		return nil
	}
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	var err error
	var retry int = 1
	for {
		this.grpcConn, err = grpc.Dial(this.ServerAddr, opts...)
		if err != nil {
			log.Fatalf("fail to dial: %v", err)
			if retry > 3 {
				return err
			}
			time.Sleep(time.Duration(math.Pow(2.0, float64(retry))) * time.Second)
			retry++
			continue
		}
		return err
	}
}

func (this *SingleConnGrpcClient) ReportData(data []byte) error {
	this.Lock()
	defer this.Unlock()

	timeout := time.Duration(15 * time.Second)
	done := make(chan error, 1)

	go func() {
		err := this.sendCall(data)
		done <- err
	}()

	select {
	case <-time.After(timeout):
		log.Printf("[WARN] grpc report timeout %v => %v", this.grpcClient, this.ServerAddr)
		this.close()
	case err := <-done:
		if err != nil {
			this.close()
			return err
		}
	}

	return nil
}

func (this *SingleConnGrpcClient) sendCall(data []byte) error {
	defer func() {
		if err := recover(); err != nil {
			log.Fatalf(">>.Panic: stopping ReportData. Error: %v, Data: %v", err, data)
		}
	}()

	if this.grpcClient == nil {
		return errors.New("sendCall client must not nil")
	}

	ctx, cancel := context.WithTimeout(context.Background(), this.Timeout)
	defer cancel()
	stream, err := this.grpcClient.Report(ctx)
	if err != nil {
		log.Fatalf("%v.Report(_) = _, %v", this.grpcClient, err)
		return err
	}
	var metric *ap.Metric
	if err := json.Unmarshal(data, &metric); err != nil {
		log.Fatalf("Json unmarshal(%v) = %v", data, err)
		return err
	}

	if err := stream.Send(metric); err != nil {
		log.Fatalf("%v.Send(%v) = %v", stream, data, err)
		return err
	}

	reply, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatalf("%v.CloseAndRecv() got error %v, want %v", stream, err, nil)
		return err
	}
	log.Printf("report summary: %v", reply)

	return nil
}