package search

import "context"

// ACBSTraceEvent records one scheduler chunk from a traced ACBS run. Tracing
// is opt-in and is never enabled on the normal routing or benchmark hot path.
type ACBSTraceEvent struct {
	Chunk                     uint64  `json:"chunk"`
	Direction                 string  `json:"direction"`
	Budget                    int     `json:"budget"`
	BeforeLowerBound          uint64  `json:"beforeLowerBound"`
	AfterLowerBound           uint64  `json:"afterLowerBound"`
	LowerBoundGain            uint64  `json:"lowerBoundGain"`
	Work                      uint64  `json:"work"`
	RelaxedDelta              uint64  `json:"relaxedDelta"`
	ExpandedDelta             uint64  `json:"expandedDelta"`
	QueuePopsDelta            uint64  `json:"queuePopsDelta"`
	StalePopsDelta            uint64  `json:"stalePopsDelta"`
	ConnectionChecksDelta     uint64  `json:"connectionChecksDelta"`
	FiniteMeetingsDelta       uint64  `json:"finiteMeetingsDelta"`
	PotentialEvaluationsDelta uint64  `json:"potentialEvaluationsDelta"`
	ForwardQueueBefore        int     `json:"forwardQueueBefore"`
	BackwardQueueBefore       int     `json:"backwardQueueBefore"`
	ForwardQueueAfter         int     `json:"forwardQueueAfter"`
	BackwardQueueAfter        int     `json:"backwardQueueAfter"`
	ForwardPriorityAfter      uint64  `json:"forwardPriorityAfter"`
	BackwardPriorityAfter     uint64  `json:"backwardPriorityAfter"`
	ForwardScoreBefore        float64 `json:"forwardScoreBefore"`
	BackwardScoreBefore       float64 `json:"backwardScoreBefore"`
	ForwardScoreAfter         float64 `json:"forwardScoreAfter"`
	BackwardScoreAfter        float64 `json:"backwardScoreAfter"`
	DirectionSwitchesTotal    uint64  `json:"directionSwitchesTotal"`
	ForwardExpandedTotal      uint64  `json:"forwardExpandedTotal"`
	BackwardExpandedTotal     uint64  `json:"backwardExpandedTotal"`
	HadUpperBoundBefore       bool    `json:"hadUpperBoundBefore"`
	HadUpperBoundAfter        bool    `json:"hadUpperBoundAfter"`
	UpperBoundBefore          uint64  `json:"upperBoundBefore,omitempty"`
	UpperBoundAfter           uint64  `json:"upperBoundAfter,omitempty"`
	LateGuardActive           bool    `json:"lateGuardActive,omitempty"`
	LateGuardTriggered        bool    `json:"lateGuardTriggered,omitempty"`
	ConnectionGuardActive     bool    `json:"connectionGuardActive,omitempty"`
	ConnectionGuardTriggered  bool    `json:"connectionGuardTriggered,omitempty"`
	ConnectionGuardMode       string  `json:"connectionGuardMode,omitempty"`
}

type acbsTraceKey struct{}

// WithACBSTrace attaches an ACBS chunk observer to ctx. The sink must not call
// back into the same search synchronously. It is intended for deterministic
// replay diagnostics, not production request logging.
func WithACBSTrace(ctx context.Context, sink func(ACBSTraceEvent)) context.Context {
	if sink == nil {
		return ctx
	}
	return context.WithValue(ctx, acbsTraceKey{}, sink)
}

func acbsTraceFromContext(ctx context.Context) func(ACBSTraceEvent) {
	if ctx == nil {
		return nil
	}
	sink, _ := ctx.Value(acbsTraceKey{}).(func(ACBSTraceEvent))
	return sink
}
