package median

import (
	"context"
	"math/big"
	"sort"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	proto "github.com/gogo/protobuf/proto"
	"github.com/pkg/errors"
	"github.com/smartcontractkit/libocr/commontypes"
	"github.com/smartcontractkit/libocr/offchainreporting2/types"
	"github.com/smartcontractkit/libocr/subprocesses"

	"github.com/InjectiveLabs/chainlink-injective/ocr2/reportingplugin/median/loghelper"
)

type MedianContract interface {
	LatestTransmissionDetails(
		ctx context.Context,
	) (
		configDigest types.ConfigDigest,
		epoch uint32,
		round uint8,
		latestAnswer *big.Int,
		latestTimestamp time.Time,
		err error,
	)

	// LatestRoundRequested returns the configDigest, epoch, and round from the latest
	// RoundRequested event emitted by the contract. LatestRoundRequested may or may not
	// return a result if the latest such event was emitted in a block b such that
	// b.timestamp < tip.timestamp - lookback.
	//
	// If no event is found, LatestRoundRequested should return zero values, not an error.
	// An error should only be returned if an actual error occurred during execution,
	// e.g. because there was an error querying the blockchain or the database.
	//
	// As an optimization, this function may also return zero values, if no
	// RoundRequested event has been emitted after the latest NewTransmission event.
	LatestRoundRequested(
		ctx context.Context,
		lookback time.Duration,
	) (
		configDigest types.ConfigDigest,
		epoch uint32,
		round uint8,
		err error,
	)
}

// DataSource implementations must be thread-safe. Observe may be called by many different threads concurrently.
type DataSource interface {
	// Observe queries the data source. Returns a value or an error. Once the
	// context is expires, Observe may still do cheap computations and return a
	// result, but should return as quickly as possible.
	//
	// More details: In the current implementation, the context passed to
	// Observe will time out after LocalConfig.DataSourceTimeout. However,
	// Observe should *not* make any assumptions about context timeout behavior.
	// Once the context times out, Observe should prioritize returning as
	// quickly as possible, but may still perform fast computations to return a
	// result rather than errror. For example, if Observe medianizes a number
	// of data sources, some of which already returned a result to Observe prior
	// to the context's expiry, Observe might still compute their median, and
	// return it instead of an error.
	//
	// Important: Observe should not perform any potentially time-consuming
	// actions like database access, once the context passed has expired.
	Observe(context.Context) (*big.Int, error)
}

type NumericalMedianFactory struct {
	ContractTransmitter MedianContract
	DataSource          DataSource
	Logger              commontypes.Logger
}

var _ types.ReportingPluginFactory = NumericalMedianFactory{}

func (fac NumericalMedianFactory) NewReportingPlugin(
	configuration types.ReportingPluginConfig,
) (types.ReportingPlugin, types.ReportingPluginInfo, error) {
	config, err := DecodeConfig(configuration.OffchainConfig)
	if err != nil {
		return nil, types.ReportingPluginInfo{}, err
	}

	logger := loghelper.MakeRootLoggerWithContext(fac.Logger).MakeChild(commontypes.LogFields{
		"configDigest":    configuration.ConfigDigest,
		"reportingPlugin": "NumericalMedian",
	})

	plugin := &numericalMedian{
		Config:              config,
		ContractTransmitter: fac.ContractTransmitter,
		DataSource:          fac.DataSource,
		Logger:              logger,

		configDigest:             configuration.ConfigDigest,
		latestAcceptedEpochRound: epochRound{},
		latestAcceptedMedian:     new(big.Int),
	}

	pluginInfo := types.ReportingPluginInfo{
		Name: "NumericalMedian",

		UniqueReports: false,

		MaxQueryLen:       0,
		MaxObservationLen: 64 * 1024, // TODO: sanity-check values against proto impl
		MaxReportLen:      64 * 1024,
	}

	return plugin, pluginInfo, nil
}

// TODO(lorenz): pass config into logic

func deviates(thresholdPPB uint64, old *big.Int, new *big.Int) bool {
	if old.Cmp(big.NewInt(0)) == 0 {
		if new.Cmp(big.NewInt(0)) == 0 {
			return false // Both values are zero; no deviation
		}

		return true // Any deviation from 0 is significant
	}

	// ||new - old|| / ||old||, approximated by a float
	change := &big.Rat{}
	change.SetFrac(big.NewInt(0).Sub(new, old), old)
	change.Abs(change)
	threshold := &big.Rat{}
	threshold.SetFrac(
		(&big.Int{}).SetUint64(thresholdPPB),
		(&big.Int{}).SetUint64(1000000000),
	)

	return change.Cmp(threshold) > 0
}

