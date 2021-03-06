package ocr2

import (
	"context"
	"encoding/hex"
	"time"

	"github.com/smartcontractkit/libocr/commontypes"
	ocrtypes "github.com/smartcontractkit/libocr/offchainreporting2/types"

	"github.com/InjectiveLabs/chainlink-injective/db"
	"github.com/InjectiveLabs/chainlink-injective/db/model"
)

type JobStateDB interface {
	ocrtypes.Database
}

var _ JobStateDB = &jobDBWrapper{}

type jobDBWrapper struct {
	svc db.JobDBService
}

func NewJobDBWrapper(dbSvc db.JobDBService) JobStateDB {
	return &jobDBWrapper{
		svc: dbSvc,
	}
}

func (j *jobDBWrapper) WriteState(ctx context.Context, configDigest ocrtypes.ConfigDigest, state ocrtypes.PersistentState) error {
	result := &model.JobPersistentState{
		JobID:                j.svc.JobID(),
		ConfigDigest:         model.ID(configDigestHex(configDigest)),
		Epoch:                state.Epoch,
		HighestSentEpoch:     state.HighestSentEpoch,
		HighestReceivedEpoch: state.HighestReceivedEpoch,
	}

	return j.svc.SetPersistentState(ctx, result)
}

func (j *jobDBWrapper) ReadState(ctx context.Context, configDigest ocrtypes.ConfigDigest) (*ocrtypes.PersistentState, error) {
	state, err := j.svc.GetPersistentState(ctx, model.ID(configDigestHex(configDigest)))
	if err != nil {
		if err == db.ErrNotFound {
			return nil, nil
		}

		return nil, err
	}

	result := &ocrtypes.PersistentState{
		Epoch:                state.Epoch,
		HighestSentEpoch:     state.HighestSentEpoch,
		HighestReceivedEpoch: state.HighestReceivedEpoch,
	}

	return result, nil
}

func (j *jobDBWrapper) WriteConfig(ctx context.Context, config ocrtypes.ContractConfig) error {
	result := &model.JobContractConfig{
		JobID:                 j.svc.JobID(),
		ConfigDigest:          model.ID(configDigestHex(config.ConfigDigest)),
		ConfigCount:           config.ConfigCount,
		Signers:               make([]model.HexBytes, 0, len(config.Signers)),
		Transmitters:          make([]model.Account, 0, len(config.Transmitters)),
		F:                     config.F,
		OnchainConfig:         config.OnchainConfig,
		OffchainConfigVersion: config.OffchainConfigVersion,
		OffchainConfig:        config.OffchainConfig,
	}

	for _, signer := range config.Signers {
		result.Signers = append(result.Signers, model.HexBytes(signer))
	}

	for _, transmitter := range config.Transmitters {
		result.Transmitters = append(result.Transmitters, model.Account(transmitter))
	}

	return j.svc.SetContractConfig(ctx, result)
}

func (j *jobDBWrapper) ReadConfig(ctx context.Context) (*ocrtypes.ContractConfig, error) {
	config, err := j.svc.GetContractConfig(ctx)
	if err != nil {
		if err == db.ErrNotFound {
			return nil, nil
		}

		return nil, err
	}

	result := &ocrtypes.ContractConfig{
		ConfigDigest:          hexToConfigDigest(string(config.ConfigDigest)),
		ConfigCount:           config.ConfigCount,
		Signers:               make([]ocrtypes.OnchainPublicKey, 0, len(config.Signers)),
		Transmitters:          make([]ocrtypes.Account, 0, len(config.Transmitters)),
		F:                     config.F,
		OnchainConfig:         config.OnchainConfig,
		OffchainConfigVersion: config.OffchainConfigVersion,
		OffchainConfig:        config.OffchainConfig,
	}

	for _, signer := range config.Signers {
		result.Signers = append(result.Signers, ocrtypes.OnchainPublicKey(signer))
	}

	for _, transmitter := range config.Transmitters {
		result.Transmitters = append(result.Transmitters, ocrtypes.Account(transmitter))
	}

	return result, nil
}

