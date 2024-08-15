package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/segmentio/kafka-go"
	lib "github.com/with-autro/autro-library"
	notification "github.com/with-autro/autro-notification/notification"
)

var (
	kafkaBroker string
	kafkaTopic  string
)

func init() {
	kafkaBroker = os.Getenv("KAFKA_BROKER")
	if kafkaBroker == "" {
		kafkaBroker = "kafka:9092"
	}

	kafkaTopic = os.Getenv("KAFKA_TOPIC")
	if kafkaTopic == "" {
		kafkaTopic = "signal-to-notification"
	}

}

func createReader() *kafka.Reader {
	return kafka.NewReader(kafka.ReaderConfig{
		Brokers:     []string{kafkaBroker},
		Topic:       kafkaTopic,
		MaxAttempts: 5,
	})
}

func processSignal(signalResult lib.SignalResult) error {
	log.Printf("Received signal: %+v", signalResult)

	discordColor := notification.GetColorForDiscord(signalResult.Signal)

	title := fmt.Sprintf("New Signal: %s", signalResult.Signal)
	description := generateDescription(signalResult)

	discordEmbed := notification.Embed{
		Title:       title,
		Description: description,
		Color:       discordColor,
	}
	if err := notification.SendDiscordAlert(discordEmbed); err != nil {
		log.Printf("Error sending Discord alert: %v", err)
		return err
	}

	log.Println("Notifications sent successfully")
	return nil
}

func generateDescription(signalResult lib.SignalResult) string {
	// Convert timestamp to Korean time
	koreaLocation, err := time.LoadLocation("Asia/Seoul")
	if err != nil {
		log.Printf("Error loading Asia/Seoul timezone: %v", err)
		koreaLocation = time.UTC
	}
	timestamp := time.Unix(signalResult.Timestamp/1000, 0).In(koreaLocation).Format("2006-01-02 15:04:05 MST")

	description := fmt.Sprintf("Signal: %s for BTCUSDT at %s\n\n", signalResult.Signal, timestamp)
	description += fmt.Sprintf("Price : %.3f\n", signalResult.Price)
	description += fmt.Sprintf("Stoploss : %.3f, Takeprofit: %.3f\n\n", signalResult.StopLoss, signalResult.TakeProfie)

	description += "=======[LONG]=======\n"
	description += fmt.Sprintf("[EMA200] : %v \n", signalResult.Conditions.Long.EMA200Condition)
	description += fmt.Sprintf("EMA200: %.3f, Diff: %.3f\n\n", signalResult.Conditions.Long.EMA200Value, signalResult.Conditions.Long.EMA200Diff)

	description += fmt.Sprintf("[MACD] : %v \n", signalResult.Conditions.Long.MACDCondition)
	description += fmt.Sprintf("MACD Line: %.3f, Signal Line: %.3f, Histogram: %.3f\n\n", signalResult.Conditions.Long.MACDNowMACDLine, signalResult.Conditions.Long.MACDNowSignalLine, signalResult.Conditions.Long.MACDHistogram)

	description += fmt.Sprintf("[Parabolic SAR] : %v \n", signalResult.Conditions.Long.ParabolicSARCondition)
	description += fmt.Sprintf("ParabolicSAR: %.3f, Diff: %.3f\n\n", signalResult.Conditions.Long.ParabolicSARValue, signalResult.Conditions.Long.ParabolicSARDiff)
	description += "=====================\n\n"

	description += "=======[SHORT]=======\n"
	description += fmt.Sprintf("[EMA200] : %v \n", signalResult.Conditions.Short.EMA200Condition)
	description += fmt.Sprintf("EMA200: %.3f, Diff: %.3f\n\n", signalResult.Conditions.Short.EMA200Value, signalResult.Conditions.Short.EMA200Diff)

	description += fmt.Sprintf("[MACD] : %v \n", signalResult.Conditions.Short.MACDCondition)
	description += fmt.Sprintf("MACD Line: %.3f, Signal Line: %.3f, Histogram: %.3f\n\n", signalResult.Conditions.Short.MACDNowMACDLine, signalResult.Conditions.Short.MACDNowSignalLine, signalResult.Conditions.Short.MACDHistogram)

	description += fmt.Sprintf("[Parabolic SAR] : %v \n", signalResult.Conditions.Short.ParabolicSARCondition)
	description += fmt.Sprintf("ParabolicSAR: %.3f, Diff: %.3f\n\n", signalResult.Conditions.Short.ParabolicSARValue, signalResult.Conditions.Short.ParabolicSARDiff)
	description += "=====================\n"

	return description
}

