package main

import (
	"context"
	"fmt"
	"heartbeat/heartbeat_pb"
	"io"
	"log"
	"math/rand"
	"sync"
	"time"

	"google.golang.org/grpc"
)

func handleError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func generateBPM() int32 {
	bpm := rand.Intn(100)
	return int32(bpm)
}

func NormalAbnormalHeartBeat(c heartbeat_pb.HeartBeatServiceClient) {
	var wg sync.WaitGroup

	stream, err := c.NormalAbnormalHeartBeat(context.Background())
	handleError(err)
	for t := 0; t < 10; t++ {
		newNARequest := &heartbeat_pb.NormalAbnormalHeartBeatRequest{
			Bpm: generateBPM(),
		}
		stream.Send(newNARequest)
		fmt.Printf("sent %v\n", newNARequest)
	}
	stream.CloseSend()

	wg.Add(1)
	go func() {
		for {
			msg, err := stream.Recv()
			handleError(err)
			if err == io.EOF {
				wg.Done()
				break
			}
			fmt.Printf("Received %v\n", msg)
		}
	}()
	wg.Wait()
}

func HeartBeatHistory(c heartbeat_pb.HeartBeatServiceClient) {
	newHistoryRequest := heartbeat_pb.HeartBeatHistoryRequest{
		Username: "manosriram",
	}
	res_stream, err := c.HeartBeatHistory(context.Background(), &newHistoryRequest)
	handleError(err)

	for {
		msg, err := res_stream.Recv()
		handleError(err)
		if err == io.EOF {
			break
		}
		time.Sleep(1 * time.Second)
		fmt.Println(msg)
	}
}

func LiveHeartBeat(c heartbeat_pb.HeartBeatServiceClient) {
	stream, err := c.LiveHeartBeat(context.Background())
	handleError(err)
	for t := 0; t < 10; t++ {
		newRequest := &heartbeat_pb.LiveHeartBeatRequest{
			Heartbeat: &heartbeat_pb.HeartBeat{
				Bpm:      generateBPM(),
				Username: "manosriram",
			},
		}
		fmt.Println("Request Sent ", newRequest, "  ", t)
		stream.Send(newRequest)
	}

	stream.CloseAndRecv()

}

func UserHeartBeat(c heartbeat_pb.HeartBeatServiceClient) {
	heartbeatRequest := heartbeat_pb.HeartBeatRequest{
		Heartbeat: &heartbeat_pb.HeartBeat{
			Bpm:      75,
			Username: "manosriram",
		},
	}

	c.UserHeartBeat(context.Background(), &heartbeatRequest)
}

func main() {
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	handleError(err)

	fmt.Println("Client Started")
	defer conn.Close()

	c := heartbeat_pb.NewHeartBeatServiceClient(conn)

	UserHeartBeat(c)
	// LiveHeartBeat(c)
	// HeartBeatHistory(c)
	// NormalAbnormalHeartBeat(c)
}
