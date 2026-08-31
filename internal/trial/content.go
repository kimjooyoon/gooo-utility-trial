package trial

import "encoding/json"

const fixtureText = `{
  "schema": "gooo/opentofu-envelope/plan-oracle/v1",
  "subject": "envelope://opentofu/intent/hello",
  "resource_actions": [{"address": "terraform_data.hello", "action": "create"}],
  "change_summary": {"add": 1, "change": 0, "forget": 0, "import": 0, "operation": "plan", "remove": 0},
  "side_effects": {"apply": 0, "cloud": 0, "network": 0, "provider": 0, "source_write": 0, "state_mutation": 0}
}`

const baselineRawText = `{
  "format": "opentofu-show-json",
  "resource_changes": [{"address": "terraform_data.hello", "change": {"actions": ["create"]}}],
  "decision_surface": "machine_resource_actions_only"
}`

const baselineEvidenceText = `{
  "claim_state": "UNKNOWN",
  "stage": "EXTERNAL_UTILITY",
  "step": "REQUIRE_INDEPENDENT_EXTERNAL_UTILITY_EVIDENCE",
  "reason": "EXTERNAL_UTILITY_NOT_OBSERVED",
  "evidence_count": 0,
  "frontier_available": false
}`

const dossierText = `# Gooo human dossier

This is a read-only decision aid derived from the immutable upstream Gooo
OpenTofu envelope release and its plan oracle fixture.

## U1 — OpenTofu drift judgment

The machine observation contains one planned action: create
` + "`terraform_data.hello`" + `.

Decision token: ` + "`DRIFT_PRESENT`" + `.

## U2 — Resume an UNKNOWN

The same causal state is UNKNOWN because independent external utility evidence
has not been observed. The complete six-field frontier is:

- stage: ` + "`EXTERNAL_UTILITY`" + `
- step: ` + "`REQUIRE_INDEPENDENT_EXTERNAL_UTILITY_EVIDENCE`" + `
- reason: ` + "`EXTERNAL_UTILITY_NOT_OBSERVED`" + `
- unknown_class: ` + "`CAUSALITY_UNPROVEN`" + `
- next_operation: ` + "`PROVIDE_INDEPENDENT_EXTERNAL_UTILITY_EVIDENCE`" + `
- blocked_by: ` + "`exact-before-after-utility-pair`" + `

Decision token: ` + "`PROVIDE_INDEPENDENT_EXTERNAL_UTILITY_EVIDENCE`" + `.

Natural-language reactions and ratings are not evidence of correctness. A
session receipt records only the exact decision digest, correctness integer,
observed monotonic timing, abandonment state, and the required provenance.
`

const goooProgramText = `package utilitytrial
namespace utilitytrial

entity TrialInput id "trial://utility/input"
entity TrialOutput id "trial://utility/output"
entity FoundationCell id "trial://utility/cell/foundation"
entity CoherenceCell id "trial://utility/cell/coherence"
entity RegressionCell id "trial://utility/cell/regression"
entity TrialDecision id "trial://utility/decision"
entity ReceiptProof id "trial://utility/receipt-proof"
entity PairProof id "trial://utility/pair-proof"
entity OriginGuard id "trial://utility/origin-guard"
entity OracleBinding id "trial://utility/oracle-binding"
entity FrontierResume id "trial://utility/frontier-resume"
entity ReleaseAnchor id "trial://utility/release-anchor"

activity DeclareTrialInput(TrialInput) -> TrialOutput
activity BindImmutableRelease(TrialOutput) -> ReleaseAnchor
activity BindPlanOracle(ReleaseAnchor) -> OracleBinding
activity EmitHumanDossier(OracleBinding) -> TrialDecision
activity PreserveUnknownFrontier(TrialDecision) -> FrontierResume
activity BindSessionReceipt(FrontierResume) -> ReceiptProof
activity VerifyPairIdentity(ReceiptProof) -> PairProof
activity VerifyExternalOrigin(PairProof) -> OriginGuard
activity CloseFoundationCell(OriginGuard) -> FoundationCell
activity CloseCoherenceCell(FoundationCell) -> CoherenceCell
activity ClassifyRefutedCell(CoherenceCell) -> RegressionCell
activity ReportUtility(RegressionCell) -> TrialOutput
`

func fixtureValue() (any, error) {
	var value any
	if err := json.Unmarshal([]byte(fixtureText), &value); err != nil {
		return nil, err
	}
	return value, nil
}

func baselineRawValue() (any, error) {
	var value any
	if err := json.Unmarshal([]byte(baselineRawText), &value); err != nil {
		return nil, err
	}
	return value, nil
}

func baselineEvidenceValue() (any, error) {
	var value any
	if err := json.Unmarshal([]byte(baselineEvidenceText), &value); err != nil {
		return nil, err
	}
	return value, nil
}

