# Independent utility trial protocol v1

The packet has two matched tasks:

* U1 presents the same OpenTofu plan state as either raw machine output or the
  Gooo human dossier. The oracle decision is the exact token `DRIFT_PRESENT`.
* U2 presents the same `EXTERNAL_UTILITY` causal state as either a baseline
  evidence set or Gooo's six-field UNKNOWN frontier. The oracle next action is
  the exact token `PROVIDE_INDEPENDENT_EXTERNAL_UTILITY_EVIDENCE`.

For each task, the participant uses both assignments in the deterministic
counterbalanced order recorded in the session receipt. The pair key binds the
same pseudonymous session nonce, task, fixture, toolchain, contract, and oracle
digest. Only a complete baseline/gooo pair can enter the utility numerator.

Receipts contain no name, email, free-form answer, rating, or other PII. A
receipt records only a digest of the exact decision token, a correctness
integer, observed monotonic elapsed milliseconds, abandonment state, origin,
consent/retention policy, the complete six-field UNKNOWN frontier, and
provenance digests.

`EXTERNAL_USER` and an explicitly isolated `INDEPENDENT_SESSION` are eligible
origins. `CI`, `MAINTAINER_SELF_TEST`, and `SYNTHETIC` records may exercise the
protocol validator but are always excluded from the utility numerator. With
zero eligible pairs, the report is `protocol_ready=CLOSED` and
`utility=UNKNOWN`.

## Use

Create a fresh caller-owned temporary directory:

```sh
go run ./cmd/gooo-utility-trial packet -output /tmp/gooo-utility-trial
```

Use the packet's exact digests and oracle decision tokens to create one receipt
per assignment, then verify without writing:

```sh
go run ./cmd/gooo-utility-trial validate \
  -packet /tmp/gooo-utility-trial \
  -receipt /path/to/receipt.json
```

Record a verified receipt and refresh the report:

```sh
go run ./cmd/gooo-utility-trial record \
  -packet /tmp/gooo-utility-trial \
  -receipt /path/to/receipt.json
```

The validator fails closed for malformed origins, digest mismatches, a missing
or mismatched pair, negative or estimated timing, oracle/correctness mismatch,
and replayed receipt IDs or contents.
