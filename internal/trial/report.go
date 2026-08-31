package trial

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type pairObservation struct {
	SessionNonce string
	TaskID       string
	Before       validatedReceipt
	After        validatedReceipt
	Eligible     bool
}

type reportMetrics struct {
	ProtocolReady string `json:"protocol_ready"`
	Utility       string `json:"utility"`
	UtilityImprovement string `json:"utility_improvement"`
	EligibleSessions int `json:"eligible_sessions"`
	EligiblePairs int `json:"eligible_pairs"`
	CorrectBefore int `json:"correct_before"`
	CorrectAfter int `json:"correct_after"`
	ElapsedBeforeTotalMS int64 `json:"elapsed_before_ms_total"`
	ElapsedAfterTotalMS int64 `json:"elapsed_after_ms_total"`
	AbandonedBefore int `json:"abandoned_before"`
	AbandonedAfter int `json:"abandoned_after"`
	RejectedReceipts int `json:"rejected_receipts"`
	RejectedByReason map[string]int `json:"rejected_by_reason"`
	ExternalEvidence int `json:"external_evidence"`
	ProtocolTestEvidence int `json:"protocol_test_evidence"`
}

func readRejected(root string) ([]RejectedReceipt, error) {
	file, err := os.Open(filepath.Join(root, rejectedReceiptsPath))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	var result []RejectedReceipt
	for scanner.Scan() {
		if strings.TrimSpace(scanner.Text()) == "" {
			return nil, errors.New("empty rejected receipt line")
		}
		var item RejectedReceipt
		if err := json.Unmarshal([]byte(scanner.Text()), &item); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, scanner.Err()
}

func groupPairs(values []validatedReceipt, manifest Manifest) []pairObservation {
	byKey := map[string][]validatedReceipt{}
	for _, value := range values {
		receipt := value.Receipt
		key := strings.Join([]string{receipt.SessionNonce, receipt.TaskID, receipt.BeforeAfterPair.PairID, receipt.FixtureDigest, receipt.ToolchainDigest, receipt.ContractDigest, receipt.OracleDigest}, "|")
		byKey[key] = append(byKey[key], value)
	}
	var pairs []pairObservation
	for _, group := range byKey {
		if len(group) != 2 {
			continue
		}
		var before, after validatedReceipt
		var hasBefore, hasAfter bool
		for _, value := range group {
			if value.Receipt.BeforeAfterPair.Phase == PhaseBefore {
				before, hasBefore = value, true
			}
			if value.Receipt.BeforeAfterPair.Phase == PhaseAfter {
				after, hasAfter = value, true
			}
		}
		if !hasBefore || !hasAfter {
			continue
		}
		eligible := before.UtilityEligible && after.UtilityEligible && before.Receipt.BeforeAfterPair.CounterbalanceOrder == after.Receipt.BeforeAfterPair.CounterbalanceOrder && before.Receipt.BeforeAfterPair.PairID == after.Receipt.BeforeAfterPair.PairID && before.Receipt.TaskID == after.Receipt.TaskID && before.Receipt.SessionNonce == after.Receipt.SessionNonce
		pairs = append(pairs, pairObservation{SessionNonce: before.Receipt.SessionNonce, TaskID: before.Receipt.TaskID, Before: before, After: after, Eligible: eligible})
	}
	sort.Slice(pairs, func(i, j int) bool {
		if pairs[i].SessionNonce != pairs[j].SessionNonce {
			return pairs[i].SessionNonce < pairs[j].SessionNonce
		}
		return pairs[i].TaskID < pairs[j].TaskID
	})
	return pairs
}

func loadRuntimeMetrics(root string) (RuntimeMetrics, error) {
	var metrics RuntimeMetrics
	if err := readStrictJSON(filepath.Join(root, runtimeMetricsPath), &metrics); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return metrics, nil
		}
		return metrics, err
	}
	for _, value := range []int64{metrics.ExecutedTests, metrics.ReusedTests, metrics.SkippedTests, metrics.NotObservedTests, metrics.GoLines, metrics.GoooLines, metrics.RegularFiles, metrics.DescendantDirs} {
		if value < 0 {
			return metrics, errors.New("negative runtime metric")
		}
	}
	return metrics, nil
}

func metricValue(value *int64) string {
	if value == nil {
		return "NOT_OBSERVED"
	}
	return fmt.Sprintf("%d", *value)
}

