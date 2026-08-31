package trial

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type assignmentExpectation struct {
	TaskID     string
	Assignment string
	InputPath  string
	InputDigest string
	Decision   string
}

type validatedReceipt struct {
	Receipt       Receipt
	ReceiptDigest string
	UtilityEligible bool
	PairEligible bool
}

type validationResult struct {
	ReceiptDigest string `json:"receipt_digest"`
	UtilityEligible bool `json:"utility_eligible"`
	PairEligible bool `json:"pair_eligible"`
	Reason string `json:"reason"`
}

func emptyDecisionDigest() string {
	return sha256Digest(nil)
}

func assignmentExpectationFor(oracle Oracle, assignment string) (assignmentExpectation, bool) {
	for _, task := range oracle.Tasks {
		for _, item := range task.Assignments {
			if item.Assignment == assignment {
				return assignmentExpectation{TaskID: task.TaskID, Assignment: assignment, InputPath: item.InputPath, InputDigest: item.InputDigest, Decision: task.OracleDecision}, true
			}
		}
	}
	return assignmentExpectation{}, false
}

func taskForAssignment(assignment string) string {
	switch assignment {
	case AssignmentU1Baseline, AssignmentU1Gooo:
		return TaskU1
	case AssignmentU2Baseline, AssignmentU2Gooo:
		return TaskU2
	default:
		return ""
	}
}

func isBaseline(assignment string) bool {
	return assignment == AssignmentU1Baseline || assignment == AssignmentU2Baseline
}

func isGooo(assignment string) bool {
	return assignment == AssignmentU1Gooo || assignment == AssignmentU2Gooo
}

func eligibleOrigin(origin string) bool {
	return origin == OriginExternal || origin == OriginIndependent
}

func allowedOrigin(origin string) bool {
	switch origin {
	case OriginExternal, OriginIndependent, OriginCI, OriginMaintainer, OriginSynthetic:
		return true
	default:
		return false
	}
}

func sessionSeed(receipt Receipt, manifest Manifest) string {
	return strings.Join([]string{receipt.SessionNonce, receipt.TaskID, manifest.Fixture.SHA256, manifest.Toolchain.SHA256, manifest.Contract.SHA256, receipt.OracleDigest}, "|")
}

func counterbalance(receipt Receipt, manifest Manifest) string {
	digest := sha256Digest([]byte(sessionSeed(receipt, manifest)))
	first, _ := hex.DecodeString(digest[len("sha256:") : len("sha256:")+2])
	if len(first) == 1 && first[0]%2 == 0 {
		return OrderBaseline
	}
	return OrderGooo
}

func pairID(receipt Receipt, manifest Manifest) string {
	return sha256Digest([]byte(sessionSeed(receipt, manifest)))
}

func expectedPhase(assignment, order string) string {
	if order == OrderBaseline {
		if isBaseline(assignment) {
			return PhaseBefore
		}
		return PhaseAfter
	}
	if isGooo(assignment) {
		return PhaseBefore
	}
	return PhaseAfter
}

