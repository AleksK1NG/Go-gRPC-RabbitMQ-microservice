package grpc

import (
	"context"
	"encoding/json"
	"github.com/AleksK1NG/email-microservice/internal/email"
	"github.com/AleksK1NG/email-microservice/internal/email/delivery/rabbitmq"
	emailService "github.com/AleksK1NG/email-microservice/internal/email/proto"
	"github.com/AleksK1NG/email-microservice/internal/models"
	"github.com/AleksK1NG/email-microservice/pkg/grpc_errors"
	"github.com/AleksK1NG/email-microservice/pkg/logger"
	"github.com/AleksK1NG/email-microservice/pkg/mime_types"
	"github.com/opentracing/opentracing-go"
	"google.golang.org/grpc/status"
)

// Email gRPC microservice
type EmailMicroservice struct {
	logger    logger.Logger
	emailUC   email.EmailsUseCase
	publisher *rabbitmq.EmailsPublisher
}

// Email gRPC microservice constructor
func NewEmailMicroservice(emailUC email.EmailsUseCase, publisher *rabbitmq.EmailsPublisher, logger logger.Logger) *EmailMicroservice {
	return &EmailMicroservice{emailUC: emailUC, publisher: publisher, logger: logger}
}

// Send emails
func (e *EmailMicroservice) SendEmails(ctx context.Context, r *emailService.SendEmailRequest) (*emailService.SendEmailResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "EmailUseCase.SendEmail")
	defer span.Finish()

	mail := &models.Email{
		ContentType: mime_types.MIMEApplicationJSON,
		To:          r.GetTo(),
		Body:        r.GetBody(),
		Subject:     r.GetSubject(),
	}

	mailBytes, err := json.Marshal(mail)
	if err != nil {
		e.logger.Errorf("registerReqToUserModel: %v", err)
		return nil, status.Errorf(grpc_errors.ParseGRPCErrStatusCode(err), "json.Marshal: %v", err)
	}

	if err := e.publisher.Publish(mailBytes, mail.ContentType); err != nil {
		e.logger.Errorf("publisher.Publish: %v", err)
		return nil, status.Errorf(grpc_errors.ParseGRPCErrStatusCode(err), "publisher.Publish: %v", err)
	}

	return &emailService.SendEmailResponse{Status: "Ok"}, nil
}
