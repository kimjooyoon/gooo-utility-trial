# Gooo Utility Trial

An independent, fail-closed user-trial packet for measuring whether a released
Gooo human-readable evidence dossier helps people make two exact decisions:
OpenTofu drift judgment and resuming a Gooo `UNKNOWN` frontier.

The canonical packet is generated from the immutable GitHub release
`kimjooyoon/gooo-opentofu-envelope@v0.1.1` and its plan-oracle fixture. The
release's `immutable`, tag, asset, and SHA-256 facts are checked through the
GitHub API in Actions; the packet's embedded release reference is not treated
as authority. No OpenTofu provider, credential, cloud, apply, state mutation,
or source-repository write is needed.

## What is measured

* U1: the same plan state is shown as baseline raw machine output or as a Gooo
  human dossier. The exact oracle token is `DRIFT_PRESENT`.
* U2: the same causal `EXTERNAL_UTILITY` UNKNOWN is shown as a baseline
  evidence set or as Gooo's six-field frontier. The exact next action is
  `PROVIDE_INDEPENDENT_EXTERNAL_UTILITY_EVIDENCE`.

Both conditions are recorded for the same pseudonymous session nonce. The
counterbalanced order is deterministic from the session nonce, task, fixture,
toolchain, contract, and oracle digests. A utility numerator requires a
complete baseline/gooo before/after pair with all of those fields equal.

Receipts contain no name, email, free-form answer, or rating. Correctness is an
integer checked against the oracle digest; timing is a non-negative observed
monotonic integer, never an estimate. `EXTERNAL_USER` and explicitly isolated
`INDEPENDENT_SESSION` are eligible. CI, maintainer self-tests, and synthetic
records may test the validator but always contribute zero to utility evidence.

## Commands

Generate into a fresh caller-owned directory:

```sh
go run ./cmd/gooo-utility-trial packet -output /tmp/gooo-utility-trial
```

Validate a receipt without writing anything:

```sh
go run ./cmd/gooo-utility-trial validate \
  -packet /tmp/gooo-utility-trial -receipt /path/to/receipt.json
```

Record a verified receipt and refresh the report:

```sh
go run ./cmd/gooo-utility-trial record \
  -packet /tmp/gooo-utility-trial -receipt /path/to/receipt.json
```

See [the protocol](docs/protocol.md) for the receipt fields and pairing rules.

With zero external receipts, the canonical result is deliberately:

```text
protocol_ready = CLOSED
utility         = UNKNOWN
utility improvement = UNKNOWN
```

The repository began as a three-file bootstrap (`.gitignore`, `LICENSE`, and
this README); implementation is delivered through a reviewed pull request and
verified by GitHub Actions. Failed Actions runs retain their evidence bundle.