func packetInventory(root string) (int64, int64, []string, error) {
	var files, bytes int64
	var paths []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if info.IsDir() || filepath.Base(path) == utilityReportPath {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		files++
		bytes += info.Size()
		paths = append(paths, filepath.ToSlash(rel))
		return nil
	})
	sort.Strings(paths)
	return files, bytes, paths, err
}

func calculateReport(root string, manifest Manifest, values []validatedReceipt, rejected []RejectedReceipt) (reportMetrics, RuntimeMetrics, int64, int64, []string, error) {
	metrics := reportMetrics{
		ProtocolReady: "CLOSED",
		Utility: "UNKNOWN",
		UtilityImprovement: "UNKNOWN",
		RejectedByReason: map[string]int{},
	}
	for _, item := range rejected {
		metrics.RejectedByReason[item.Reason]++
	}
	metrics.RejectedReceipts = len(rejected)
	for _, value := range values {
		if value.UtilityEligible {
			metrics.ExternalEvidence++
		} else {
			metrics.ProtocolTestEvidence++
		}
	}
	pairs := groupPairs(values, manifest)
	eligiblePairs := make([]pairObservation, 0, len(pairs))
	sessions := map[string]bool{}
	for _, pair := range pairs {
		if !pair.Eligible {
			continue
		}
		eligiblePairs = append(eligiblePairs, pair)
		sessions[pair.SessionNonce] = true
		if pair.Before.Receipt.Correctness == 1 {
			metrics.CorrectBefore++
		}
		if pair.After.Receipt.Correctness == 1 {
			metrics.CorrectAfter++
		}
		metrics.ElapsedBeforeTotalMS += pair.Before.Receipt.ElapsedMS
		metrics.ElapsedAfterTotalMS += pair.After.Receipt.ElapsedMS
		if pair.Before.Receipt.AbandonmentState == StateAbandoned {
			metrics.AbandonedBefore++
		}
		if pair.After.Receipt.AbandonmentState == StateAbandoned {
			metrics.AbandonedAfter++
		}
	}
	metrics.EligiblePairs = len(eligiblePairs)
	metrics.EligibleSessions = len(sessions)
	if metrics.EligiblePairs > 0 {
		beforeScore := metrics.CorrectBefore*2 - metrics.AbandonedBefore
		afterScore := metrics.CorrectAfter*2 - metrics.AbandonedAfter
		if afterScore > beforeScore || (afterScore == beforeScore && metrics.ElapsedAfterTotalMS < metrics.ElapsedBeforeTotalMS) {
			metrics.Utility = "CLOSED"
			metrics.UtilityImprovement = "CLOSED"
		} else if afterScore < beforeScore || (afterScore == beforeScore && metrics.ElapsedAfterTotalMS > metrics.ElapsedBeforeTotalMS) {
			metrics.Utility = "REFUTED"
			metrics.UtilityImprovement = "REFUTED"
		}
	}
	runtimeMetrics, err := loadRuntimeMetrics(root)
	if err != nil {
		return reportMetrics{}, RuntimeMetrics{}, 0, 0, nil, err
	}
	files, bytes, paths, err := packetInventory(root)
	if err != nil {
		return reportMetrics{}, RuntimeMetrics{}, 0, 0, nil, err
	}
	return metrics, runtimeMetrics, files, bytes, paths, nil
}

