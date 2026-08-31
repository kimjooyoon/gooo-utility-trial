package trial

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	fixturePath             = "gooo-pack/fixture.json"
	goooReleasePath         = "gooo-pack/release-reference.json"
	dossierPath             = "gooo-pack/human-dossier.md"
	baselineRawPath         = "baseline-pack/u1-baseline-raw.json"
	baselineEvidencePath    = "baseline-pack/u2-baseline-evidence.json"
	assignmentsPath         = "assignments.ndjson"
	oraclePath              = "oracle.json"
	contractPath            = "contract.json"
	receiptSchemaPath       = "session-receipt.schema.json"
	validatedEvidencePath   = "validated-evidence.ndjson"
	rejectedReceiptsPath    = "rejected-receipts.ndjson"
	runtimeMetricsPath      = "runtime-metrics.json"
	utilityReportPath       = "utility-report.md"
)

func contractValue() map[string]any {
	return map[string]any{
		"schema": ContractSchema,
		"version": 1,
		"target_cells": 12,
		"target_activities": 12,
		"proof_distribution": map[string]int{"FOUNDATION": 4, "COHERENCE": 4, "REGRESSION": 4},
		"indicator_distribution": map[string]int{"DRIVER": 4, "OUTCOME": 4, "GUARDRAIL": 4},
		"case_distribution": map[string]int{"CLOSED": 3, "UNKNOWN": 3, "REFUTED": 6},
		"precedence": []string{"REFUTED", "UNKNOWN", "CLOSED"},
		"cells": Cells(),
		"unknown_tuple": []string{"stage", "step", "reason", "unknown_class", "next_operation", "blocked_by"},
		"eligible_origins": []string{OriginExternal, OriginIndependent},
		"protocol_test_origins": []string{OriginCI, OriginMaintainer, OriginSynthetic},
	}
}

func buildOracle(fixtureDigest, rawDigest, evidenceDigest, dossierDigest string) Oracle {
	return Oracle{
		Schema: ProtocolSchema + "/oracle/v1",
		FixtureDigest: fixtureDigest,
		Tasks: []OracleTask{
			{
				TaskID: TaskU1,
				CausalState: "OPENTOFU_PLAN_OBSERVES_CREATE_ACTION",
				OracleDecision: "DRIFT_PRESENT",
				OracleDecisionDigest: sha256Digest([]byte("DRIFT_PRESENT")),
				Assignments: []OracleAssignment{
					{Assignment: AssignmentU1Baseline, InputPath: baselineRawPath, InputDigest: rawDigest},
					{Assignment: AssignmentU1Gooo, InputPath: dossierPath, InputDigest: dossierDigest},
				},
			},
			{
				TaskID: TaskU2,
				CausalState: "EXTERNAL_UTILITY_UNKNOWN",
				OracleDecision: "PROVIDE_INDEPENDENT_EXTERNAL_UTILITY_EVIDENCE",
				OracleDecisionDigest: sha256Digest([]byte("PROVIDE_INDEPENDENT_EXTERNAL_UTILITY_EVIDENCE")),
				Assignments: []OracleAssignment{
					{Assignment: AssignmentU2Baseline, InputPath: baselineEvidencePath, InputDigest: evidenceDigest},
					{Assignment: AssignmentU2Gooo, InputPath: dossierPath, InputDigest: dossierDigest},
				},
				UnknownFrontier: &UnknownFrontier{
					Stage: "EXTERNAL_UTILITY",
					Step: "REQUIRE_INDEPENDENT_EXTERNAL_UTILITY_EVIDENCE",
					Reason: "EXTERNAL_UTILITY_NOT_OBSERVED",
					UnknownClass: "CAUSALITY_UNPROVEN",
					NextOperation: "PROVIDE_INDEPENDENT_EXTERNAL_UTILITY_EVIDENCE",
					BlockedBy: []string{"exact-before-after-utility-pair"},
				},
			},
		},
	}
}

func buildAssignments() []map[string]any {
	assignments := make([]map[string]any, 0, 12)
	for _, cell := range Cells() {
		assignments = append(assignments, map[string]any{
			"ordinal": cell.Ordinal,
			"cell_id": cell.ID,
			"activity": cell.Activity,
			"proof_family": cell.ProofFamily,
			"indicator": cell.Indicator,
			"case_state": cell.CaseState,
			"task_ids": []string{TaskU1, TaskU2},
			"assignment_rule": "session_nonce_deterministic_counterbalance",
		})
	}
	return assignments
}

func writeNDJSON(path string, values []any) error {
	var builder strings.Builder
	for _, value := range values {
		data, err := json.Marshal(value)
		if err != nil {
			return err
		}
		builder.Write(data)
		builder.WriteByte('\n')
	}
	return os.WriteFile(path, []byte(builder.String()), 0o644)
}

func makeToolchain() ToolchainReference {
	value := map[string]string{"go_version": "1.27.0", "product": "gooo-utility-trial/0.1.0"}
	digest, _ := canonicalDigest(value)
	return ToolchainReference{GoVersion: value["go_version"], Product: value["product"], SHA256: digest}
}

