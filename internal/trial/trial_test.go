package trial

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func testReceipt(t *testing.T, root, session, receiptID, assignment string, elapsed int64, origin string, abandoned bool) Receipt {
	t.Helper()
	manifest, oracle, err := LoadManifest(root)
	if err != nil {
		t.Fatal(err)
	}
	expected, ok := assignmentExpectationFor(oracle, assignment)
	if !ok {
		t.Fatal("missing assignment expectation")
	}
	decision := sha256Digest([]byte(expected.Decision))
	state := StateCompleted
	correctness := 1
	if abandoned {
		state = StateAbandoned
		decision = emptyDecisionDigest()
		correctness = 0
	}
	return Receipt{
		Schema: ProtocolSchema + "/session-receipt/v1",
		ReceiptID: receiptID,
		SessionNonce: session,
		TaskID: expected.TaskID,
		Assignment: assignment,
		FixtureDigest: manifest.Fixture.SHA256,
		ToolchainDigest: manifest.Toolchain.SHA256,
		ContractDigest: manifest.Contract.SHA256,
		OracleDigest: mustOracleDigest(oracle),
		InputDigest: expected.InputDigest,
		DecisionDigest: decision,
		Correctness: correctness,
		ElapsedMS: elapsed,
		TimingKind: TimingObserved,
		Estimated: false,
		AbandonmentState: state,
		EnvironmentDigest: sha256Digest([]byte("test-environment")),
		Origin: origin,
		ConsentRetentionPolicy: ConsentRetentionPolicy{ConsentGiven: true, RetentionPolicy: "test-only; delete after test", RetentionDays: 1},
		UnknownFrontier: oracleFrontier(oracle),
		BeforeAfterPair: PairMetadata{PairID: pairID(Receipt{SessionNonce: session, TaskID: expected.TaskID, OracleDigest: mustOracleDigest(oracle)}, manifest), CounterbalanceOrder: "", Phase: "", PairEligible: false},
	}
}

func finishReceipt(t *testing.T, root string, receipt Receipt) Receipt {
	t.Helper()
	manifest, _, err := LoadManifest(root)
	if err != nil {
		t.Fatal(err)
	}
	receipt.BeforeAfterPair.CounterbalanceOrder = counterbalance(receipt, manifest)
	receipt.BeforeAfterPair.Phase = expectedPhase(receipt.Assignment, receipt.BeforeAfterPair.CounterbalanceOrder)
	receipt.BeforeAfterPair.PairID = pairID(receipt, manifest)
	return receipt
}

