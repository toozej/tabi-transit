# Vehicle-history maximum-page benchmark

The rider-visible history endpoint accepts at most 500 observations per page.
This benchmark measures the API work for that maximum permitted page with a
deterministic normalized fixture. It is a repeatable local baseline, not a
production capacity claim.

## What is measured

`make test-history-benchmark` runs
`BenchmarkVehicleHistoryMaximumPage` five times with:

- 500 observations, each spaced 30 seconds apart;
- source-qualified vehicle, route, and trip identifiers;
- stable Portland-area coordinates and `fresh` timestamps;
- route assignment on 80% of observations and trip assignment on 75%, matching
  the optional fields emitted by a normalized real-time history page.

The benchmark uses the public HTTP handler and an in-memory history store. It
therefore includes validation, response mapping, JSON encoding, ETag hashing,
and response allocation. `TestVehicleHistoryMaximumPageBenchmarkFixture`
prevents the benchmark fixture from silently shrinking below the endpoint's
maximum response size.

## Reproducing the baseline

```sh
make test-history-benchmark
```

For a comparison between revisions, keep the same host and Go toolchain, then
save the output from the command above and compare it with `benchstat` if it
is installed. The benchmark itself reports nanoseconds/op, bytes/op, and
allocations/op; it intentionally does not set a timing threshold because those
values are host-dependent.

## Recorded local baseline

On 2026-07-28, using Go 1.26.0 on an Intel Core Ultra 7 265K Linux host, five
independent runs reported 0.81–1.06 ms/op, 0.97–0.98 MB/op, and 15,406
allocations/op. Those results cover the maximum-page handler path only and are
recorded for regression comparison; they are not a production latency target.

## Scope and limitations

This does **not** measure PostgreSQL query planning or index use, database
connection contention, retained-history cardinality, HTTP/TLS/network latency,
concurrent clients, or mobile parsing/rendering. It cannot support a
release-scale or production-SLO claim by itself.

Before making such a claim, run a separate controlled PostGIS measurement with
production-shaped retained-row counts and query plans, then a load test through
the deployed HTTP/TLS path. Record database sizing, concurrency, percentile
latency, payload size, and the tested retention/cadence assumptions with that
environment's results.
