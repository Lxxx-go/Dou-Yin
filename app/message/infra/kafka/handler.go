package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/IBM/sarama"
	"github.com/Lxxx-go/Dou-Yin/app/message/infra/db"
	"github.com/Lxxx-go/Dou-Yin/app/message/infra/mongodb"
	"time"
)

func MessageHandler(kfMessage *sarama.ConsumerMessage) error {

	var msg db.Message
	if err := json.Unmarshal(kfMessage.Value, &msg); err != nil {
		fmt.Printf("Unmarshal message fail " + err.Error())
		return err
	}

	parent := context.Background()
	ctx, cancel := context.WithTimeout(parent, 5*time.Second)
	defer cancel()
	if err := db.CreateMessage(ctx, &msg); err != nil {
		fmt.Printf("Insert message in db fail " + err.Error())
		return err
	}

	if err := mongodb.InsertMessage(ctx, &mongodb.MgMessage{
		ThreadId:    msg.ThreadId,
		FromUserId:  msg.FromUserId,
		ToUserId:    msg.ToUserId,
		Contents:    msg.Contents,
		MessageUUID: msg.MessageUUID,
		CreateTime:  msg.CreateTime,
	}); err != nil {
		fmt.Printf("Insert message in mongo db fail " + err.Error())
		return err
	}

	return nil
}