var _ types.ReportingPlugin = &numericalMedian{}

type numericalMedian struct {
	Config              OffchainConfig
	ContractTransmitter MedianContract
	DataSource          DataSource
	Logger              loghelper.LoggerWithContext

	configDigest             types.ConfigDigest
	latestAcceptedEpochRound epochRound
	latestAcceptedMedian     *big.Int
}

func (nm *numericalMedian) Query(
	ctx context.Context,
	eportTimestamps types.ReportTimestamp,
) (types.Query, error) {
	return nil, nil
}

func (nm *numericalMedian) Observation(
	ctx context.Context,
	reportTimestamps types.ReportTimestamp,
	query types.Query,
) (types.Observation, error) {
	if len(query) != 0 {
		return nil, errors.New("expected empty query")
	}

	value, err := nm.DataSource.Observe(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "error during DataSource.Observe")
	} else if value == nil {
		return nil, errors.New("DataSource.Observe returned nil big.Int which should never happen")
	}

	data, err := proto.Marshal(&MedianObservation{
		Timestamp: time.Now().Unix(),
		Value:     value.Bytes(),
	})
	if err != nil {
		err = errors.Wrap(err, "failed to marshal MedianObservation message")
		return nil, err
	}

	return types.Observation(data), nil
}

type attributedObservation struct {
	Timestamp int64
	Value     *big.Int
	Observer  commontypes.OracleID
}

func unmarshalAttributedObservations(attributedObservations []types.AttributedObservation, skipBad bool) ([]attributedObservation, error) {
	observations := make([]attributedObservation, 0, len(attributedObservations))

	for idx, attributed := range attributedObservations {
		var observation MedianObservation

		if err := proto.Unmarshal(attributed.Observation, &observation); err != nil {
			if skipBad {
				continue
			} else {
				return nil, errors.Wrapf(err, "Error while unmarshalling %v-th attributed observation", idx)
			}
		}

		observations = append(observations, attributedObservation{
			Timestamp: observation.Timestamp,
			Value:     new(big.Int).SetBytes(observation.Value),
			Observer:  attributed.Observer,
		})
	}

	return observations, nil
}

func (nm *numericalMedian) Report(
	ctx context.Context,
	reportTimestamps types.ReportTimestamp,
	query types.Query,
	attributedObservations []types.AttributedObservation,
) (bool, types.Report, error) {
	if len(query) != 0 {
		return false, nil, errors.New("expected empty query")
	}

	observations, err := unmarshalAttributedObservations(attributedObservations, true)
	if err != nil {
		return false, nil, errors.Wrap(err, "error in unmarshalAttributedObservations")
	}

	should, err := nm.shouldReport(ctx, reportTimestamps, observations)
	if err != nil {
		return false, nil, err
	} else if !should {
		return false, nil, nil
	}

	report, err := nm.buildReport(ctx, observations)
	if err != nil {
		return false, nil, err
	}

	return true, report, nil
}

