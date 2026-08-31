package trial

func Cells() []Cell {
	return []Cell{
		{1, "TRIAL_INPUT_DECLARED", "DeclareIndependentTrialInput", "INPUT", "DECLARE_TRIAL_INPUT", "FOUNDATION", "DRIVER", "CLOSED", nil},
		{2, "RELEASE_ANCHORED", "AnchorImmutableGoooRelease", "RELEASE", "ANCHOR_IMMUTABLE_GOOO_RELEASE", "FOUNDATION", "DRIVER", "CLOSED", []string{"TRIAL_INPUT_DECLARED"}},
		{3, "FIXTURE_BOUND", "BindOpenTofuPlanOracleFixture", "FIXTURE", "BIND_OPENTOFU_PLAN_ORACLE", "FOUNDATION", "DRIVER", "CLOSED", []string{"RELEASE_ANCHORED"}},
		{4, "PACKET_GENERATED", "GenerateCallerOwnedTrialPacket", "GENERATION", "GENERATE_CALLER_OWNED_PACKET", "FOUNDATION", "DRIVER", "UNKNOWN", []string{"FIXTURE_BOUND"}},
		{5, "BASELINE_PATH_FIXED", "FixBaselineMachinePath", "DESIGN", "FIX_BASELINE_RAW_MACHINE_PATH", "COHERENCE", "OUTCOME", "UNKNOWN", []string{"PACKET_GENERATED"}},
		{6, "GOOO_PATH_FIXED", "FixGoooHumanPath", "DESIGN", "FIX_GOOO_HUMAN_DOSSIER_PATH", "COHERENCE", "OUTCOME", "UNKNOWN", []string{"PACKET_GENERATED"}},
		{7, "ORACLE_FIXED", "FixExactDecisionOracle", "ORACLE", "FIX_EXACT_DECISION_ORACLE", "COHERENCE", "OUTCOME", "REFUTED", []string{"BASELINE_PATH_FIXED", "GOOO_PATH_FIXED"}},
		{8, "COUNTERBALANCE_FIXED", "AssignDeterministicCounterbalance", "ASSIGNMENT", "ASSIGN_DETERMINISTIC_COUNTERBALANCE", "COHERENCE", "OUTCOME", "REFUTED", []string{"ORACLE_FIXED"}},
		{9, "RECEIPT_GUARDED", "ValidateSessionReceipt", "RECEIPT", "VALIDATE_SESSION_RECEIPT", "REGRESSION", "GUARDRAIL", "REFUTED", []string{"COUNTERBALANCE_FIXED"}},
		{10, "PAIR_GUARDED", "RequireMatchedBeforeAfterPair", "PAIR", "REQUIRE_MATCHED_BEFORE_AFTER_PAIR", "REGRESSION", "GUARDRAIL", "REFUTED", []string{"RECEIPT_GUARDED"}},
		{11, "ORIGIN_GUARDED", "ExcludeNonExternalUtilityEvidence", "ORIGIN", "EXCLUDE_NON_EXTERNAL_UTILITY_EVIDENCE", "REGRESSION", "GUARDRAIL", "REFUTED", []string{"RECEIPT_GUARDED"}},
		{12, "UTILITY_REPORTED", "ReportExternalUtilityBoundary", "REPORT", "REPORT_EXTERNAL_UTILITY_BOUNDARY", "REGRESSION", "GUARDRAIL", "REFUTED", []string{"PAIR_GUARDED", "ORIGIN_GUARDED"}},
	}
}

func countDenominator() Counts {
	counts := Counts{Cells: 12, Activities: 12, Precedence: []string{"REFUTED", "UNKNOWN", "CLOSED"}}
	for _, cell := range Cells() {
		switch cell.ProofFamily {
		case "FOUNDATION":
			counts.ProofFoundation++
		case "COHERENCE":
			counts.ProofCoherence++
		case "REGRESSION":
			counts.ProofRegression++
		}
		switch cell.Indicator {
		case "DRIVER":
			counts.IndicatorDriver++
		case "OUTCOME":
			counts.IndicatorOutcome++
		case "GUARDRAIL":
			counts.IndicatorGuardrail++
		}
		switch cell.CaseState {
		case "CLOSED":
			counts.Closed++
		case "UNKNOWN":
			counts.Unknown++
		case "REFUTED":
			counts.Refuted++
		}
	}
	return counts
}