func GeneratePacket(output string) (string, error) {
	root, err := cleanOutputPath(output)
	if err != nil {
		return "", err
	}
	for _, dir := range []string{filepath.Join(root, "baseline-pack"), filepath.Join(root, "gooo-pack")} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return "", fmt.Errorf("%w: create packet directory: %v", ErrInvalidPacket, err)
		}
	}

	fixture, err := fixtureValue()
	if err != nil {
		return "", fmt.Errorf("%w: fixture: %v", ErrInvalidPacket, err)
	}
	raw, err := baselineRawValue()
	if err != nil {
		return "", fmt.Errorf("%w: baseline raw: %v", ErrInvalidPacket, err)
	}
	evidence, err := baselineEvidenceValue()
	if err != nil {
		return "", fmt.Errorf("%w: baseline evidence: %v", ErrInvalidPacket, err)
	}
	fixtureDigest, err := canonicalDigest(fixture)
	if err != nil {
		return "", err
	}
	rawDigest, err := canonicalDigest(raw)
	if err != nil {
		return "", err
	}
	evidenceDigest, err := canonicalDigest(evidence)
	if err != nil {
		return "", err
	}
	dossierDigest := sha256Digest([]byte(dossierText))
	contract := contractValue()
	contractDigest, err := canonicalDigest(contract)
	if err != nil {
		return "", err
	}
	oracle := buildOracle(fixtureDigest, rawDigest, evidenceDigest, dossierDigest)
	oracleDigest, err := canonicalDigest(oracle)
	if err != nil {
		return "", err
	}
	toolchain := makeToolchain()

	if err := writeJSON(filepath.Join(root, fixturePath), fixture); err != nil {
		return "", err
	}
	if err := writeJSON(filepath.Join(root, baselineRawPath), raw); err != nil {
		return "", err
	}
	if err := writeJSON(filepath.Join(root, baselineEvidencePath), evidence); err != nil {
		return "", err
	}
	if err := writeJSON(filepath.Join(root, goooReleasePath), Release()); err != nil {
		return "", err
	}
	if err := os.WriteFile(filepath.Join(root, dossierPath), []byte(dossierText), 0o644); err != nil {
		return "", err
	}
	if err := writeJSON(filepath.Join(root, oraclePath), oracle); err != nil {
		return "", err
	}
	if err := writeJSON(filepath.Join(root, contractPath), contract); err != nil {
		return "", err
	}
	if err := os.WriteFile(filepath.Join(root, receiptSchemaPath), []byte(receiptSchemaJSON+"\n"), 0o644); err != nil {
		return "", err
	}
	assignmentValues := make([]any, 0, 12)
	for _, value := range buildAssignments() {
		assignmentValues = append(assignmentValues, value)
	}
	if err := writeNDJSON(filepath.Join(root, assignmentsPath), assignmentValues); err != nil {
		return "", err
	}
	if err := os.WriteFile(filepath.Join(root, validatedEvidencePath), nil, 0o644); err != nil {
		return "", err
	}
	if err := os.WriteFile(filepath.Join(root, rejectedReceiptsPath), nil, 0o644); err != nil {
		return "", err
	}
	if err := writeJSON(filepath.Join(root, runtimeMetricsPath), RuntimeMetrics{}); err != nil {
		return "", err
	}

	manifest := Manifest{
		Schema: ProtocolSchema + "/manifest/v1",
		ProtocolVersion: "v1",
		Product: "gooo-utility-trial",
		Release: Release(),
		Fixture: FixtureReference{Schema: "gooo/opentofu-envelope/plan-oracle/v1", Path: fixturePath, SHA256: fixtureDigest},
		Contract: ContractReference{Schema: ContractSchema, Path: contractPath, SHA256: contractDigest, CellCount: 12},
		Toolchain: toolchain,
		OracleDigest: oracleDigest,
		Counts: countDenominator(),
		Counterbalance: CounterbalanceRule{
			Algorithm: "sha256(session_nonce|task_id|fixture_digest|toolchain_digest|contract_digest|oracle_digest)[0] mod 2",
			Seed: "session_nonce|task_id|fixture_digest|toolchain_digest|contract_digest|oracle_digest",
			Even: OrderBaseline,
			Odd: OrderGooo,
		},
		EvidenceOriginPolicy: OriginPolicy{
			Eligible: []string{OriginExternal, OriginIndependent},
			ProtocolTestOnly: []string{OriginCI, OriginMaintainer, OriginSynthetic},
			PIICollected: false,
			SessionIdentity: "pseudonymous_session_nonce_only",
		},
		PairPolicy: PairPolicy{
			RequiredFields: []string{"session_nonce", "task_id", "fixture_digest", "toolchain_digest", "contract_digest", "oracle_digest"},
			RequiredAssignments: []string{"baseline", "gooo"},
			RequiredPhases: []string{PhaseBefore, PhaseAfter},
			NoPairState: "UTILITY_UNKNOWN",
		},
		UtilityState: "UNKNOWN",
		ProtocolReady: "CLOSED",
		ExternalEvidenceCount: 0,
		Authority: Authority{},
	}
	manifest.Artifacts, err = packetArtifacts(root, false)
	if err != nil {
		return "", err
	}
	if err := writeJSON(filepath.Join(root, "trial-manifest.json"), manifest); err != nil {
		return "", err
	}
	if err := WriteReport(root); err != nil {
		return "", err
	}
	return root, nil
}