func validateReceiptShape(receipt Receipt) error {
	if receipt.Schema != ProtocolSchema+"/session-receipt/v1" {
		return errors.New("SCHEMA_MISMATCH")
	}
	if !validNonce(receipt.ReceiptID) {
		return errors.New("RECEIPT_ID_INVALID")
	}
	if !validNonce(receipt.SessionNonce) {
		return errors.New("SESSION_NONCE_INVALID")
	}
	if taskForAssignment(receipt.Assignment) == "" {
		return errors.New("ASSIGNMENT_INVALID")
	}
	if receipt.TaskID != taskForAssignment(receipt.Assignment) {
		return errors.New("TASK_ASSIGNMENT_MISMATCH")
	}
	for _, digest := range []string{receipt.FixtureDigest, receipt.ToolchainDigest, receipt.ContractDigest, receipt.OracleDigest, receipt.InputDigest, receipt.DecisionDigest, receipt.EnvironmentDigest, receipt.BeforeAfterPair.PairID} {
		if !validDigest(digest) {
			return errors.New("DIGEST_INVALID")
		}
	}
	if receipt.Correctness != 0 && receipt.Correctness != 1 {
		return errors.New("CORRECTNESS_NOT_INTEGER_FLAG")
	}
	if receipt.ElapsedMS < 0 {
		return errors.New("NEGATIVE_ELAPSED_MS")
	}
	if receipt.TimingKind != TimingObserved || receipt.Estimated {
		return errors.New("ESTIMATED_OR_UNOBSERVED_TIMING")
	}
	if receipt.AbandonmentState != StateCompleted && receipt.AbandonmentState != StateAbandoned {
		return errors.New("ABANDONMENT_STATE_INVALID")
	}
	if !allowedOrigin(receipt.Origin) {
		return errors.New("ORIGIN_INVALID")
	}
	if strings.TrimSpace(receipt.ConsentRetentionPolicy.RetentionPolicy) == "" || receipt.ConsentRetentionPolicy.RetentionDays < 1 {
		return errors.New("CONSENT_RETENTION_POLICY_INVALID")
	}
	if eligibleOrigin(receipt.Origin) && !receipt.ConsentRetentionPolicy.ConsentGiven {
		return errors.New("ELIGIBLE_ORIGIN_WITHOUT_CONSENT")
	}
	if receipt.BeforeAfterPair.Phase != PhaseBefore && receipt.BeforeAfterPair.Phase != PhaseAfter {
		return errors.New("PAIR_PHASE_INVALID")
	}
	if receipt.BeforeAfterPair.CounterbalanceOrder != OrderBaseline && receipt.BeforeAfterPair.CounterbalanceOrder != OrderGooo {
		return errors.New("COUNTERBALANCE_INVALID")
	}
	return nil
}

func validNonce(value string) bool {
	if len(value) != 32 {
		return false
	}
	for _, char := range value {
		if !((char >= '0' && char <= '9') || (char >= 'a' && char <= 'f')) {
			return false
		}
	}
	return true
}

func receiptDigest(receipt Receipt) (string, error) {
	data, err := canonicalJSON(receipt)
	if err != nil {
		return "", err
	}
	return sha256Digest(data), nil
}

func loadReceipt(path string) (Receipt, []byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Receipt{}, nil, err
	}
	var receipt Receipt
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&receipt); err != nil {
		return Receipt{}, data, fmt.Errorf("%w: %v", ErrInvalidReceipt, err)
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return Receipt{}, data, fmt.Errorf("%w: trailing JSON", ErrInvalidReceipt)
	}
	return receipt, data, nil
}

func loadEvidence(root string) ([]validatedReceipt, error) {
	path := filepath.Join(root, validatedEvidencePath)
	file, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	manifest, oracle, err := LoadManifest(root)
	if err != nil {
		return nil, err
	}
	var result []validatedReceipt
	seenIDs := map[string]bool{}
	seenDigests := map[string]bool{}
	line := 0
	for scanner.Scan() {
		line++
		if strings.TrimSpace(scanner.Text()) == "" {
			return nil, fmt.Errorf("%w: empty evidence line %d", ErrInvalidReceipt, line)
		}
		var item EvidenceLine
		decoder := json.NewDecoder(strings.NewReader(scanner.Text()))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&item); err != nil {
			return nil, fmt.Errorf("%w: evidence line %d: %v", ErrInvalidReceipt, line, err)
		}
		if item.ReceiptID() == "" {
			return nil, fmt.Errorf("%w: evidence line %d receipt id", ErrInvalidReceipt, line)
		}
		if seenIDs[item.ReceiptID()] || seenDigests[item.ReceiptDigest] {
			return nil, fmt.Errorf("%w: REPLAYED_RECEIPT", ErrInvalidReceipt)
		}
		if err := validateAgainstPacket(item.Receipt, manifest, oracle); err != nil {
			return nil, fmt.Errorf("%w: evidence line %d: %v", ErrInvalidReceipt, line, err)
		}
		digest, err := receiptDigest(item.Receipt)
		if err != nil || digest != item.ReceiptDigest {
			return nil, fmt.Errorf("%w: evidence line %d digest", ErrInvalidReceipt, line)
		}
		seenIDs[item.ReceiptID()] = true
		seenDigests[item.ReceiptDigest] = true
		result = append(result, validatedReceipt{Receipt: item.Receipt, ReceiptDigest: item.ReceiptDigest, UtilityEligible: item.UtilityEligible, PairEligible: item.PairEligible})
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	for _, item := range result {
		if item.UtilityEligible != eligibleOrigin(item.Receipt.Origin) {
			return nil, fmt.Errorf("%w: UTILITY_ELIGIBILITY_MISMATCH", ErrInvalidReceipt)
		}
		if item.PairEligible != pairExists(item.Receipt, result, manifest) {
			return nil, fmt.Errorf("%w: PAIR_ELIGIBILITY_MISMATCH", ErrInvalidReceipt)
		}
	}
	return result, nil
}