func writeReceipt(t *testing.T, dir string, receipt Receipt) string {
	t.Helper()
	path := filepath.Join(dir, receipt.ReceiptID+".json")
	if err := writeJSON(path, receipt); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestGeneratePacketHasFixedBoundary(t *testing.T) {
	root := filepath.Join(t.TempDir(), "packet")
	if _, err := GeneratePacket(root); err != nil {
		t.Fatal(err)
	}
	manifest, oracle, err := LoadManifest(root)
	if err != nil {
		t.Fatal(err)
	}
	if manifest.ProtocolReady != "CLOSED" || manifest.UtilityState != "UNKNOWN" || manifest.ProcessState != "REFUTED" || manifest.ExternalEvidenceCount != 0 {
		t.Fatalf("unexpected manifest state: %+v", manifest)
	}
	if !reflect.DeepEqual(manifest.Counts, countDenominator()) || len(oracle.Tasks) != 2 {
		t.Fatalf("unexpected fixed denominator: %+v", manifest.Counts)
	}
	data, err := os.ReadFile(filepath.Join(root, utilityReportPath))
	if err != nil {
		t.Fatal(err)
	}
	report := string(data)
	for _, expected := range []string{"protocol_ready: `CLOSED`", "utility: `UNKNOWN`", "process: `REFUTED`", "eligible sessions | 0", "eligible pairs | 0", "REFUTED 6", "Go physical lines | 0"} {
		if !strings.Contains(report, expected) {
			t.Errorf("report missing %q", expected)
		}
	}
	if _, err := os.Stat(filepath.Join(root, validatedEvidencePath)); err != nil {
		t.Fatal(err)
	}
}

func TestRecordMatchedPairAndExcludeProtocolTest(t *testing.T) {
	root := filepath.Join(t.TempDir(), "packet")
	if _, err := GeneratePacket(root); err != nil {
		t.Fatal(err)
	}
	receiptDir := filepath.Join(t.TempDir(), "receipts")
	if err := os.MkdirAll(receiptDir, 0o755); err != nil {
		t.Fatal(err)
	}
	base := finishReceipt(t, root, testReceipt(t, root, "11111111111111111111111111111111", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", AssignmentU1Baseline, 120, OriginExternal, false))
	gooo := finishReceipt(t, root, testReceipt(t, root, "11111111111111111111111111111111", "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb", AssignmentU1Gooo, 80, OriginExternal, false))
	basePath := writeReceipt(t, receiptDir, base)
	goooPath := writeReceipt(t, receiptDir, gooo)
	first, err := RecordReceipt(root, basePath)
	if err != nil || first.PairEligible {
		t.Fatalf("first receipt: %+v %v", first, err)
	}
	second, err := RecordReceipt(root, goooPath)
	if err != nil || !second.PairEligible {
		t.Fatalf("second receipt: %+v %v", second, err)
	}
	if _, err := RecordReceipt(root, goooPath); err == nil || !strings.Contains(err.Error(), "REPLAYED_RECEIPT") {
		t.Fatalf("replay was not rejected: %v", err)
	}
	protocol := finishReceipt(t, root, testReceipt(t, root, "22222222222222222222222222222222", "cccccccccccccccccccccccccccccccc", AssignmentU2Baseline, 40, OriginCI, false))
	protocolPath := writeReceipt(t, receiptDir, protocol)
	if _, err := RecordReceipt(root, protocolPath); err != nil {
		t.Fatal(err)
	}
	reportBytes, err := os.ReadFile(filepath.Join(root, utilityReportPath))
	if err != nil {
		t.Fatal(err)
	}
	report := string(reportBytes)
	for _, expected := range []string{"eligible sessions | 1", "eligible pairs | 1", "correct before | 1", "correct after | 1", "external_evidence: `2`", "protocol_test_evidence: `1`", "utility_improvement: `CLOSED`"} {
		if !strings.Contains(report, expected) {
			t.Errorf("report missing %q", expected)
		}
	}
}

func TestInvalidReceiptReasonsFailClosed(t *testing.T) {
	root := filepath.Join(t.TempDir(), "packet")
	if _, err := GeneratePacket(root); err != nil {
		t.Fatal(err)
	}
	receipt := finishReceipt(t, root, testReceipt(t, root, "33333333333333333333333333333333", "dddddddddddddddddddddddddddddddd", AssignmentU2Gooo, 5, OriginExternal, false))
	for name, mutate := range map[string]func(*Receipt){
		"origin": func(value *Receipt) { value.Origin = "invented" },
		"negative timing": func(value *Receipt) { value.ElapsedMS = -1 },
		"estimated timing": func(value *Receipt) { value.Estimated = true },
		"oracle": func(value *Receipt) { value.OracleDigest = sha256Digest([]byte("wrong")) },
		"participant mismatch": func(value *Receipt) { value.SessionNonce = "not-a-pseudonym" },
	} {
		copy := receipt
		mutate(&copy)
		if _, err := ValidateReceipt(root, copy); err == nil {
			t.Errorf("%s was accepted", name)
		}
	}
}

func TestReceiptJSONHasNoNaturalLanguageDecision(t *testing.T) {
	root := filepath.Join(t.TempDir(), "packet")
	if _, err := GeneratePacket(root); err != nil {
		t.Fatal(err)
	}
	receipt := finishReceipt(t, root, testReceipt(t, root, "44444444444444444444444444444444", "eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee", AssignmentU1Baseline, 1, OriginExternal, false))
	data, err := json.Marshal(receipt)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), "DRIFT_PRESENT") || strings.Contains(string(data), "looks good") {
		t.Fatal("receipt contains a raw natural-language or decision token")
	}
}
