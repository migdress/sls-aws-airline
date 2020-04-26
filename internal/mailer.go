package internal

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ses"
)

type Mailer struct {
	client *ses.SES
}

func (m *Mailer) SendEmail(
	subject string,
	body string,
	from string,
	to []string,
	cc []string,
) error {
	_, err := m.client.SendEmail(&ses.SendEmailInput{
		Destination: &ses.Destination{
			CcAddresses: m.toPtrSlice(cc),
			ToAddresses: m.toPtrSlice(to),
		},
		Message: &ses.Message{
			Body: &ses.Body{
				Text: &ses.Content{
					Charset: aws.String("UTF-8"),
					Data:    aws.String(body),
				},
			},
			Subject: &ses.Content{
				Charset: aws.String("UTF-8"),
				Data:    aws.String(subject),
			},
		},
		Source: aws.String(from),
	})
	return err
}

func (m *Mailer) toPtrSlice(ss []string) []*string {
	ptrSlice := []*string{}
	for _, s := range ss {
		ptrSlice = append(ptrSlice, &s)
	}
	return ptrSlice
}

func NewMailer(client *ses.SES) *Mailer {
	return &Mailer{
		client: client,
	}
}
