package grpc

import (
	"context"
	"github.com/AleksK1NG/email-microservice/config"
	"github.com/AleksK1NG/email-microservice/internal/email/mock"
	emailService "github.com/AleksK1NG/email-microservice/internal/email/proto"
	"github.com/AleksK1NG/email-microservice/internal/models"
	"github.com/AleksK1NG/email-microservice/pkg/logger"
	"github.com/AleksK1NG/email-microservice/pkg/mime_types"
	"github.com/AleksK1NG/email-microservice/pkg/utils"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestEmailMicroservice_SendEmails(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	emailsUC := mock.NewMockEmailsUseCase(ctrl)

	cfg := &config.Config{
		Logger: config.Logger{},
		Smtp: config.Smtp{
			User: "mailservice@mail.ru",
		},
	}

	appLogger := logger.NewApiLogger(cfg)
	appLogger.InitLogger()

	emailMicroservice := NewEmailMicroservice(emailsUC, appLogger, cfg)

	req := &emailService.SendEmailRequest{
		To:      []string{"alex@gmail.com"},
		Body:    "<span>some text content</span>",
		Subject: "Confirm your email",
	}

	emailsUC.EXPECT().PublishEmailToQueue(gomock.Any(), gomock.Any()).Return(nil)

	response, err := emailMicroservice.SendEmails(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, response)
}

func TestEmailMicroservice_FindEmailById(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	emailsUC := mock.NewMockEmailsUseCase(ctrl)

	cfg := &config.Config{
		Logger: config.Logger{},
		Smtp: config.Smtp{
			User: "mailservice@mail.ru",
		},
	}

	appLogger := logger.NewApiLogger(cfg)
	appLogger.InitLogger()

	emailMicroservice := NewEmailMicroservice(emailsUC, appLogger, cfg)

	emailUUID := uuid.New()
	mockEmail := &models.Email{
		EmailID:     emailUUID,
		To:          []string{"alex@gmail.com"},
		Body:        "<span>some text content</span>",
		Subject:     "Confirm your email",
		From:        cfg.Smtp.User,
		ContentType: mime_types.MIMEApplicationJSON,
		CreatedAt:   time.Now(),
	}

	emailsUC.EXPECT().FindEmailById(gomock.Any(), gomock.Any()).Return(mockEmail, nil)

	req := &emailService.FindEmailByIdRequest{
		EmailUuid: emailUUID.String(),
	}

	response, err := emailMicroservice.FindEmailById(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, response)
	require.Equal(t, response.Email.EmailId, req.EmailUuid)
}

func TestEmailMicroservice_FindEmailsByReceiver(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	emailsUC := mock.NewMockEmailsUseCase(ctrl)

	cfg := &config.Config{
		Logger: config.Logger{},
		Smtp: config.Smtp{
			User: "mailservice@mail.ru",
		},
	}

	appLogger := logger.NewApiLogger(cfg)
	appLogger.InitLogger()

	emailMicroservice := NewEmailMicroservice(emailsUC, appLogger, cfg)

	emailUUID := uuid.New()
	mockEmail := &models.Email{
		EmailID:     emailUUID,
		To:          []string{"mail@gmail.com"},
		From:        "alex@gmail.com",
		Body:        "<span>some text content</span>",
		Subject:     "Confirm your email",
		ContentType: mime_types.MIMEApplicationJSON,
	}

	pq := &utils.PaginationQuery{
		Size: 10,
		Page: 1,
	}

	mockEmailsList := &models.EmailsList{
		TotalCount: 1,
		TotalPages: 1,
		Page:       1,
		Size:       1,
		HasMore:    false,
		Emails:     []*models.Email{mockEmail},
	}

	ctx := context.Background()

	emailsUC.EXPECT().FindEmailsByReceiver(gomock.Any(), mockEmail.To[0], pq).Return(mockEmailsList, nil)

	req := &emailService.FindEmailsByReceiverRequest{
		ReceiverEmail: mockEmail.To[0],
		Page:          pq.GetPage(),
		Size:          pq.GetSize(),
	}

	result, err := emailMicroservice.FindEmailsByReceiver(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, result.TotalCount, mockEmailsList.TotalCount)
}