func (j *jobDBWrapper) StorePendingTransmission(ctx context.Context, reportTimestamp ocrtypes.ReportTimestamp, tx ocrtypes.PendingTransmission) error {
	pendingTransmission := &model.JobPendingTransmission{
		JobID:        j.svc.JobID(),
		ConfigDigest: model.ID(reportTimestamp.ConfigDigest.Hex()),
		ReportTimestamp: model.ReportTimestamp{
			Epoch: reportTimestamp.Epoch,
			Round: reportTimestamp.Round,
		},
		Transmission: model.PendingTransmission{
			Time:                 tx.Time,
			ExtraHash:            model.HexBytes(tx.ExtraHash[:]),
			Report:               model.HexBytes(tx.Report),
			AttributedSignatures: make([]model.AttributedOnchainSignature, 0, len(tx.AttributedSignatures)),
		},

		CreatedAt: time.Now().UTC(),
	}

	for _, sig := range tx.AttributedSignatures {
		pendingTransmission.Transmission.AttributedSignatures = append(pendingTransmission.Transmission.AttributedSignatures,
			model.AttributedOnchainSignature{
				Signature: model.HexBytes(sig.Signature),
				Signer:    int(sig.Signer),
			})
	}

	return j.svc.InsertPendingTranmission(ctx, pendingTransmission)
}

func (j *jobDBWrapper) PendingTransmissionsWithConfigDigest(ctx context.Context, configDigest ocrtypes.ConfigDigest) (map[ocrtypes.ReportTimestamp]ocrtypes.PendingTransmission, error) {
	transmissions, err := j.svc.ListPendingTransmissions(ctx, model.ID(hex.EncodeToString(configDigest[:])), &model.Cursor{
		Limit: 10000,
	})
	if err != nil {
		return nil, err
	}

	result := make(map[ocrtypes.ReportTimestamp]ocrtypes.PendingTransmission, len(transmissions))
	for _, tx := range transmissions {
		pendingTransmission := ocrtypes.PendingTransmission{
			Time:                 tx.Transmission.Time,
			Report:               ocrtypes.Report(tx.Transmission.Report),
			AttributedSignatures: make([]ocrtypes.AttributedOnchainSignature, 0, len(tx.Transmission.AttributedSignatures)),
		}

		for _, sig := range tx.Transmission.AttributedSignatures {
			pendingTransmission.AttributedSignatures = append(pendingTransmission.AttributedSignatures, ocrtypes.AttributedOnchainSignature{
				Signature: sig.Signature,
				Signer:    commontypes.OracleID(sig.Signer),
			})
		}

		n := copy(pendingTransmission.ExtraHash[:], tx.Transmission.ExtraHash)
		if n != len(pendingTransmission.ExtraHash) {
			panic("short read")
		}

		result[ocrtypes.ReportTimestamp{
			Epoch: tx.ReportTimestamp.Epoch,
			Round: tx.ReportTimestamp.Round,
		}] = pendingTransmission
	}

	return result, nil
}

func (j *jobDBWrapper) DeletePendingTransmission(ctx context.Context, reportTimestamp ocrtypes.ReportTimestamp) error {
	return j.svc.DeletePendingTransmission(ctx, model.ReportTimestamp{
		Epoch: reportTimestamp.Epoch,
		Round: reportTimestamp.Round,
	})
}

func (j *jobDBWrapper) DeletePendingTransmissionsOlderThan(ctx context.Context, timestamp time.Time) error {
	return j.svc.DeletePendingTransmissionsOlderThan(ctx, timestamp)
}

func configDigestHex(configDigest [32]byte) string {
	return hex.EncodeToString(configDigest[:])
}

func hexToConfigDigest(digestHex string) [32]byte {
	b, err := hex.DecodeString(digestHex)
	if err != nil {
		panic(err)
	}

	var result [32]byte
	if copy(result[:], b) != len(result) {
		panic("short read")
	}

	return result
}
