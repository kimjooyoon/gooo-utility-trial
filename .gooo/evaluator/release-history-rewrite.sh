#!/usr/bin/env bash
set -euo pipefail

root=${1:?repository root is required}
output=${2:?output path is required}
receipt="$root/.gooo/activity/release-history-rewrite-process-receipt.json"
contract="$root/.gooo/contract/release-history-rewrite-v1.json"
generated="$root/.gooo/generated/evaluator/release-history-rewrite.json"
counterexample="$root/.gooo/counterexamples/release-history-rewrite.json"

test -s "$receipt"
test -s "$contract"
test -s "$generated"
test -s "$counterexample"
jq -e '
  .release_history_rewrite_process == "REFUTED" and
  .state == "REFUTED" and
  .score_included == false and
  .protocol_denominator_delta == 0 and
  .failed_attempt.run_id == 33407273856 and
  .failed_attempt.artifact_id == 9763659711 and
  .failed_attempt.release_id == 379848683 and
  .failed_attempt.immutable == false and
  .replacement_attempt.run_id == 33407562271 and
  .replacement_attempt.artifact_id == 9763767396 and
  .replacement_attempt.release_id == 379850805 and
  .replacement_attempt.immutable == true
' "$receipt" >/dev/null
jq -e '
  .decision == "RELEASE_HISTORY_REWRITE_PROCESS=REFUTED" and
  .process_state == "REFUTED" and
  .score_included == false and
  .axes.protocol_ready == "CLOSED" and
  .axes.utility == "UNKNOWN" and
  .axes.external_evidence_count == 0 and
  .axes.eligible_pairs == 0 and
  .denominator_migration.operation == "NONE" and
  .denominator_migration.protocol_cells_before == 12 and
  .denominator_migration.protocol_cells_after == 12
' "$contract" >/dev/null
jq -e '
  .decision == "RELEASE_HISTORY_REWRITE_PROCESS=REFUTED" and
  .protocol_ready == "CLOSED" and
  .utility == "UNKNOWN" and
  .external_evidence_count == 0 and
  .eligible_pairs == 0 and
  .score == "NOT_COMBINED" and
  .failed_run_id == 33407273856 and .failed_artifact_id == 9763659711 and
  .old_release_id == 379848683 and .old_immutable == false and .old_asset_ids == [538154567,538154571] and
  .old_asset_sizes == [11377,104] and
  .old_asset_digests == ["sha256:1350c51f5f7db9dc2c6ac64523229f75cf7ec9ebdffaf99aa2c6edf32a40aa72","sha256:1848476904a222e9fab2fbe2186ba7315ca0fb1df247ea5ddfb5d9462f1997ba"] and
  .replacement_run_id == 33407562271 and .replacement_artifact_id == 9763767396 and
  .new_release_id == 379850805 and .new_immutable == true and .new_asset_ids == [538157619,538157605] and
  .new_asset_sizes == [11376,104] and
  .new_asset_digests == ["sha256:734082b840e915b48c42e14e93252624f5e441538509ff29dafe259b851f9a9e","sha256:152398724ab80d4ad5dbaaa040fc2d63261431322d9344d1305e6c840bbf5ffa"]
' "$generated" >/dev/null
jq -e '
  .decision == "REFUTED" and
  .release_history_rewrite_process == "REFUTED" and
  .counterexample.old_release_id == 379848683 and
  .counterexample.new_release_id == 379850805 and
  .denominator_migration.operation == "NONE" and
  .denominator_migration.delta == 0
' "$counterexample" >/dev/null

evaluator_digest="sha256:$(sha256sum "$root/.gooo/evaluator/release-history-rewrite.sh" | awk '{print $1}')"
receipt_digest="sha256:$(sha256sum "$receipt" | awk '{print $1}')"
counterexample_digest="sha256:$(sha256sum "$counterexample" | awk '{print $1}')"
jq -S -n \
  --arg evaluator_digest "$evaluator_digest" \
  --arg receipt_digest "$receipt_digest" \
  --arg counterexample_digest "$counterexample_digest" \
  --arg run_id "${GITHUB_RUN_ID:-LOCAL_NOT_A_PRODUCT_RUN}" \
  '{schema:"gooo/utility-trial/generated-evaluator-run/v1",
    decision:"RELEASE_HISTORY_REWRITE_PROCESS=REFUTED", process_state:"REFUTED",
    protocol_ready:"CLOSED", utility:"UNKNOWN", external_evidence_count:0,
    eligible_pairs:0, score:"NOT_COMBINED",
    denominator_migration:{operation:"NONE", protocol_cells_before:12, protocol_cells_after:12, process_cells_added:0, hidden_mutation:false},
    evaluator:{path:".gooo/evaluator/release-history-rewrite.sh", digest:$evaluator_digest},
    input_digests:{process_receipt:$receipt_digest, canonical_counterexample:$counterexample_digest},
    run_id:$run_id}' > "$output"