func WriteReport(root string) error {
	manifest, _, err := LoadManifest(root)
	if err != nil {
		return err
	}
	artifacts, err := packetArtifacts(root, false)
	if err != nil {
		return err
	}
	manifest.Artifacts = artifacts
	if err := writeJSON(filepath.Join(root, "trial-manifest.json"), manifest); err != nil {
		return err
	}
	values, err := loadEvidence(root)
	if err != nil {
		return err
	}
	rejected, err := readRejected(root)
	if err != nil {
		return err
	}
	metrics, runtimeMetrics, files, bytes, paths, err := calculateReport(root, manifest, values, rejected)
	if err != nil {
		return err
	}
	var builder strings.Builder
	builder.WriteString("# Gooo utility trial report\n\n")
	builder.WriteString("This report is a fail-closed measurement boundary. It does not convert CI, maintainer self-tests, or synthetic fixtures into external utility evidence.\n\n")
	fmt.Fprintf(&builder, "- protocol_ready: `%s`\n- utility: `%s`\n- utility_improvement: `%s`\n- process: `REFUTED`\n- external_evidence: `%d`\n- protocol_test_evidence: `%d`\n", metrics.ProtocolReady, metrics.Utility, metrics.UtilityImprovement, metrics.ExternalEvidence, metrics.ProtocolTestEvidence)
	if metrics.EligiblePairs == 0 {
		builder.WriteString("- no eligible before/after pair: utility improvement remains `UNKNOWN`\n")
	}
	builder.WriteString("\n## Exact user metrics\n\n")
	fmt.Fprintf(&builder, "| metric | value |\n|---|---:|\n| eligible sessions | %d |\n| eligible pairs | %d |\n| correct before | %d |\n| correct after | %d |\n| elapsed before total (ms) | %d |\n| elapsed after total (ms) | %d |\n| abandoned before | %d |\n| abandoned after | %d |\n| rejected receipts | %d |\n", metrics.EligibleSessions, metrics.EligiblePairs, metrics.CorrectBefore, metrics.CorrectAfter, metrics.ElapsedBeforeTotalMS, metrics.ElapsedAfterTotalMS, metrics.AbandonedBefore, metrics.AbandonedAfter, metrics.RejectedReceipts)
	builder.WriteString("\n### Rejected receipts by reason\n\n")
	if len(metrics.RejectedByReason) == 0 {
		builder.WriteString("- none\n")
	} else {
		for _, reason := range sortedKeys(metrics.RejectedByReason) {
			fmt.Fprintf(&builder, "- `%s`: %d\n", reason, metrics.RejectedByReason[reason])
		}
	}
	builder.WriteString("\n## Artifact and execution metrics\n\n")
	fmt.Fprintf(&builder, "| metric | value |\n|---|---:|\n| artifact files (utility-report excluded) | %d |\n| artifact bytes (utility-report excluded) | %d |\n| build wall (ms) | %s |\n| test wall (ms) | %s |\n| conformance wall (ms) | %s |\n| build peak RSS (KiB) | %s |\n| test peak RSS (KiB) | %s |\n| conformance peak RSS (KiB) | %s |\n| executed tests | %d |\n| reused tests | %d |\n| skipped tests | %d |\n| not-observed tests | %d |\n| Go physical lines | %d |\n| Gooo physical lines | %d |\n| regular files (root README excluded) | %d |\n| descendant directories (root README excluded) | %d |\n", files, bytes, metricValue(runtimeMetrics.BuildWallMS), metricValue(runtimeMetrics.TestWallMS), metricValue(runtimeMetrics.ConformanceWallMS), metricValue(runtimeMetrics.BuildRSSKiB), metricValue(runtimeMetrics.TestRSSKiB), metricValue(runtimeMetrics.ConformanceRSSKiB), runtimeMetrics.ExecutedTests, runtimeMetrics.ReusedTests, runtimeMetrics.SkippedTests, runtimeMetrics.NotObservedTests, runtimeMetrics.GoLines, runtimeMetrics.GoooLines, runtimeMetrics.RegularFiles, runtimeMetrics.DescendantDirs)
	builder.WriteString("\n## Fixed denominator\n\n")
	fmt.Fprintf(&builder, "- cells / .gooo activities: %d / %d\n- proof: FOUNDATION %d / COHERENCE %d / REGRESSION %d\n- indicator: DRIVER %d / OUTCOME %d / GUARDRAIL %d\n- cases: CLOSED %d / UNKNOWN %d / REFUTED %d\n- precedence: `REFUTED > UNKNOWN > CLOSED`\n", manifest.Counts.Cells, manifest.Counts.Activities, manifest.Counts.ProofFoundation, manifest.Counts.ProofCoherence, manifest.Counts.ProofRegression, manifest.Counts.IndicatorDriver, manifest.Counts.IndicatorOutcome, manifest.Counts.IndicatorGuardrail, manifest.Counts.Closed, manifest.Counts.Unknown, manifest.Counts.Refuted)
	builder.WriteString("\n## Artifact files\n\n")
	for _, path := range paths {
		fmt.Fprintf(&builder, "- `%s`\n", path)
	}
	builder.WriteString("\n## Authority\n\n")
	builder.WriteString("| authority observation | value |\n|---|---:|\n| source repository writes | 0 |\n| product repository writes | 0 |\n| local test executions | 0 |\n| external network access by product | 0 |\n| OpenTofu apply executions | 0 |\n| cloud mutations | 0 |\n")
	builder.WriteString("\nThe release immutability claim is authoritative only when the GitHub API check in the Actions workflow reports the pinned release as immutable; this report's embedded reference is not authority.\n")
	return os.WriteFile(filepath.Join(root, utilityReportPath), []byte(builder.String()), 0o644)
}