// ReceiptID is kept as a method so the evidence reader cannot confuse the
// nested receipt identity with the digest of the enclosing evidence line.
func (e EvidenceLine) ReceiptID() string {
	return e.Receipt.ReceiptID
}

func validateAgainstPacket(receipt Receipt, manifest Manifest, oracle Oracle) error {
	if err := validateReceiptShape(receipt); err != nil {
		return err
	}
	if receipt.FixtureDigest != manifest.Fixture.SHA256 || receipt.ToolchainDigest != manifest.Toolchain.SHA256 || receipt.ContractDigest != manifest.Contract.SHA256 {
		return errors.New("INPUT_CONTEXT_DIGEST_MISMATCH")
	}
	if receipt.OracleDigest != mustOracleDigest(oracle) {
		return errors.New("ORACLE_DIGEST_MISMATCH")
	}
	expected, ok := assignmentExpectationFor(oracle, receipt.Assignment)
	if !ok || expected.TaskID != receipt.TaskID || expected.InputDigest != receipt.InputDigest {
		return errors.New("ASSIGNMENT_INPUT_DIGEST_MISMATCH")
	}
	order := counterbalance(receipt, manifest)
	if receipt.BeforeAfterPair.CounterbalanceOrder != order || receipt.BeforeAfterPair.Phase != expectedPhase(receipt.Assignment, order) {
		return errors.New("COUNTERBALANCE_PHASE_MISMATCH")
	}
	if receipt.BeforeAfterPair.PairID != pairID(receipt, manifest) {
		return errors.New("PAIR_ID_MISMATCH")
	}
	if receipt.AbandonmentState == StateAbandoned {
		if receipt.Correctness != 0 || receipt.DecisionDigest != emptyDecisionDigest() {
			return errors.New("ABANDONED_RECEIPT_HAS_DECISION")
		}
	} else {
		correct := receipt.DecisionDigest == sha256Digest([]byte(expected.Decision))
		if receipt.Correctness != boolInt(correct) {
			return errors.New("ORACLE_CORRECTNESS_MISMATCH")
		}
	}
	return nil
}

func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

func mustOracleDigest(oracle Oracle) string {
	digest, _ := canonicalDigest(oracle)
	return digest
}

func validateReceiptForPacket(root string, receipt Receipt) (validationResult, error) {
	manifest, oracle, err := LoadManifest(root)
	if err != nil {
		return validationResult{}, err
	}
	if err := validateAgainstPacket(receipt, manifest, oracle); err != nil {
		return validationResult{}, fmt.Errorf("%w: %v", ErrInvalidReceipt, err)
	}
	digest, err := receiptDigest(receipt)
	if err != nil {
		return validationResult{}, err
	}
	existing, err := loadEvidence(root)
	if err != nil {
		return validationResult{}, err
	}
	for _, item := range existing {
		if item.Receipt.ReceiptID == receipt.ReceiptID || item.ReceiptDigest == digest {
			return validationResult{}, fmt.Errorf("%w: REPLAYED_RECEIPT", ErrInvalidReceipt)
		}
	}
	pairEligible := pairExists(receipt, existing, manifest)
	if receipt.BeforeAfterPair.PairEligible && !pairEligible {
		return validationResult{}, fmt.Errorf("%w: PAIR_ELIGIBILITY_MISMATCH", ErrInvalidReceipt)
	}
	return validationResult{ReceiptDigest: digest, UtilityEligible: eligibleOrigin(receipt.Origin), PairEligible: pairEligible, Reason: "RECEIPT_ACCEPTED"}, nil
}

