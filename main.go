package main

import (
	"context"
	"log"
	"net"
	"os"

	notification "github.com/Lux-N-Sal/autro-notification/notification"
	pb "github.com/Lux-N-Sal/autro-notification/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type server struct {
	pb.UnimplementedNotificationServiceServer
}

func (s *server) SendNotification(ctx context.Context, req *pb.NotificationRequest) (*pb.NotificationResponse, error) {
	log.Printf("Received notification request: %v", req)

	discordColor, slackColor := notification.GetColorForSignal(req.Signal)

	// Discord 알림 전송
	discordEmbed := notification.Embed{
		Title:       req.Title,
		Description: req.Description,
		Color:       discordColor,
	}
	if err := notification.SendDiscordAlert(discordEmbed); err != nil {
		log.Printf("Error sending Discord alert: %v", err)
		return &pb.NotificationResponse{Success: false, Message: "Failed to send Discord alert"}, nil
	}

	// Slack 알림 전송
	slackAttachment := notification.Attachment{
		Color: slackColor,
		Text:  req.Description,
	}
	if err := notification.SendSlackAlert(slackAttachment); err != nil {
		log.Printf("Error sending Slack alert: %v", err)
		return &pb.NotificationResponse{Success: false, Message: "Failed to send Slack alert"}, nil
	}

	return &pb.NotificationResponse{Success: true, Message: "Notifications sent successfully"}, nil
}

func main() {
	if err := notification.InitNotifications(); err != nil {
		log.Fatalf("Failed to initialize notifications: %v", err)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "50052"
	}

	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterNotificationServiceServer(s, &server{})
	reflection.Register(s)

	log.Printf("Notification service gRPC server listening on :%s", port)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