func main() {
	if err := notification.InitNotifications(); err != nil {
		log.Fatalf("Failed to initialize notifications: %v", err)
	}

	reader := createReader()
	defer reader.Close()

	log.Printf("Notification service Kafka consumer started. Listening on topic: %s", kafkaTopic)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle graceful shutdown
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	go func() {
		for {
			select {
			case <-signals:
				log.Println("Interrupt received, shutting down...")
				cancel()
				return
			case <-ctx.Done():
				return
			default:
				msg, err := reader.ReadMessage(ctx)
				if err != nil {
					log.Printf("Error reading message: %v", err)
					continue
				}

				var signalResult lib.SignalResult
				if err := json.Unmarshal(msg.Value, &signalResult); err != nil {
					log.Printf("Error unmarshalling message: %v", err)
					continue
				}

				if err := processSignal(signalResult); err != nil {
					log.Printf("Error processing signal: %v", err)
				}
			}
		}
	}()

	<-ctx.Done()
	log.Println("Notification service shutting down")

}

// package main

// import (
// 	"context"
// 	"log"
// 	"net"
// 	"os"

// 	notification "github.com/Lux-N-Sal/autro-notification/notification"
// 	pb "github.com/Lux-N-Sal/autro-notification/proto"
// 	"google.golang.org/grpc"
// 	"google.golang.org/grpc/reflection"
// )

// type server struct {
// 	pb.UnimplementedNotificationServiceServer
// }

// func (s *server) SendNotification(ctx context.Context, req *pb.NotificationRequest) (*pb.NotificationResponse, error) {
// 	log.Printf("Received notification request: %v", req)

// 	discordColor, slackColor := notification.GetColorForSignal(req.Signal)

// 	// Discord 알림 전송
// 	discordEmbed := notification.Embed{
// 		Title:       req.Title,
// 		Description: req.Description,
// 		Color:       discordColor,
// 	}
// 	if err := notification.SendDiscordAlert(discordEmbed); err != nil {
// 		log.Printf("Error sending Discord alert: %v", err)
// 		return &pb.NotificationResponse{Success: false, Message: "Failed to send Discord alert"}, nil
// 	}

// 	// Slack 알림 전송
// 	slackAttachment := notification.Attachment{
// 		Color: slackColor,
// 		Text:  req.Description,
// 	}
// 	if err := notification.SendSlackAlert(slackAttachment); err != nil {
// 		log.Printf("Error sending Slack alert: %v", err)
// 		return &pb.NotificationResponse{Success: false, Message: "Failed to send Slack alert"}, nil
// 	}

// 	return &pb.NotificationResponse{Success: true, Message: "Notifications sent successfully"}, nil
// }

// func main() {
// 	if err := notification.InitNotifications(); err != nil {
// 		log.Fatalf("Failed to initialize notifications: %v", err)
// 	}

// 	port := os.Getenv("PORT")
// 	if port == "" {
// 		port = "50052"
// 	}

// 	lis, err := net.Listen("tcp", ":"+port)
// 	if err != nil {
// 		log.Fatalf("failed to listen: %v", err)
// 	}

// 	s := grpc.NewServer()
// 	pb.RegisterNotificationServiceServer(s, &server{})
// 	reflection.Register(s)

// 	log.Printf("Notification service gRPC server listening on :%s", port)
// 	if err := s.Serve(lis); err != nil {
// 		log.Fatalf("failed to serve: %v", err)
// 	}
// }