func (nm *numericalMedian) shouldReport(
	ctx context.Context,
	reportTimestamps types.ReportTimestamp,
	observations []attributedObservation,
) (bool, error) {
	if len(observations) == 0 {
		return false, errors.New("cannot handle empty attributed observations")
	}

	var resultTransmissionDetails struct {
		configDigest    types.ConfigDigest
		epoch           uint32
		round           uint8
		latestAnswer    *big.Int
		latestTimestamp time.Time
		err             error
	}

	var resultRoundRequested struct {
		configDigest types.ConfigDigest
		epoch        uint32
		round        uint8
		err          error
	}

	var subs subprocesses.Subprocesses

	subs.Go(func() {
		resultTransmissionDetails.configDigest,
			resultTransmissionDetails.epoch,
			resultTransmissionDetails.round,
			resultTransmissionDetails.latestAnswer,
			resultTransmissionDetails.latestTimestamp,
			resultTransmissionDetails.err =
			nm.ContractTransmitter.LatestTransmissionDetails(ctx)
	})

	subs.Go(func() {
		resultRoundRequested.configDigest,
			resultRoundRequested.epoch,
			resultRoundRequested.round,
			resultRoundRequested.err =
			nm.ContractTransmitter.LatestRoundRequested(ctx, nm.Config.DeltaC)
	})

	subs.Wait()

	if resultTransmissionDetails.err != nil {
		return true, errors.Wrap(resultTransmissionDetails.err, "error during LatestTransmissionDetails")
	}

	if resultRoundRequested.err != nil {
		return true, errors.Wrap(resultTransmissionDetails.err, "error during LatestRoundRequested")
	}

	// sort by values
	sort.Slice(observations, func(i, j int) bool {
		return observations[i].Value.Cmp(observations[j].Value) < 0
	})

	answer := observations[len(observations)/2].Value

	initialRound := // Is this the first round for this configuration?
		resultTransmissionDetails.configDigest == reportTimestamps.ConfigDigest &&
			resultTransmissionDetails.epoch == 0 &&
			resultTransmissionDetails.round == 0

	deviation := // Has the result changed enough to merit a new report?
		deviates(nm.Config.AlphaPPB, resultTransmissionDetails.latestAnswer, answer)

	// TODO: would it make sense to compare with observationsTimestamp here?
	deltaCTimeout := // Has enough time passed since the last report, to merit a new one?
		resultTransmissionDetails.latestTimestamp.Add(nm.Config.DeltaC).
			Before(time.Now())

	unfulfilledRequest := // Has a new report been requested explicitly?
		resultRoundRequested.configDigest == reportTimestamps.ConfigDigest &&
			!(epochRound{resultRoundRequested.epoch, resultRoundRequested.round}).
				Less(epochRound{resultTransmissionDetails.epoch, resultTransmissionDetails.round})

	logger := nm.Logger.MakeChild(commontypes.LogFields{
		"timestamp":                 reportTimestamps,
		"initialRound":              initialRound,
		"alphaPPB":                  nm.Config.AlphaPPB,
		"deviation":                 deviation,
		"deltaC":                    nm.Config.DeltaC,
		"deltaCTimeout":             deltaCTimeout,
		"lastTransmissionTimestamp": resultTransmissionDetails.latestTimestamp,
		"unfulfilledRequest":        unfulfilledRequest,
	})

	// The following is more succinctly expressed as a disjunction, but breaking
	// the branches up into their own conditions makes it easier to check that
	// each branch is tested, and also allows for more expressive log messages
	if initialRound {
		logger.Info("shouldReport: yes, because it's the first round of the first epoch", commontypes.LogFields{
			"result": true,
		})

		return true, nil
	}
	if deviation {
		logger.Info("shouldReport: yes, because new median deviates sufficiently from current onchain value", commontypes.LogFields{
			"result": true,
		})

		return true, nil
	}
	if deltaCTimeout {
		logger.Info("shouldReport: yes, because deltaC timeout since last onchain report", commontypes.LogFields{
			"result": true,
		})

		return true, nil
	}
	if unfulfilledRequest {
		logger.Info("shouldReport: yes, because a new report has been explicitly requested", commontypes.LogFields{
			"result": true,
		})

		return true, nil
	}

	logger.Info("shouldReport: no", commontypes.LogFields{"result": false})
	return false, nil
}

func (nm *numericalMedian) buildReport(ctx context.Context, observations []attributedObservation) (types.Report, error) {
	if len(observations) == 0 {
		return nil, errors.New("Cannot build report from empty attributed observations")
	}

	// get median timestamp
	sort.Slice(observations, func(i, j int) bool {
		return observations[i].Timestamp < observations[j].Timestamp
	})

	timestamp := observations[len(observations)/2].Timestamp

	// sort by values
	sort.Slice(observations, func(i, j int) bool {
		return observations[i].Value.Cmp(observations[j].Value) < 0
	})

	observers := make([]byte, 0, 32)
	observationValues := make([]sdk.Dec, 0, 32)

	for _, attributed := range observations {
		observers = append(observers, byte(attributed.Observer))
		observationValues = append(observationValues, sdk.NewDecFromBigInt(attributed.Value))
	}

	reportBytes, err := proto.Marshal(&Report{
		ObservationsTimestamp: timestamp,
		Observers:             observers,
		Observations:          observationValues,
	})
	if err != nil {
		err = errors.Wrap(err, "failed to marshal MedianObservation message")
		return nil, err
	}

	return types.Report(reportBytes), err
}

