package api

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	log "github.com/xlab/suplog"

	"github.com/InjectiveLabs/chainlink-injective/db/model"
	"github.com/InjectiveLabs/chainlink-injective/metrics"
)

const (
	externalInitiatorAccessKeyHeader = "X-Chainlink-EA-AccessKey"
	externalInitiatorSecretHeader    = "X-Chainlink-EA-Secret"
)

type JobService interface {
	StartJob(jobID string, spec *model.JobSpec) error
	RunJob(jobID, result string) error
	StopJob(jobID string) error
}

type AuthCredentials struct {
	AccessKey string
	Secret    string
}

type HTTPServer interface {
	ListenAndServe(listenAddr string) error
	ServeHTTP(w http.ResponseWriter, r *http.Request)
	Stop(ctx context.Context) error
}

type httpServer struct {
	router  *gin.Engine
	server  *http.Server
	svc     JobService
	logger  log.Logger
	svcTags metrics.Tags
}

func NewServer(
	auth AuthCredentials,
	svc JobService,
) (HTTPServer, error) {
	if len(auth.AccessKey) == 0 {
		err := errors.New("mandatory acces key is not provided")
		return nil, err
	} else if len(auth.Secret) == 0 {
		err := errors.New("mandatory secret is not provided")
		return nil, err
	}

	srv := &httpServer{
		router: gin.Default(),
		svc:    svc,

		logger: log.WithFields(log.Fields{
			"svc": "api_srv",
		}),
		svcTags: metrics.Tags{
			"svc": "api_srv",
		},
	}

	srv.router.GET("/health", handleShowHealth())
	srv.router.POST("/runs", srv.handleJobRun())

	privateGroup := srv.router.Group("/")
	privateGroup.Use(authenticated(auth.AccessKey, auth.Secret))
	privateGroup.POST("/jobs", srv.handleJobCreate())
	privateGroup.DELETE("/jobs/:jobid", srv.handleJobStop())

	return srv, nil
}

func (s *httpServer) ListenAndServe(listenAddr string) error {
	s.server = &http.Server{
		Addr:    listenAddr,
		Handler: s.router,
	}

	return s.server.ListenAndServe()
}

func (s *httpServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

func (s *httpServer) Stop(ctx context.Context) error {
	if s.server == nil {
		return nil
	}

	if err := s.server.Shutdown(ctx); err != nil {
		return err
	}

	s.server = nil
	return nil
}

type JobCreateRequest struct {
	JobID  string        `json:"jobId"`
	Name   string        `json:"type"`
	Params model.JobSpec `json:"params"`
}

type JobHandleResponse struct {
	ID string `json:"id"`
}

func (s *httpServer) handleJobCreate() gin.HandlerFunc {
	return func(c *gin.Context) {
		metrics.ReportFuncCall(s.svcTags)
		doneFn := metrics.ReportFuncTiming(s.svcTags)
		defer doneFn()

		handlerLog := s.logger.WithField("handler", "handleJobCreate")

		var req JobCreateRequest

		if err := c.BindJSON(&req); err != nil {
			metrics.ReportFuncError(s.svcTags)
			handlerLog.WithError(err).Warningln("failed to map JSON request body")
			c.JSON(http.StatusBadRequest, nil)
			return
		}

		if err := s.svc.StartJob(req.JobID, &req.Params); err != nil {
			metrics.ReportFuncError(s.svcTags)
			handlerLog.WithError(err).Errorln("failed to start Job")
			c.JSON(http.StatusInternalServerError, nil)
			return
		}

		c.JSON(http.StatusCreated, JobHandleResponse{
			ID: req.JobID,
		})
	}
}

type JobRunRequest struct {
	JobID  string `json:"jobID"`
	Result string `json:"result"`
}

func (s *httpServer) handleJobRun() gin.HandlerFunc {
	return func(c *gin.Context) {
		metrics.ReportFuncCall(s.svcTags)
		doneFn := metrics.ReportFuncTiming(s.svcTags)
		defer doneFn()

		handlerLog := s.logger.WithField("handler", "handleJobRun")

		var req JobRunRequest

		if err := c.BindJSON(&req); err != nil {
			metrics.ReportFuncError(s.svcTags)
			handlerLog.WithError(err).Warningln("failed to map JSON request body")
			c.JSON(http.StatusBadRequest, nil)
			return
		}

		if err := s.svc.RunJob(req.JobID, req.Result); err != nil {
			metrics.ReportFuncError(s.svcTags)
			handlerLog.WithError(err).Errorln("failed to run Job")
			c.JSON(http.StatusInternalServerError, nil)
			return
		}

		c.JSON(http.StatusCreated, JobHandleResponse{
			ID: req.JobID,
		})
	}
}

func (s *httpServer) handleJobStop() gin.HandlerFunc {
	return func(c *gin.Context) {
		metrics.ReportFuncCall(s.svcTags)
		doneFn := metrics.ReportFuncTiming(s.svcTags)
		defer doneFn()

		handlerLog := s.logger.WithField("handler", "handleJobStop")

		jobID := c.Param("jobid")

		if err := s.svc.StopJob(jobID); err != nil {
			metrics.ReportFuncError(s.svcTags)
			handlerLog.WithError(err).Errorln("failed to delete Job")
			c.JSON(http.StatusInternalServerError, nil)
			return
		}

		c.JSON(http.StatusOK, JobHandleResponse{
			ID: jobID,
		})
	}
}

func authenticated(accessKey, secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		reqAccessKey := c.GetHeader(externalInitiatorAccessKeyHeader)
		reqSecret := c.GetHeader(externalInitiatorSecretHeader)
		if reqAccessKey == accessKey && reqSecret == secret {
			c.Next()
		} else {
			c.AbortWithStatus(http.StatusUnauthorized)
		}
	}
}

func handleShowHealth() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(200, gin.H{"chainlink": true})
	}
}
