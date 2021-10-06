package ocr2

import (
	"github.com/smartcontractkit/libocr/commontypes"
	log "github.com/xlab/suplog"
)

type monitor struct {
	logger log.Logger
}

var _ commontypes.MonitoringEndpoint = &monitor{}

func NewMonitor() commontypes.MonitoringEndpoint {
	return &monitor{
		logger: log.WithFields(log.Fields{
			"svc": "ocr2_monitor",
		}),
	}
}

func (m monitor) SendLog(log []byte) {
	m.logger.Debugln(string(log))
}
