package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"

	"heartbeat/heartbeat_pb"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"google.golang.org/grpc"
)

func handleError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

type server struct {
	heartbeat_pb.UnimplementedHeartBeatServiceServer
}

type heart_item struct {
	Id       primitive.ObjectID `bson:"_id,omitempty"`
	Bpm      int32              `bson:"bpm"`
	Username string             `bson:"username"`
}

func pushUserToDb(ctx context.Context, item heart_item) primitive.ObjectID {
	res, err := collection.InsertOne(ctx, item)
	handleError(err)

	return res.InsertedID.(primitive.ObjectID)
}

func (*server) NormalAbnormalHeartBeat(stream heartbeat_pb.HeartBeatService_NormalAbnormalHeartBeatServer) error {
	fmt.Println("NormalAbnormalHeartBeat() called")
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		bpm := req.GetBpm()
		var result string
		if bpm < 60 || bpm > 100 {
			result = fmt.Sprintf("User HeartBeat of %v is Abnormal", bpm)
		} else {
			result = fmt.Sprintf("User HeartBeat of %v is Normal", bpm)
		}
		NAResponse := heartbeat_pb.NormalAbnormalHeartBeatResponse{
			Result: result,
		}
		fmt.Printf("Sending back response %v\n", result)
		stream.Send(&NAResponse)
	}
}

func (*server) HeartBeatHistory(req *heartbeat_pb.HeartBeatHistoryRequest, stream heartbeat_pb.HeartBeatService_HeartBeatHistoryServer) error {
	fmt.Println("HeartBeatHistory() called")
	username := req.GetUsername()

	filter := bson.M{
		"username": username,
	}

	var result_data []heart_item
	cursor, err := collection.Find(context.TODO(), filter)
	handleError(err)

	cursor.All(context.Background(), &result_data)
	for _, v := range result_data {
		historyResponse := heartbeat_pb.HeartBeatHistoryResponse{
			Heartbeat: &heartbeat_pb.HeartBeat{
				Bpm:      v.Bpm,
				Username: v.Username,
			},
		}
		stream.Send(&historyResponse)
	}
	return nil
}

func (*server) LiveHeartBeat(stream heartbeat_pb.HeartBeatService_LiveHeartBeatServer) error {
	result := ""
	for {
		msg, err := stream.Recv()
		fmt.Println(msg)
		if err == io.EOF {
			return stream.SendAndClose(&heartbeat_pb.LiveHeartBeatResponse{
				Result: result,
			})
		}
		handleError(err)
		bpm := msg.GetHeartbeat().GetBpm()
		docid := pushUserToDb(context.TODO(), heart_item{
			Bpm:      bpm,
			Username: msg.GetHeartbeat().GetUsername(),
		})
		result += fmt.Sprintf("User HeartBeat = %v, docid = %v", bpm, docid)
	}
}

func (*server) UserHeartBeat(ctx context.Context, req *heartbeat_pb.HeartBeatRequest) (*heartbeat_pb.HeartBeatResponse, error) {
	fmt.Println(req)
	bpm := req.GetHeartbeat().GetBpm()
	username := req.GetHeartbeat().GetUsername()

	// newHeartItem := heart_item{
	// 	Bpm:      bpm,
	// 	Username: username,
	// }
	// docid := pushUserToDb(ctx, newHeartItem)
	result := fmt.Sprintf("User HeartBeat is %v, newly created docid is %v", bpm, username)

	heartBeatResponse := heartbeat_pb.HeartBeatResponse{
		Result: result,
	}

	return &heartBeatResponse, nil
}

var collection *mongo.Collection

func main() {
	//50051
	opts := []grpc.ServerOption{}
	s := grpc.NewServer(opts...)

	heartbeat_pb.RegisterHeartBeatServiceServer(s, &server{})

	lis, err := net.Listen("tcp", "0.0.0.0:50051")
	handleError(err)

	go func() {
		if err := s.Serve(lis); err != nil {
			handleError(err)
		}
	}()

	// client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017"))
	// handleError(err)
	// fmt.Println("MongoDB connected")

	// err = client.Connect(context.TODO())
	// handleError(err)

	// collection = client.Database("heartbeat").Collection("heartbeat")

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)

	<-ch

	// fmt.Println("Closing Mongo Connection")
	// if err := client.Disconnect(context.TODO()); err != nil {
	// 	handleError(err)
	// }

	s.Stop()
}