func packetArtifacts(root string, includeReport bool) ([]Artifact, error) {
	var artifacts []Artifact
	err := filepath.Walk(root, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if info.IsDir() || filepath.Base(path) == "trial-manifest.json" || (!includeReport && filepath.Base(path) == utilityReportPath) {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		digest, bytes, err := digestFile(path)
		if err != nil {
			return err
		}
		kind := "artifact"
		if strings.HasSuffix(rel, ".ndjson") {
			kind = "evidence"
		} else if rel == runtimeMetricsPath {
			kind = "runtime_metrics"
		}
		artifacts = append(artifacts, Artifact{Path: filepath.ToSlash(rel), Kind: kind, SHA256: digest, Bytes: bytes})
		return nil
	})
	if err != nil {
		return nil, err
	}
	// The manifest is written after this inventory; include it in the stable
	// inventory only on the next report refresh.
	return artifacts, nil
}

func LoadManifest(root string) (Manifest, Oracle, error) {
	var manifest Manifest
	var oracle Oracle
	if err := readStrictJSON(filepath.Join(root, "trial-manifest.json"), &manifest); err != nil {
		return manifest, oracle, fmt.Errorf("%w: manifest: %v", ErrInvalidPacket, err)
	}
	if err := readStrictJSON(filepath.Join(root, oraclePath), &oracle); err != nil {
		return manifest, oracle, fmt.Errorf("%w: oracle: %v", ErrInvalidPacket, err)
	}
	if manifest.Schema != ProtocolSchema+"/manifest/v1" || manifest.ProtocolVersion != "v1" || manifest.Product != "gooo-utility-trial" {
		return manifest, oracle, fmt.Errorf("%w: manifest identity", ErrInvalidPacket)
	}
	if manifest.Counts.Cells != 12 || manifest.Counts.Activities != 12 || manifest.Counts.ProofFoundation != 4 || manifest.Counts.ProofCoherence != 4 || manifest.Counts.ProofRegression != 4 || manifest.Counts.IndicatorDriver != 4 || manifest.Counts.IndicatorOutcome != 4 || manifest.Counts.IndicatorGuardrail != 4 || manifest.Counts.Closed != 3 || manifest.Counts.Unknown != 3 || manifest.Counts.Refuted != 6 {
		return manifest, oracle, fmt.Errorf("%w: fixed denominator", ErrInvalidPacket)
	}
	if manifest.Release != Release() || manifest.Fixture.SHA256 == "" || manifest.Contract.SHA256 == "" || manifest.Toolchain.SHA256 == "" {
		return manifest, oracle, fmt.Errorf("%w: release or digest binding", ErrInvalidPacket)
	}
	var release ReleaseReference
	if err := readStrictJSON(filepath.Join(root, goooReleasePath), &release); err != nil || release != Release() {
		return manifest, oracle, fmt.Errorf("%w: release reference mismatch", ErrInvalidPacket)
	}
	fixtureDigest, _, err := digestFile(filepath.Join(root, fixturePath))
	if err != nil || fixtureDigest != manifest.Fixture.SHA256 {
		return manifest, oracle, fmt.Errorf("%w: fixture digest mismatch", ErrInvalidPacket)
	}
	contractDigest, _, err := digestFile(filepath.Join(root, contractPath))
	if err != nil || contractDigest != manifest.Contract.SHA256 {
		return manifest, oracle, fmt.Errorf("%w: contract digest mismatch", ErrInvalidPacket)
	}
	if oracle.FixtureDigest != manifest.Fixture.SHA256 || manifest.OracleDigest != mustOracleDigest(oracle) || len(oracle.Tasks) != 2 {
		return manifest, oracle, fmt.Errorf("%w: oracle binding", ErrInvalidPacket)
	}
	for _, item := range []struct{ path, expected string }{
		{baselineRawPath, findOracleInput(oracle, AssignmentU1Baseline)},
		{baselineEvidencePath, findOracleInput(oracle, AssignmentU2Baseline)},
		{dossierPath, findOracleInput(oracle, AssignmentU1Gooo)},
	} {
		digest, _, digestErr := digestFile(filepath.Join(root, item.path))
		if digestErr != nil || digest != item.expected {
			return manifest, oracle, fmt.Errorf("%w: input artifact digest mismatch", ErrInvalidPacket)
		}
	}
	return manifest, oracle, nil
}

func findOracleInput(oracle Oracle, assignment string) string {
	for _, task := range oracle.Tasks {
		for _, item := range task.Assignments {
			if item.Assignment == assignment {
				return item.InputDigest
			}
		}
	}
	return ""
}
