package trial

import (
	"crypto/sha256"
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

const (
	ProtocolSchema = "gooo/utility-trial/v1"
	ContractSchema = "gooo/utility-trial/denominator/v1"

	UpstreamRepository = "kimjooyoon/gooo-opentofu-envelope"
	UpstreamTag        = "v0.1.1"
	UpstreamReleaseID  = int64(379769579)
	UpstreamTagObject  = "1d8e07c49b700b7e72ec9c413157d851188c09b6"
	UpstreamCommit     = "6b482d402f4ceff8a8b23205cce8db5154305382"
	UpstreamAssetID    = int64(538012631)
	UpstreamAssetName  = "gooo-opentofu-envelope-v0.1.0-evidence.zip"
	UpstreamAssetSize  = int64(17829)
	UpstreamAssetSHA   = "sha256:dc8a466ecc4ea2f5c127dcc59985252ba8a01a7ae6eb8fb1caa63f7f595fcd31"

	TaskU1 = "U1_OPENTOFU_DRIFT"
	TaskU2 = "U2_UNKNOWN_RESUME"

	AssignmentU1Baseline = "U1_BASELINE_RAW"
	AssignmentU1Gooo     = "U1_GOOO_DOSSIER"
	AssignmentU2Baseline = "U2_BASELINE_EVIDENCE_SET"
	AssignmentU2Gooo     = "U2_GOOO_FRONTIER"

	OriginExternal    = "EXTERNAL_USER"
	OriginIndependent = "INDEPENDENT_SESSION"
	OriginCI          = "CI"
	OriginMaintainer  = "MAINTAINER_SELF_TEST"
	OriginSynthetic   = "SYNTHETIC"

	TimingObserved = "OBSERVED_MONOTONIC"
	PhaseBefore    = "BEFORE"
	PhaseAfter     = "AFTER"
	OrderBaseline = "BASELINE_FIRST"
	OrderGooo     = "GOOO_FIRST"

	StateCompleted = "COMPLETED"
	StateAbandoned = "ABANDONED"
)

var (
	ErrInvalidPacket  = errors.New("invalid trial packet")
	ErrInvalidReceipt = errors.New("invalid session receipt")
)

type ReleaseReference struct {
	Source          string `json:"source"`
	Repository      string `json:"repository"`
	Tag             string `json:"tag"`
	ReleaseID       int64  `json:"release_id"`
	Immutable       bool   `json:"immutable"`
	TagObjectSHA    string `json:"tag_object_sha"`
	TargetCommitSHA string `json:"target_commit_sha"`
	Asset           Asset  `json:"asset"`
	Authority       string `json:"authority"`
}

type Asset struct {
	ID     int64  `json:"id"`
	Name   string `json:"name"`
	Size   int64  `json:"size"`
	SHA256 string `json:"sha256"`
}

type ContractReference struct {
	Schema   string `json:"schema"`
	Path     string `json:"path"`
	SHA256   string `json:"sha256"`
	CellCount int   `json:"cell_count"`
}

type FixtureReference struct {
	Schema string `json:"schema"`
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
}

type ToolchainReference struct {
	GoVersion string `json:"go_version"`
	Product   string `json:"product"`
	SHA256    string `json:"sha256"`
}

type Counts struct {
	Cells              int            `json:"cells"`
	Activities         int            `json:"activities"`
	ProofFoundation    int            `json:"proof_foundation"`
	ProofCoherence     int            `json:"proof_coherence"`
	ProofRegression    int            `json:"proof_regression"`
	IndicatorDriver    int            `json:"indicator_driver"`
	IndicatorOutcome   int            `json:"indicator_outcome"`
	IndicatorGuardrail int            `json:"indicator_guardrail"`
	Closed             int            `json:"closed"`
	Unknown            int            `json:"unknown"`
	Refuted            int            `json:"refuted"`
	Precedence         []string       `json:"precedence"`
}

type Artifact struct {
	Path   string `json:"path"`
	Kind   string `json:"kind"`
	SHA256 string `json:"sha256"`
	Bytes  int64  `json:"bytes"`
}

type Manifest struct {
	Schema                  string             `json:"schema"`
	ProtocolVersion         string             `json:"protocol_version"`
	Product                 string             `json:"product"`
	Release                 ReleaseReference   `json:"release"`
	Fixture                 FixtureReference   `json:"fixture"`
	Contract                ContractReference  `json:"contract"`
	Toolchain               ToolchainReference `json:"toolchain"`
	OracleDigest             string             `json:"oracle_digest"`
	Counts                  Counts             `json:"counts"`
	Counterbalance          CounterbalanceRule  `json:"counterbalance"`
	EvidenceOriginPolicy    OriginPolicy       `json:"evidence_origin_policy"`
	PairPolicy              PairPolicy         `json:"pair_policy"`
	Artifacts               []Artifact         `json:"artifacts"`
	UtilityState             string             `json:"utility_state"`
	ProtocolReady            string             `json:"protocol_ready"`
	ProcessState             string             `json:"process_state"`
	ExternalEvidenceCount    int                `json:"external_evidence_count"`
	Authority                Authority          `json:"authority"`
}

type CounterbalanceRule struct {
	Algorithm string `json:"algorithm"`
	Seed      string `json:"seed"`
	Even      string `json:"even"`
	Odd       string `json:"odd"`
}

type OriginPolicy struct {
	Eligible []string `json:"eligible"`
	ProtocolTestOnly []string `json:"protocol_test_only"`
	PIICollected bool `json:"pii_collected"`
	SessionIdentity string `json:"session_identity"`
}

type PairPolicy struct {
	RequiredFields []string `json:"required_fields"`
	RequiredAssignments []string `json:"required_assignments"`
	RequiredPhases []string `json:"required_phases"`
	NoPairState string `json:"no_pair_state"`
}

type Authority struct {
	SourceRepositoryWrites int `json:"source_repository_writes"`
	ProductRepositoryWrites int `json:"product_repository_writes"`
	LocalTestExecutions int `json:"local_test_executions"`
	ExternalNetworkAccess int `json:"external_network_access"`
	OpenTofuApplyExecutions int `json:"opentofu_apply_executions"`
	CloudMutations int `json:"cloud_mutations"`
}

type Cell struct {
	Ordinal     int      `json:"ordinal"`
	ID          string   `json:"id"`
	Activity    string   `json:"activity"`
	Stage       string   `json:"stage"`
	Step        string   `json:"step"`
	ProofFamily string   `json:"proof_family"`
	Indicator   string   `json:"indicator"`
	CaseState   string   `json:"case_state"`
	DependsOn   []string `json:"depends_on"`
}

type ConsentRetentionPolicy struct {
	ConsentGiven bool   `json:"consent_given"`
	RetentionPolicy string `json:"retention_policy"`
	RetentionDays int `json:"retention_days"`
}

type PairMetadata struct {
	PairID              string `json:"pair_id"`
	Phase               string `json:"phase"`
	CounterbalanceOrder string `json:"counterbalance_order"`
	PairEligible        bool   `json:"pair_eligible"`
}

type Receipt struct {
	Schema                 string                 `json:"schema"`
	ReceiptID              string                 `json:"receipt_id"`
	SessionNonce           string                 `json:"session_nonce"`
	TaskID                 string                 `json:"task_id"`
	Assignment             string                 `json:"assignment"`
	FixtureDigest          string                 `json:"fixture_digest"`
	ToolchainDigest        string                 `json:"toolchain_digest"`
	ContractDigest         string                 `json:"contract_digest"`
	OracleDigest           string                 `json:"oracle_digest"`
	InputDigest            string                 `json:"input_digest"`
	DecisionDigest         string                 `json:"decision_digest"`
	Correctness            int                    `json:"correctness"`
	ElapsedMS              int64                  `json:"elapsed_ms"`
	TimingKind             string                 `json:"timing_kind"`
	Estimated              bool                   `json:"estimated"`
	AbandonmentState       string                 `json:"abandonment_state"`
	EnvironmentDigest      string                 `json:"environment_digest"`
	Origin                 string                 `json:"origin"`
	ConsentRetentionPolicy ConsentRetentionPolicy `json:"consent_retention_policy"`
	UnknownFrontier        UnknownFrontier        `json:"unknown_frontier"`
	BeforeAfterPair        PairMetadata           `json:"before_after_pair"`
}

type EvidenceLine struct {
	Receipt        Receipt `json:"receipt"`
	ReceiptDigest  string  `json:"receipt_digest"`
	UtilityEligible bool   `json:"utility_eligible"`
	PairEligible   bool    `json:"pair_eligible"`
}

type RejectedReceipt struct {
	ObservedAtDigest string `json:"observed_at_digest"`
	ReceiptID        string `json:"receipt_id,omitempty"`
	Reason           string `json:"reason"`
}

type Oracle struct {
	Schema       string         `json:"schema"`
	FixtureDigest string        `json:"fixture_digest"`
	Tasks        []OracleTask   `json:"tasks"`
}

type OracleTask struct {
	TaskID             string            `json:"task_id"`
	CausalState        string            `json:"causal_state"`
	OracleDecision     string            `json:"oracle_decision"`
	OracleDecisionDigest string          `json:"oracle_decision_digest"`
	Assignments        []OracleAssignment `json:"assignments"`
	UnknownFrontier    *UnknownFrontier `json:"unknown_frontier,omitempty"`
}

type OracleAssignment struct {
	Assignment string `json:"assignment"`
	InputPath  string `json:"input_path"`
	InputDigest string `json:"input_digest"`
}

type UnknownFrontier struct {
	Stage         string   `json:"stage"`
	Step          string   `json:"step"`
	Reason        string   `json:"reason"`
	UnknownClass  string   `json:"unknown_class"`
	NextOperation string   `json:"next_operation"`
	BlockedBy     []string `json:"blocked_by"`
}

type RuntimeMetrics struct {
	BuildWallMS       *int64 `json:"build_wall_ms"`
	TestWallMS        *int64 `json:"test_wall_ms"`
	ConformanceWallMS *int64 `json:"conformance_wall_ms"`
	BuildRSSKiB       *int64 `json:"build_rss_kib"`
	TestRSSKiB        *int64 `json:"test_rss_kib"`
	ConformanceRSSKiB *int64 `json:"conformance_rss_kib"`
	ExecutedTests     int64 `json:"executed_tests"`
	ReusedTests       int64 `json:"reused_tests"`
	SkippedTests      int64 `json:"skipped_tests"`
	NotObservedTests  int64 `json:"not_observed_tests"`
	GoLines           int64 `json:"go_lines"`
	GoooLines         int64 `json:"gooo_lines"`
	RegularFiles      int64 `json:"regular_files"`
	DescendantDirs    int64 `json:"descendant_directories"`
}

func Release() ReleaseReference {
	return ReleaseReference{
		Source: "immutable_github_release",
		Repository: UpstreamRepository,
		Tag: UpstreamTag,
		ReleaseID: UpstreamReleaseID,
		Immutable: true,
		TagObjectSHA: UpstreamTagObject,
		TargetCommitSHA: UpstreamCommit,
		Asset: Asset{ID: UpstreamAssetID, Name: UpstreamAssetName, Size: UpstreamAssetSize, SHA256: UpstreamAssetSHA},
		Authority: "github_api_release_immutability_and_asset_digest",
	}
}

func sha256Digest(data []byte) string {
	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func validDigest(value string) bool {
	if len(value) != len("sha256:")+64 || !strings.HasPrefix(value, "sha256:") {
		return false
	}
	_, err := hex.DecodeString(strings.TrimPrefix(value, "sha256:"))
	return err == nil
}

func canonicalJSON(value any) ([]byte, error) {
	return json.Marshal(value)
}

func canonicalDigest(value any) (string, error) {
	data, err := canonicalJSON(value)
	if err != nil {
		return "", err
	}
	// Normalize structs through an interface round-trip so a digest does not
	// depend on whether the same JSON was produced from a struct or a map.
	var normalized any
	if err := json.Unmarshal(data, &normalized); err != nil {
		return "", err
	}
	data, err = canonicalJSON(normalized)
	if err != nil {
		return "", err
	}
	return sha256Digest(data), nil
}

func digestFile(path string) (string, int64, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", 0, err
	}
	info, err := os.Stat(path)
	if err != nil {
		return "", 0, err
	}
	if strings.HasSuffix(path, ".json") {
		var value any
		if err := json.Unmarshal(data, &value); err != nil {
			return "", 0, err
		}
		data, err = canonicalJSON(value)
		if err != nil {
			return "", 0, err
		}
	}
	return sha256Digest(data), info.Size(), nil
}

func writeJSON(path string, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(path, data, 0o644)
}

func cleanOutputPath(path string) (string, error) {
	if strings.TrimSpace(path) == "" {
		return "", fmt.Errorf("%w: output path is required", ErrInvalidPacket)
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("%w: output path: %v", ErrInvalidPacket, err)
	}
	if info, statErr := os.Stat(abs); statErr == nil {
		if !info.IsDir() {
			return "", fmt.Errorf("%w: output is not a directory", ErrInvalidPacket)
		}
		entries, readErr := os.ReadDir(abs)
		if readErr != nil {
			return "", fmt.Errorf("%w: inspect output: %v", ErrInvalidPacket, readErr)
		}
		if len(entries) != 0 {
			return "", fmt.Errorf("%w: output directory must be empty", ErrInvalidPacket)
		}
		return abs, nil
	}
	if err := os.MkdirAll(abs, 0o755); err != nil {
		return "", fmt.Errorf("%w: create output: %v", ErrInvalidPacket, err)
	}
	return abs, nil
}

func readStrictJSON(path string, target any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	decoder := json.NewDecoder(strings.NewReader(string(data)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return errors.New("more than one JSON value")
	}
	return nil
}

func sortedKeys[T any](values map[string]T) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
