# Release-history rewrite dossier

`RELEASE_HISTORY_REWRITE_PROCESS=REFUTED` is a historical process finding,
not a utility result. The failed `v0.1.0` publication run `33407273856`
produced artifact `9763659711` and release `379848683` with `immutable=false`.
Its assets were:

| id | size | digest |
|---:|---:|---|
| 538154567 | 11377 | `sha256:1350c51f5f7db9dc2c6ac64523229f75cf7ec9ebdffaf99aa2c6edf32a40aa72` |
| 538154571 | 104 | `sha256:1848476904a222e9fab2fbe2186ba7315ca0fb1df247ea5ddfb5d9462f1997ba` |

That release was deleted and recreated. The replacement run `33407562271`
and artifact `9763767396` produced release `379850805` with `immutable=true`.
Its current assets were:

| id | size | digest |
|---:|---:|---|
| 538157619 | 11376 | `sha256:734082b840e915b48c42e14e93252624f5e441538509ff29dafe259b851f9a9e` |
| 538157605 | 104 | `sha256:152398724ab80d4ad5dbaaa040fc2d63261431322d9344d1305e6c840bbf5ffa` |

Both historical attempts point to tag object
`e30bca521d127d929043557198557710d35afcd2` and target commit
`6521e699f1e1180b7e942ae18d0948383c3d544e`. The failed run and both sets of
assets remain preserved as evidence. `v0.1.0`, its tag, and its current assets
are read-only and must never be touched again.

## Independent result axes

| axis | result |
|---|---|
| protocol_ready | `CLOSED` |
| utility | `UNKNOWN` |
| external evidence | `0` |
| eligible pairs | `0` |
| process | `REFUTED` |
| combined score | `NOT_COMBINED` |

The process receipt is separate non-score evidence. The protocol denominator
remains exactly 12 cells and 12 activities. No process cell was added, so the
denominator migration is explicitly `NONE` with delta `0`; process state is
never averaged with readiness or utility.
