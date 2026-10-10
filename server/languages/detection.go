// SPDX-License-Identifier: AGPL-3.0-or-later

package languages

// Detection records evidence for one language decision. Confidence values are
// scores relative to the configured candidates, not calibrated probabilities.
// Language is unknown when the evidence does not justify selecting Candidate.
// This result is used during processing and is not stored in document metadata.
type Detection struct {
	Language           string
	Candidate          string
	Confidence         float64
	RunnerUp           string
	RunnerUpConfidence float64
	Margin             float64
	Letters            int
	Reason             string
}

const (
	ReasonAccepted     = "accepted"
	ReasonInsufficient = "insufficient_text"
	ReasonConfidence   = "low_confidence"
	ReasonAmbiguous    = "ambiguous"
	ReasonNoScores     = "no_scores"
	ReasonDisabled     = "disabled"
)
