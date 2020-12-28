package grpc

import (
	"context"
	"github.com/AleksK1NG/email-microservice/internal/email"
	emailService "github.com/AleksK1NG/email-microservice/internal/email/proto"
	"github.com/opentracing/opentracing-go"
)

// Email gRPC microservice
type EmailMicroservice struct {
	emailUC email.EmailsUseCase
}

// Email gRPC microservice constructor
func NewEmailMicroservice(emailUC email.EmailsUseCase) *EmailMicroservice {
	return &EmailMicroservice{emailUC: emailUC}
}

// Send emails
func (e *EmailMicroservice) SendEmails(ctx context.Context, request *emailService.SendEmailRequest) (*emailService.SendEmailResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "EmailUseCase.SendEmail")
	defer span.Finish()

	return nil, nil
}
