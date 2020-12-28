package grpc

import (
	"context"
	"github.com/AleksK1NG/email-microservice/config"
	"github.com/AleksK1NG/email-microservice/internal/email"
	emailService "github.com/AleksK1NG/email-microservice/internal/email/proto"
	"github.com/AleksK1NG/email-microservice/internal/models"
	"github.com/AleksK1NG/email-microservice/pkg/grpc_errors"
	"github.com/AleksK1NG/email-microservice/pkg/logger"
	"github.com/opentracing/opentracing-go"
	"google.golang.org/grpc/status"
)

// Email gRPC microservice
type EmailMicroservice struct {
	cfg     *config.Config
	logger  logger.Logger
	emailUC email.EmailsUseCase
}

// Email gRPC microservice constructor
func NewEmailMicroservice(emailUC email.EmailsUseCase, logger logger.Logger, cfg *config.Config) *EmailMicroservice {
	return &EmailMicroservice{emailUC: emailUC, logger: logger, cfg: cfg}
}

// Send emails
func (e *EmailMicroservice) SendEmails(ctx context.Context, r *emailService.SendEmailRequest) (*emailService.SendEmailResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "EmailUseCase.SendEmail")
	defer span.Finish()

	mail := &models.Email{
		From:    e.cfg.Smtp.User,
		To:      r.GetTo(),
		Body:    r.GetBody(),
		Subject: r.GetSubject(),
	}

	if err := mail.PrepareAndValidate(ctx); err != nil {
		e.logger.Errorf("PrepareAndValidate: %v", err)
		return nil, status.Errorf(grpc_errors.ParseGRPCErrStatusCode(err), "PrepareAndValidate: %v", err)
	}

	if err := e.emailUC.PublishEmailToQueue(ctx, mail); err != nil {
		e.logger.Errorf("emailUC.PublishEmailToQueue: %v", err)
		return nil, status.Errorf(grpc_errors.ParseGRPCErrStatusCode(err), "emailUC.PublishEmailToQueue: %v", err)
	}

	return &emailService.SendEmailResponse{Status: "Ok"}, nil
}

// Find email by id
func (e *EmailMicroservice) FindEmailById(ctx context.Context, r *emailService.FindEmailByIdRequest) (*emailService.FindEmailByIdResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "EmailUseCase.FindEmailById")
	defer span.Finish()

	return &emailService.FindEmailByIdResponse{Email: nil}, nil
}
