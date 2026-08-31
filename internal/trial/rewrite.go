package trial

// releaseHistoryRewriteValue is a separate, non-score process receipt. It is
// intentionally not a Cell and therefore does not alter the fixed protocol
// denominator.
func releaseHistoryRewriteValue() map[string]any {
	return map[string]any{
		"schema": "gooo/utility-trial/release-history-rewrite-process-receipt/v1",
		"activity": "RecordReleaseHistoryRewriteProcess",
		"release_history_rewrite_process": "REFUTED",
		"state": "REFUTED",
		"score_included": false,
		"protocol_denominator_delta": 0,
		"denominator_migration": map[string]any{
			"operation": "NONE",
			"protocol_cells_before": 12,
			"protocol_cells_after": 12,
			"protocol_activities_before": 12,
			"protocol_activities_after": 12,
			"process_cells_added": 0,
			"hidden_mutation": false,
		},
		"failed_attempt": map[string]any{
			"run_id": 33407273856,
			"artifact_id": 9763659711,
			"release_id": 379848683,
			"immutable": false,
			"tag": "v0.1.0",
			"tag_object_sha": "e30bca521d127d929043557198557710d35afcd2",
			"target_commit_sha": "6521e699f1e1180b7e942ae18d0948383c3d544e",
			"assets": []map[string]any{
				{"id": 538154567, "name": "gooo-utility-trial-0.1.0-evidence.zip", "size": 11377, "digest": "sha256:1350c51f5f7db9dc2c6ac64523229f75cf7ec9ebdffaf99aa2c6edf32a40aa72"},
				{"id": 538154571, "name": "SHA256SUMS", "size": 104, "digest": "sha256:1848476904a222e9fab2fbe2186ba7315ca0fb1df247ea5ddfb5d9462f1997ba"},
			},
		},
		"replacement_attempt": map[string]any{
			"run_id": 33407562271,
			"artifact_id": 9763767396,
			"release_id": 379850805,
			"immutable": true,
			"tag": "v0.1.0",
			"tag_object_sha": "e30bca521d127d929043557198557710d35afcd2",
			"target_commit_sha": "6521e699f1e1180b7e942ae18d0948383c3d544e",
			"assets": []map[string]any{
				{"id": 538157619, "name": "gooo-utility-trial-0.1.0-evidence.zip", "size": 11376, "digest": "sha256:734082b840e915b48c42e14e93252624f5e441538509ff29dafe259b851f9a9e"},
				{"id": 538157605, "name": "SHA256SUMS", "size": 104, "digest": "sha256:152398724ab80d4ad5dbaaa040fc2d63261431322d9344d1305e6c840bbf5ffa"},
			},
		},
		"reason": "FAILED_RELEASE_WAS_DELETED_AND_RECREATED",
		"next_operation": "PRESERVE_RELEASE_HISTORY_AND_CREATE_NEXT_PATCH",
		"preservation_policy": "v0.1.0 tag and all current and failed release assets are read-only historical evidence",
	}
}
