package chainlink

import (
	"fmt"
	"net/http"
	"time"

	"github.com/avast/retry-go"
	"github.com/pkg/errors"
	log "github.com/xlab/suplog"
)

type WebhookClient interface {
	TriggerJob(jobID string) error
}

type webhookClient struct {
	nodeURL  string
	icKey    string
	icSecret string

	c      *http.Client
	logger log.Logger
}

// NewChainlink creates a Chainlink Node client for triggering jobs using webhooks.
func NewWebhookClient(url, icKey, icSecret string) WebhookClient {
	return &webhookClient{
		nodeURL:  url,
		icKey:    icKey,
		icSecret: icSecret,

		c: &http.Client{
			Timeout: 15 * time.Second,
			Transport: &http.Transport{
				ResponseHeaderTimeout: 10 * time.Second,
			},
		},
		logger: log.DefaultLogger.WithField("svc", "cl_client"),
	}
}

func (c *webhookClient) TriggerJob(jobID string) error {
	jobLogger := c.logger.WithField("job", jobID)

	url := fmt.Sprintf("%s/v2/jobs/%s/runs", c.nodeURL, jobID)

	req, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		err = errors.Wrap(err, "failed to create HTTP request")
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Add("X-Chainlink-EA-AccessKey", c.icKey)
	req.Header.Add("X-Chainlink-EA-Secret", c.icSecret)

	if err := retry.Do(func() error {
		if _, err := c.c.Do(req); err != nil {
			jobLogger.WithError(err).Warningln("HTTP request temporary error")
			return err
		}

		return nil
	}); err != nil {
		jobLogger.WithError(err).Errorln("failed to trigger a job")
		return err
	}

	return nil
}