// ValidateReceipt checks a receipt without writing evidence or rejection
// records. It is the safe inspection path for participants and reviewers.
func ValidateReceipt(root string, receipt Receipt) (validationResult, error) {
	return validateReceiptForPacket(root, receipt)
}

func pairExists(receipt Receipt, existing []validatedReceipt, manifest Manifest) bool {
	for _, item := range existing {
		other := item.Receipt
		if other.SessionNonce == receipt.SessionNonce && other.TaskID == receipt.TaskID && other.FixtureDigest == receipt.FixtureDigest && other.ToolchainDigest == receipt.ToolchainDigest && other.ContractDigest == receipt.ContractDigest && other.OracleDigest == receipt.OracleDigest && other.BeforeAfterPair.PairID == receipt.BeforeAfterPair.PairID && other.BeforeAfterPair.Phase != receipt.BeforeAfterPair.Phase && ((isBaseline(other.Assignment) && isGooo(receipt.Assignment)) || (isGooo(other.Assignment) && isBaseline(receipt.Assignment))) {
			return true
		}
	}
	return false
}

func appendRejected(root string, data []byte, receiptID, reason string) error {
	item := RejectedReceipt{ObservedAtDigest: sha256Digest(data), ReceiptID: receiptID, Reason: reason}
	encoded, err := json.Marshal(item)
	if err != nil {
		return err
	}
	file, err := os.OpenFile(filepath.Join(root, rejectedReceiptsPath), os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.Write(append(encoded, '\n'))
	return err
}

func writeEvidence(root string, values []validatedReceipt) error {
	sort.Slice(values, func(i, j int) bool {
		if values[i].Receipt.SessionNonce != values[j].Receipt.SessionNonce {
			return values[i].Receipt.SessionNonce < values[j].Receipt.SessionNonce
		}
		if values[i].Receipt.TaskID != values[j].Receipt.TaskID {
			return values[i].Receipt.TaskID < values[j].Receipt.TaskID
		}
		return values[i].Receipt.BeforeAfterPair.Phase < values[j].Receipt.BeforeAfterPair.Phase
	})
	file, err := os.Create(filepath.Join(root, validatedEvidencePath))
	if err != nil {
		return err
	}
	defer file.Close()
	manifest := mustManifest(root)
	for index, value := range values {
		pairEligible := pairExists(value.Receipt, values, manifest)
		values[index].PairEligible = pairEligible
		encoded, err := json.Marshal(EvidenceLine{Receipt: value.Receipt, ReceiptDigest: value.ReceiptDigest, UtilityEligible: value.UtilityEligible, PairEligible: pairEligible})
		if err != nil {
			return err
		}
		if _, err := file.Write(append(encoded, '\n')); err != nil {
			return err
		}
	}
	return nil
}

func RecordReceipt(root, receiptPath string) (validationResult, error) {
	receipt, raw, err := loadReceipt(receiptPath)
	if err != nil {
		_ = appendRejected(root, raw, "", "RECEIPT_JSON_INVALID")
		return validationResult{}, err
	}
	result, err := validateReceiptForPacket(root, receipt)
	if err != nil {
		_ = appendRejected(root, raw, receipt.ReceiptID, reasonOf(err))
		return validationResult{}, err
	}
	values, err := loadEvidence(root)
	if err != nil {
		return validationResult{}, err
	}
	values = append(values, validatedReceipt{Receipt: receipt, ReceiptDigest: result.ReceiptDigest, UtilityEligible: result.UtilityEligible, PairEligible: result.PairEligible})
	if err := writeEvidence(root, values); err != nil {
		return validationResult{}, err
	}
	if err := WriteReport(root); err != nil {
		return validationResult{}, err
	}
	result.PairEligible = pairExists(receipt, values, mustManifest(root))
	return result, nil
}

func mustManifest(root string) Manifest {
	manifest, _, _ := LoadManifest(root)
	return manifest
}

func reasonOf(err error) string {
	text := err.Error()
	if index := strings.Index(text, ": "); index >= 0 {
		text = text[index+2:]
	}
	if index := strings.Index(text, ": "); index >= 0 {
		text = text[:index]
	}
	return text
}

func ReadReceipt(path string) (Receipt, error) {
	receipt, _, err := loadReceipt(path)
	return receipt, err
}
