package email

import (
	"bytes"
	"context"
	"log"
	"os"
	"strings"
	"text/template"

	"github.com/Becram/devops-automation/pkg/config"
	"github.com/Becram/devops-automation/pkg/logging"

	"github.com/nikoksr/notify"
	"github.com/nikoksr/notify/service/sendgrid"
)

var logger logging.Logger

type (
	Client struct {
		cfg      *config.ScrapeConf
		notify   *notify.Notify
		sendFunc sendFunc
	}
	sendFunc func(ctx context.Context, subject, message string) error
)

func NewMailerClient(cfg *config.ScrapeConf) *Client {
	n := notify.New()
	c := &Client{
		cfg:      cfg,
		notify:   n,
		sendFunc: n.Send,
	}
	if cfg.Email.Enable == "true" {
		c.notify.UseServices(c.email())
	}

	return c
}

func (c *Client) SendEmail() {
	// Create a telegram service. Ignoring error for demo simplicity.
	template := getTemplate(c)
	var emails = strings.Split(c.cfg.Email.SendTo, ",")
	logger.Info("running /email", "sending email to", emails[:])
	// fmt.Printf("Sending email to %s\n", emails[:])
	err := c.sendFunc(
		context.Background(),
		c.cfg.Email.Subject,
		template,
	)
	if err != nil {
		logger.Error(err, "cannot send email")
	}
}

func (c *Client) email() notify.Notifier {
	if c.cfg.Email.Enable == "true" {
		m := sendgrid.New(os.Getenv("SG_API_KEY"), c.cfg.Email.Sender, c.cfg.Email.SenderName)

		m.AddReceivers(c.cfg.Email.SendTo)
		return m
	}
	return nil
}

func getTemplate(c *Client) string {
	t := template.Must(template.New("email.html").ParseFiles(c.cfg.Email.TemplatePath))
	logger.Info("running /email", "getting template", c.cfg.Email.TemplatePath)
	var buffer bytes.Buffer
	err := t.Execute(&buffer, c.cfg.Email)
	if err != nil {
		log.Fatalln(err)
	}

	return buffer.String()
}
