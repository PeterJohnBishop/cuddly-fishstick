package emailapi

import (
	"log"
	"os"

	"github.com/resend/resend-go/v3"
)

func InitMailer() *resend.Client {
	apiKey := os.Getenv("RESEND")

	var client *resend.Client

	client = resend.NewClient(apiKey)

	return client
}

func SendMail(client *resend.Client, to []string, from string, subject string, htmlContent string) {

	params := &resend.SendEmailRequest{
		From:    from,        //"notifications@gin-svc.com",
		To:      to,          //[]string{"peterjbishop.denver@gmail.com"},
		Subject: subject,     //"Hello World",
		Html:    htmlContent, //"<p>Congrats on sending your <strong>first email</strong>!</p>",
	}

	sent, err := client.Emails.Send(params)
	if err != nil {
		log.Fatalf("failed to send email: %v", err)
	}

	log.Printf("email sent: %s", sent.Id)
}