func (nm *numericalMedian) ShouldAcceptFinalizedReport(
	ctx context.Context,
	reportTimestamps types.ReportTimestamp,
	report types.Report,
) (bool, error) {
	reportEpochRound := epochRound{reportTimestamps.Epoch, reportTimestamps.Round}
	if !nm.latestAcceptedEpochRound.Less(reportEpochRound) {
		nm.Logger.Debug("ShouldAcceptFinalizedReport() = false, report is stale", commontypes.LogFields{
			"latestAcceptedEpochRound": nm.latestAcceptedEpochRound,
			"reportEpochRound":         reportEpochRound,
		})

		return false, nil
	}

	contractConfigDigest, contractEpoch, contractRound, _, _, err := nm.ContractTransmitter.LatestTransmissionDetails(ctx)
	if err != nil {
		return false, err
	}

	contractEpochRound := epochRound{contractEpoch, contractRound}

	if contractConfigDigest != nm.configDigest {
		nm.Logger.Debug("ShouldAcceptFinalizedReport() = false, config digest mismatch", commontypes.LogFields{
			"contractConfigDigest": contractConfigDigest,
			"reportConfigDigest":   nm.configDigest,
			"reportEpochRound":     reportEpochRound,
		})

		return false, nil
	}

	if !contractEpochRound.Less(reportEpochRound) {
		nm.Logger.Debug("ShouldAcceptFinalizedReport() = false, report is stale", commontypes.LogFields{
			"contractEpochRound": contractEpochRound,
			"reportEpochRound":   reportEpochRound,
		})

		return false, nil
	}

	reportMedian, err := medianFromReport(report)
	if err != nil {
		return false, errors.Wrap(err, "error during medianFromReport")
	}

	deviates := deviates(nm.Config.AlphaPPB, nm.latestAcceptedMedian, reportMedian)
	nothingPending := !contractEpochRound.Less(nm.latestAcceptedEpochRound)
	result := deviates || nothingPending

	nm.Logger.Debug("ShouldAcceptFinalizedReport() = result", commontypes.LogFields{
		"contractEpochRound":       contractEpochRound,
		"reportEpochRound":         reportEpochRound,
		"latestAcceptedEpochRound": nm.latestAcceptedEpochRound,
		"deviates":                 deviates,
		"result":                   result,
	})

	if result {
		nm.latestAcceptedEpochRound = reportEpochRound
		nm.latestAcceptedMedian = reportMedian
	}

	return result, nil
}

func medianFromReport(reportBytes types.Report) (*big.Int, error) {
	var report Report

	err := proto.Unmarshal([]byte(reportBytes), &report)
	if err != nil {
		return nil, errors.Wrap(err, "error during report unmarshalling")
	}

	if len(report.Observations) == 0 {
		return nil, errors.New("observations are empty")
	}

	median := report.Observations[len(report.Observations)/2]
	if median.IsNil() {
		return nil, errors.New("median is nil")
	}

	return median.BigInt(), nil
}

func (nm *numericalMedian) ShouldTransmitAcceptedReport(
	ctx context.Context,
	reportTimestamps types.ReportTimestamp,
	report types.Report,
) (bool, error) {
	reportEpochRound := epochRound{reportTimestamps.Epoch, reportTimestamps.Round}

	contractConfigDigest, contractEpoch, contractRound, _, _, err := nm.ContractTransmitter.LatestTransmissionDetails(ctx)
	if err != nil {
		return false, err
	}

	contractEpochRound := epochRound{contractEpoch, contractRound}

	if contractConfigDigest != nm.configDigest {
		nm.Logger.Debug("ShouldTransmitAcceptedReport() = false, config digest mismatch", commontypes.LogFields{
			"contractConfigDigest": contractConfigDigest,
			"reportConfigDigest":   nm.configDigest,
			"reportEpochRound":     reportEpochRound,
		})

		return false, nil
	}

	if !contractEpochRound.Less(reportEpochRound) {
		nm.Logger.Debug("ShouldTransmitAcceptedReport() = false, report is stale", commontypes.LogFields{
			"contractEpochRound": contractEpochRound,
			"reportEpochRound":   reportEpochRound,
		})

		return false, nil
	}

	// TODO: Should we log here?

	return true, nil
}

func (nm *numericalMedian) Start() error {
	return nil
}

func (nm *numericalMedian) Close() error {
	return nil
}

type epochRound struct {
	Epoch uint32
	Round uint8
}

func (x epochRound) Less(y epochRound) bool {
	return x.Epoch < y.Epoch || (x.Epoch == y.Epoch && x.Round < y.Round)
}
