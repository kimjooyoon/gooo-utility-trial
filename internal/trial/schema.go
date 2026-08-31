package trial

const receiptSchemaJSON = `{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "$id": "https://github.com/kimjooyoon/gooo-utility-trial/schema/session-receipt/v1",
  "title": "Gooo independent utility trial session receipt",
  "type": "object",
  "additionalProperties": false,
  "required": ["schema", "receipt_id", "session_nonce", "task_id", "assignment", "fixture_digest", "toolchain_digest", "contract_digest", "oracle_digest", "input_digest", "decision_digest", "correctness", "elapsed_ms", "timing_kind", "estimated", "abandonment_state", "environment_digest", "origin", "consent_retention_policy", "before_after_pair"],
  "properties": {
    "schema": {"const": "gooo/utility-trial/session-receipt/v1"},
    "receipt_id": {"type": "string", "pattern": "^[0-9a-f]{32}$"},
    "session_nonce": {"type": "string", "pattern": "^[0-9a-f]{32}$"},
    "task_id": {"enum": ["U1_OPENTOFU_DRIFT", "U2_UNKNOWN_RESUME"]},
    "assignment": {"enum": ["U1_BASELINE_RAW", "U1_GOOO_DOSSIER", "U2_BASELINE_EVIDENCE_SET", "U2_GOOO_FRONTIER"]},
    "fixture_digest": {"type": "string", "pattern": "^sha256:[0-9a-f]{64}$"},
    "toolchain_digest": {"type": "string", "pattern": "^sha256:[0-9a-f]{64}$"},
    "contract_digest": {"type": "string", "pattern": "^sha256:[0-9a-f]{64}$"},
    "oracle_digest": {"type": "string", "pattern": "^sha256:[0-9a-f]{64}$"},
    "input_digest": {"type": "string", "pattern": "^sha256:[0-9a-f]{64}$"},
    "decision_digest": {"type": "string", "pattern": "^sha256:[0-9a-f]{64}$"},
    "correctness": {"type": "integer", "enum": [0, 1]},
    "elapsed_ms": {"type": "integer", "minimum": 0},
    "timing_kind": {"const": "OBSERVED_MONOTONIC"},
    "estimated": {"const": false},
    "abandonment_state": {"enum": ["COMPLETED", "ABANDONED"]},
    "environment_digest": {"type": "string", "pattern": "^sha256:[0-9a-f]{64}$"},
    "origin": {"enum": ["EXTERNAL_USER", "INDEPENDENT_SESSION", "CI", "MAINTAINER_SELF_TEST", "SYNTHETIC"]},
    "consent_retention_policy": {
      "type": "object", "additionalProperties": false,
      "required": ["consent_given", "retention_policy", "retention_days"],
      "properties": {
        "consent_given": {"type": "boolean"},
        "retention_policy": {"type": "string", "minLength": 1},
        "retention_days": {"type": "integer", "minimum": 1}
      }
    },
    "before_after_pair": {
      "type": "object", "additionalProperties": false,
      "required": ["pair_id", "phase", "counterbalance_order", "pair_eligible"],
      "properties": {
        "pair_id": {"type": "string", "pattern": "^sha256:[0-9a-f]{64}$"},
        "phase": {"enum": ["BEFORE", "AFTER"]},
        "counterbalance_order": {"enum": ["BASELINE_FIRST", "GOOO_FIRST"]},
        "pair_eligible": {"type": "boolean"}
      }
    }
  }
}`

