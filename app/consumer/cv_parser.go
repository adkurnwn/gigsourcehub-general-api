package consumer

import (
	"context"
	"encoding/json"

	"github.com/adkurnwn/gigsourcehub-general-api/domain"
	"github.com/sirupsen/logrus"
)

type CVParserConsumer struct {
	mq   domain.MessageBroker
	repo domain.GormRepo
}

func NewCVParserConsumer(mq domain.MessageBroker, repo domain.GormRepo) *CVParserConsumer {
	return &CVParserConsumer{
		mq:   mq,
		repo: repo,
	}
}

func (c *CVParserConsumer) Start(ctx context.Context) error {
	return c.mq.Consume(ctx, "parsed_cv", c.handleMessage)
}

func (c *CVParserConsumer) handleMessage(msg []byte) error {
	var payload map[string]interface{}
	if err := json.Unmarshal(msg, &payload); err != nil {
		logrus.Errorf("CVParserConsumer: Failed to unmarshal message: %v", err)
		return err
	}

	idVal, ok := payload["id"]
	if !ok {
		logrus.Errorf("CVParserConsumer: Message missing 'id' field")
		return nil // Return nil so it doesn't requeue endlessly
	}

	idStr, ok := idVal.(string)
	if !ok {
		logrus.Errorf("CVParserConsumer: 'id' field is not a string")
		return nil
	}

	parsedData, err := json.Marshal(payload)
	if err != nil {
		logrus.Errorf("CVParserConsumer: Failed to marshal parsed data: %v", err)
		return err
	}

	ctx := context.Background()

	// Fetch existing CV
	cv, err := c.repo.GetCVByID(ctx, idStr)
	if err != nil {
		logrus.Errorf("CVParserConsumer: CV not found for ID %s: %v", idStr, err)
		return nil // Don't requeue if CV doesn't exist in DB
	}

	// Update CV
	jsonStr := string(parsedData)
	cv.ParsedData = &jsonStr
	cv.Status = "PARSED"

	if err := c.repo.UpdateCV(ctx, cv); err != nil {
		logrus.Errorf("CVParserConsumer: Failed to update CV ID %s: %v", idStr, err)
		return err
	}

	logrus.Infof("CVParserConsumer: Successfully updated parsed data for CV ID %s", idStr)
	return nil
}
