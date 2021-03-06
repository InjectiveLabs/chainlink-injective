package p2p

import (
	"context"
	"io"
	"sync"
	"time"

	"github.com/pkg/errors"
	log "github.com/xlab/suplog"

	ocrcommontypes "github.com/smartcontractkit/libocr/commontypes"
	ocrnetworking "github.com/smartcontractkit/libocr/networking"
	ocrtypes "github.com/smartcontractkit/libocr/offchainreporting2/types"

	"github.com/InjectiveLabs/chainlink-injective/keys/p2pkey"
	"github.com/InjectiveLabs/chainlink-injective/logging"
)

type NetworkingConfig struct {
	DHTLookupInterval         int
	IncomingMessageBufferSize int
	OutgoingMessageBufferSize int
	NewStreamTimeout          time.Duration
	BootstrapCheckInterval    time.Duration

	TraceLogging bool

	P2PV2AnnounceAddresses []string
	P2PV2Bootstrappers     []ocrcommontypes.BootstrapperLocator
	P2PV2DeltaDial         time.Duration
	P2PV2DeltaReconcile    time.Duration
	P2PV2ListenAddresses   []string
}

var ErrConfigNoValue = errors.New("field value is not specified")

func (n *NetworkingConfig) Validate() error {
	return nil
}

type Peer interface {
	ocrtypes.BootstrapperFactory
	ocrtypes.BinaryNetworkEndpointFactory
	Close() error
}

var _ Peer = &peerAdapter{}

type peerAdapter struct {
	ocrtypes.BootstrapperFactory
	ocrtypes.BinaryNetworkEndpointFactory
	io.Closer
}

type DiscovererDatabase interface {
	// StoreAnnouncement has key-value-store semantics and stores a peerID (key) and an associated serialized
	// announcement (value).
	StoreAnnouncement(ctx context.Context, peerID string, ann []byte) error

	// ReadAnnouncements returns one serialized announcement (if available) for each of the peerIDs in the form of a map
	// keyed by each announcement's corresponding peer ID.
	ReadAnnouncements(ctx context.Context, peerIDs []string) (map[string][]byte, error)
}

type Service interface {
	Peer() Peer
	IsStarted() bool
	Start() error
	Close() error
}

type peerService struct {
	cfg     NetworkingConfig
	peer    Peer
	peerKey p2pkey.Key
	peerID  p2pkey.PeerID
	peerDB  DiscovererDatabase

	runningMux sync.RWMutex
	running    bool
	onceStart  sync.Once
	onceStop   sync.Once

	logger log.Logger
}

type DBDriver interface {
	String() string
}

func NewService(
	key p2pkey.Key,
	cfg NetworkingConfig,
	peerDB DiscovererDatabase,
) (Service, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	peerID, err := key.GetPeerID()
	if err != nil {
		err = errors.Wrap(err, "failed to get PeerID from key")
		return nil, err
	}

	minLogLevel := log.InfoLevel
	if cfg.TraceLogging {
		minLogLevel = log.TraceLevel
	}

	svc := &peerService{
		cfg:     cfg,
		peer:    nil,
		peerKey: key,
		peerID:  peerID,

		peerDB: peerDB,

		logger: logging.NewSuplog(minLogLevel, false).WithFields(log.Fields{
			"svc": "p2p_peer",
		}),
	}

	return svc, nil
}

func (p *peerService) IsStarted() bool {
	p.runningMux.RLock()
	defer p.runningMux.RUnlock()

	return p.running
}

func (p *peerService) Start() (err error) {
	p.logger.Infoln("P2P Peer service starting")

	p.onceStart.Do(func() {
		p.runningMux.Lock()
		defer p.runningMux.Unlock()
		defer func() {
			p.running = true
		}()

		peerConfig := ocrnetworking.PeerConfig{
			NetworkingStack:      ocrnetworking.NetworkingStackV2,
			PrivKey:              p.peerKey,
			Logger:               logging.WrapCommonLogger(p.logger),
			V2ListenAddresses:    p.cfg.P2PV2ListenAddresses,
			V2AnnounceAddresses:  p.cfg.P2PV2AnnounceAddresses,
			V2DeltaReconcile:     p.cfg.P2PV2DeltaReconcile,
			V2DeltaDial:          p.cfg.P2PV2DeltaDial,
			V2DiscovererDatabase: p.peerDB,
			EndpointConfig: ocrnetworking.EndpointConfig{
				IncomingMessageBufferSize: p.cfg.IncomingMessageBufferSize,
				OutgoingMessageBufferSize: p.cfg.OutgoingMessageBufferSize,
				NewStreamTimeout:          p.cfg.NewStreamTimeout,
				DHTLookupInterval:         p.cfg.DHTLookupInterval,
				BootstrapCheckInterval:    p.cfg.BootstrapCheckInterval,
			},
		}

		peer, peerInitErr := ocrnetworking.NewPeer(peerConfig)
		if peerInitErr != nil {
			err = errors.Wrap(peerInitErr, "failed to init peer")
			return
		}

		p.peer = &peerAdapter{
			BinaryNetworkEndpointFactory: peer.OCR2BinaryNetworkEndpointFactory(),
			BootstrapperFactory:          peer.OCR2BootstrapperFactory(),
			Closer:                       peer,
		}
	})

	return err
}

func (p *peerService) Peer() Peer {
	p.runningMux.RLock()
	defer p.runningMux.RUnlock()

	if !p.running {
		return nil
	}

	return p.peer
}

func (p *peerService) Close() (err error) {
	p.logger.Infoln("P2P Peer service stopping")

	p.onceStop.Do(func() {
		p.runningMux.Lock()
		defer p.runningMux.Unlock()
		defer func() {
			p.running = false
		}()

		if err = p.peer.Close(); err != nil {
			err = errors.Wrap(err, "failed to close P2P Peer")
		}
	})

	return err
}
