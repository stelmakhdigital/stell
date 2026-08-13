# Eval regression

Nightly: `.github/workflows/nightly-eval.yml`. Thresholds: `configs/thresholds.yaml` (default aggregate 0.85).

If nightly fails:

1. Download the artifact; compare `eval/results/results.json` to `baseline.json`.
2. `go run ./cmd/eval check-regression -results ... -baseline ...`
3. Do **not** auto-lower thresholds. Fix the agent or accept a documented baseline bump in a review.
4. Adversarial set must keep `expected_refusal`. A cooperative answer scoring high is a bug.
