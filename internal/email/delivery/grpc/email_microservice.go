package grpc

import (
	"context"

	"github.com/google/uuid"
	"github.com/opentracing/opentracing-go"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/AleksK1NG/email-microservice/config"
	"github.com/AleksK1NG/email-microservice/internal/email"
	emailService "github.com/AleksK1NG/email-microservice/internal/email/proto"
	"github.com/AleksK1NG/email-microservice/internal/models"
	"github.com/AleksK1NG/email-microservice/pkg/grpc_errors"
	"github.com/AleksK1NG/email-microservice/pkg/logger"
	"github.com/AleksK1NG/email-microservice/pkg/utils"
)

// EmailMicroservice Email gRPC microservice
type EmailMicroservice struct {
	cfg     *config.Config
	logger  logger.Logger
	emailUC email.EmailsUseCase
}

// NewEmailMicroservice Email gRPC microservice constructor
func NewEmailMicroservice(emailUC email.EmailsUseCase, logger logger.Logger, cfg *config.Config) *EmailMicroservice {
	return &EmailMicroservice{emailUC: emailUC, logger: logger, cfg: cfg}
}

// SendEmails Send emails
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

	mail.Body = utils.SanitizeString(mail.Body)

	if err := e.emailUC.PublishEmailToQueue(ctx, mail); err != nil {
		e.logger.Errorf("emailUC.PublishEmailToQueue: %v", err)
		return nil, status.Errorf(grpc_errors.ParseGRPCErrStatusCode(err), "emailUC.PublishEmailToQueue: %v", err)
	}

	return &emailService.SendEmailResponse{Status: "Ok"}, nil
}

// FindEmailById Find email by id
func (e *EmailMicroservice) FindEmailById(ctx context.Context, r *emailService.FindEmailByIdRequest) (*emailService.FindEmailByIdResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "EmailUseCase.FindEmailById")
	defer span.Finish()

	emailUUID, err := uuid.Parse(r.GetEmailUuid())
	if err != nil {
		e.logger.Errorf("uuid.Parse: %v", err)
		return nil, status.Errorf(grpc_errors.ParseGRPCErrStatusCode(err), "uuid.Parse: %v", err)
	}

	emailById, err := e.emailUC.FindEmailById(ctx, emailUUID)
	if err != nil {
		e.logger.Errorf("emailUC.FindEmailById: %v", err)
		return nil, status.Errorf(grpc_errors.ParseGRPCErrStatusCode(err), "emailUC.FindEmailById: %v", err)
	}

	return &emailService.FindEmailByIdResponse{Email: e.convertEmailToProto(emailById)}, nil
}

// FindEmailsByReceiver Find emails by receiver address
func (e *EmailMicroservice) FindEmailsByReceiver(ctx context.Context, r *emailService.FindEmailsByReceiverRequest) (*emailService.FindEmailsByReceiverResponse, error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "EmailUseCase.FindEmailById")
	defer span.Finish()

	paginationQuery := &utils.PaginationQuery{
		Size: r.GetSize(),
		Page: r.GetPage(),
	}

	emails, err := e.emailUC.FindEmailsByReceiver(ctx, r.GetReceiverEmail(), paginationQuery)
	if err != nil {
		e.logger.Errorf("emailUC.FindEmailsByReceiver: %v", err)
		return nil, status.Errorf(grpc_errors.ParseGRPCErrStatusCode(err), "emailUC.FindEmailsByReceiver: %v", err)
	}

	return &emailService.FindEmailsByReceiverResponse{
		Emails:     e.convertEmailsListToProto(emails.Emails),
		TotalPages: emails.TotalPages,
		TotalCount: emails.TotalCount,
		HasMore:    emails.HasMore,
		Page:       emails.Page,
		Size:       emails.Size,
	}, nil
}

func (e *EmailMicroservice) convertEmailsListToProto(emails []*models.Email) []*emailService.Email {
	protoEmails := make([]*emailService.Email, 0, len(emails))
	for _, m := range emails {
		protoEmails = append(protoEmails, e.convertEmailToProto(m))
	}
	return protoEmails
}

func (e *EmailMicroservice) convertEmailToProto(email *models.Email) *emailService.Email {
	return &emailService.Email{
		EmailId:     email.EmailID.String(),
		To:          email.To,
		From:        email.From,
		Body:        email.Body,
		Subject:     email.Subject,
		ContentType: email.ContentType,
		CreatedAt:   timestamppb.New(email.CreatedAt),
	}
}
