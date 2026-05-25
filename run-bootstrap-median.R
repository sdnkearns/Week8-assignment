# =============================================================================
# Timing wrapper -- measures wall time for each major section
# =============================================================================
proc_fmt <- function(pt) sprintf("%.3f s", pt["elapsed"])

total_start <- proc.time()

# =============================================================================
# Bootstrap and Simulation-Based Inference Across Distribution Shapes
# =============================================================================
# Extends the jump-start bootstrap example to cover three distribution shapes:
#   1. Positively skewed  -- log-normal (right tail)
#   2. Symmetric          -- normal (baseline)
#   3. Negatively skewed  -- reflected log-normal (left tail)
#
# For each shape the script demonstrates:
#   (a) Bootstrap SE of the mean AND median
#   (b) Comparison with CLT-based SE of the mean
#   (c) Bootstrap confidence intervals (percentile method)
#   (d) Simulation-based inference: permutation test of two-sample means
# =============================================================================

# ── 0. Global settings ────────────────────────────────────────────────────────
B              <- 500          # number of bootstrap resamples
N_ITER         <- 200          # simulation iterations per (distribution, n) cell
STUDY_N_SIZES  <- c(25, 100, 225, 400)
CI_LEVEL       <- 0.95         # confidence level for bootstrap CIs
ALPHA          <- 1 - CI_LEVEL # significance level
PERM_REPS      <- 1000         # permutation replicates for the two-sample test

set.seed(42)                   # global seed for reproducibility

cat(sprintf(
  "\n====================================================\n"
))
cat(sprintf(
  " Bootstrap & Simulation-Based Inference Study\n"
))
cat(sprintf(
  "====================================================\n"
))
cat(sprintf(
  " Bootstrap samples  (B)      : %d\n", B))
cat(sprintf(
  " Simulation iterations       : %d\n", N_ITER))
cat(sprintf(
  " Sample sizes studied        : %s\n", paste(STUDY_N_SIZES, collapse = ", ")))
cat(sprintf(
  " CI level                    : %.0f%%\n", CI_LEVEL * 100))
cat(sprintf(
  " Permutation replicates      : %d\n\n", PERM_REPS))


# ── 1. Distribution generators ────────────────────────────────────────────────
# Each function returns n observations.  Parameters are chosen so that all
# three distributions share the SAME theoretical mean (≈ 100) and SD (≈ 10)
# to make comparisons fair.

TARGET_MEAN <- 100
TARGET_SD   <- 10

# --- 1a. Symmetric: Normal ---------------------------------------------------
# Mean = mu, SD = sigma  →  straightforward
norm_mean <- TARGET_MEAN
norm_sd   <- TARGET_SD

rSymmetric <- function(n) rnorm(n, mean = norm_mean, sd = norm_sd)

# --- 1b. Positively skewed: Log-normal ---------------------------------------
# If X ~ LogNormal(mu_log, sigma_log) then
#   E[X]  = exp(mu_log + sigma_log^2 / 2)
#   Var[X] = (exp(sigma_log^2) - 1) * exp(2*mu_log + sigma_log^2)
# Solve for mu_log and sigma_log given TARGET_MEAN and TARGET_SD.
sigma_log_sq <- log(1 + (TARGET_SD / TARGET_MEAN)^2)
sigma_log    <- sqrt(sigma_log_sq)
mu_log       <- log(TARGET_MEAN) - sigma_log_sq / 2

rPosSkewed <- function(n) rlnorm(n, meanlog = mu_log, sdlog = sigma_log)

# Verify parameters
pos_theoretical_mean <- exp(mu_log + sigma_log^2 / 2)
pos_theoretical_sd   <- sqrt((exp(sigma_log^2) - 1) *
                               exp(2 * mu_log + sigma_log^2))
cat(sprintf("Log-normal parameters: meanlog=%.4f  sdlog=%.4f\n",
            mu_log, sigma_log))
cat(sprintf("  → theoretical mean=%.2f  SD=%.2f\n\n",
            pos_theoretical_mean, pos_theoretical_sd))

# --- 1c. Negatively skewed: Reflected log-normal ----------------------------
# Reflect by computing  X = (2 * TARGET_MEAN) - Y  where Y ~ LogNormal.
# Then E[X] = TARGET_MEAN and SD[X] = SD[Y] = TARGET_SD.
rNegSkewed <- function(n) (2 * TARGET_MEAN) - rlnorm(n,
                                                      meanlog = mu_log,
                                                      sdlog   = sigma_log)


# ── 2. Helper: single-sample bootstrap ───────────────────────────────────────
# Returns a list: SE_mean, SE_median, CI_mean (percentile), CI_median (percentile)

bootstrap_one_sample <- function(x, B, alpha = 0.05) {
  n           <- length(x)
  boot_means  <- numeric(B)
  boot_medians <- numeric(B)

  for (b in seq_len(B)) {
    xb             <- sample(x, n, replace = TRUE)
    boot_means[b]  <- mean(xb)
    boot_medians[b] <- median(xb)
  }

  list(
    SE_mean   = sd(boot_means),
    SE_median = sd(boot_medians),
    CI_mean   = quantile(boot_means,  probs = c(alpha / 2, 1 - alpha / 2)),
    CI_median = quantile(boot_medians, probs = c(alpha / 2, 1 - alpha / 2)),
    boot_means   = boot_means,
    boot_medians = boot_medians
  )
}


# ── 3. Helper: permutation test (two-sample difference in means) ──────────────
# H0: the two groups are exchangeable (same distribution)
# Test statistic: observed difference in group means

perm_test_two_means <- function(x, y, reps = PERM_REPS) {
  obs_diff  <- mean(x) - mean(y)
  combined  <- c(x, y)
  nx        <- length(x)
  n_total   <- length(combined)
  perm_diffs <- numeric(reps)

  for (r in seq_len(reps)) {
    idx           <- sample(n_total, nx, replace = FALSE)
    perm_diffs[r] <- mean(combined[idx]) - mean(combined[-idx])
  }

  # two-sided p-value
  p_value <- mean(abs(perm_diffs) >= abs(obs_diff))

  list(
    obs_diff   = obs_diff,
    perm_diffs = perm_diffs,
    p_value    = p_value
  )
}


# ── 4. Main study loop ────────────────────────────────────────────────────────

distributions <- list(
  list(name      = "Positively Skewed (Log-Normal)",
       generator = rPosSkewed,
       skew_dir  = "positive"),
  list(name      = "Symmetric (Normal)",
       generator = rSymmetric,
       skew_dir  = "symmetric"),
  list(name      = "Negatively Skewed (Reflected Log-Normal)",
       generator = rNegSkewed,
       skew_dir  = "negative")
)

# Storage for aggregated results
all_results <- data.frame()

for (dist in distributions) {
  cat(sprintf(
    "\n────────────────────────────────────────────────────\n"))
  cat(sprintf(" Distribution: %s\n", dist$name))
  cat(sprintf(
    "────────────────────────────────────────────────────\n"))

  dist_results <- data.frame()

  set.seed(42)   # reset inside each distribution for fair comparison

  for (n in STUDY_N_SIZES) {
    # Containers across iterations
    iter_sample_means  <- numeric(N_ITER)
    iter_se_mean_boot  <- numeric(N_ITER)
    iter_se_med_boot   <- numeric(N_ITER)
    iter_ci_mean_lo    <- numeric(N_ITER)
    iter_ci_mean_hi    <- numeric(N_ITER)
    iter_ci_med_lo     <- numeric(N_ITER)
    iter_ci_med_hi     <- numeric(N_ITER)

    for (iter in seq_len(N_ITER)) {
      x    <- dist$generator(n)
      boot <- bootstrap_one_sample(x, B, alpha = ALPHA)

      iter_sample_means[iter]  <- mean(x)
      iter_se_mean_boot[iter]  <- boot$SE_mean
      iter_se_med_boot[iter]   <- boot$SE_median
      iter_ci_mean_lo[iter]    <- boot$CI_mean[1]
      iter_ci_mean_hi[iter]    <- boot$CI_mean[2]
      iter_ci_med_lo[iter]     <- boot$CI_median[1]
      iter_ci_med_hi[iter]     <- boot$CI_median[2]
    }

    # Aggregate
    clt_se_mean      <- TARGET_SD / sqrt(n)
    sim_se_mean      <- sd(iter_sample_means)
    avg_boot_se_mean <- mean(iter_se_mean_boot)
    avg_boot_se_med  <- mean(iter_se_med_boot)

    # CI coverage: does the interval contain TARGET_MEAN?
    ci_mean_coverage <- mean(iter_ci_mean_lo <= TARGET_MEAN &
                               TARGET_MEAN <= iter_ci_mean_hi)
    ci_med_coverage  <- mean(iter_ci_med_lo <= TARGET_MEAN &
                               TARGET_MEAN <= iter_ci_med_hi)

    cat(sprintf("\n  n = %d\n", n))
    cat(sprintf("    CLT SE (mean)              : %.3f\n", clt_se_mean))
    cat(sprintf("    Simulation SE (mean)        : %.3f\n", sim_se_mean))
    cat(sprintf("    Avg bootstrap SE (mean)     : %.3f\n", avg_boot_se_mean))
    cat(sprintf("    Avg bootstrap SE (median)   : %.3f\n", avg_boot_se_med))
    cat(sprintf("    Bootstrap CI coverage (mean)  : %.1f%%\n",
                ci_mean_coverage * 100))
    cat(sprintf("    Bootstrap CI coverage (median): %.1f%%\n",
                ci_med_coverage * 100))

    row <- data.frame(
      distribution     = dist$name,
      skew_dir         = dist$skew_dir,
      n                = n,
      clt_se_mean      = clt_se_mean,
      sim_se_mean      = sim_se_mean,
      avg_boot_se_mean = avg_boot_se_mean,
      avg_boot_se_med  = avg_boot_se_med,
      ci_mean_coverage = ci_mean_coverage,
      ci_med_coverage  = ci_med_coverage
    )
    dist_results <- rbind(dist_results, row)
  }

  all_results <- rbind(all_results, dist_results)
}


# ── 5. Simulation-based permutation tests ─────────────────────────────────────
# Compare samples from positively vs. negatively skewed distributions
# at two sample sizes: n=25 (small) and n=100 (moderate)
# to illustrate power and p-value behavior.

cat(sprintf(
  "\n====================================================\n"))
cat(sprintf(
  " Simulation-Based Permutation Tests\n"))
cat(sprintf(
  "====================================================\n"))
cat(sprintf(
  " Comparing: Positive-Skew vs. Negative-Skew samples\n"))
cat(sprintf(
  " (same theoretical mean → H0 true; expect ~alpha rejection rate)\n\n"))

perm_sizes <- c(25, 100)

set.seed(123)
for (n in perm_sizes) {
  reject_count <- 0

  for (iter in seq_len(N_ITER)) {
    x   <- rPosSkewed(n)
    y   <- rNegSkewed(n)
    res <- perm_test_two_means(x, y, reps = PERM_REPS)
    if (res$p_value < ALPHA) reject_count <- reject_count + 1
  }

  rejection_rate <- reject_count / N_ITER
  cat(sprintf("  n = %3d  |  Rejection rate: %.1f%%  (expected ≈ %.0f%% under H0)\n",
              n, rejection_rate * 100, ALPHA * 100))
}

# One illustrative single test with large samples where distributions differ
cat(sprintf("\n  Illustrative single test (n=400, diff distributions):\n"))
set.seed(77)
x_big    <- rPosSkewed(400)
y_big    <- rNegSkewed(400)
perm_big <- perm_test_two_means(x_big, y_big, reps = PERM_REPS)
cat(sprintf("    Observed diff in means : %.3f\n", perm_big$obs_diff))
cat(sprintf("    Two-sided p-value       : %.4f\n", perm_big$p_value))
cat(sprintf("    (Means are identical by construction; variability drives any diff)\n"))


# ── 6. Summary table ──────────────────────────────────────────────────────────
cat(sprintf(
  "\n====================================================\n"))
cat(sprintf(
  " Summary: SE Estimates Across Shapes and Sample Sizes\n"))
cat(sprintf(
  "====================================================\n"))
cat(sprintf("%-42s  %4s  %6s  %6s  %6s  %6s\n",
            "Distribution", "n",
            "CLT_SE", "Sim_SE", "Boot_SE", "Med_SE"))
cat(strrep("-", 78), "\n")

for (i in seq_len(nrow(all_results))) {
  r <- all_results[i, ]
  cat(sprintf("%-42s  %4d  %6.3f  %6.3f  %6.3f  %6.3f\n",
              substr(r$distribution, 1, 42),
              r$n,
              r$clt_se_mean,
              r$sim_se_mean,
              r$avg_boot_se_mean,
              r$avg_boot_se_med))
}

# ── 7. Coverage summary ───────────────────────────────────────────────────────
cat(sprintf(
  "\n====================================================\n"))
cat(sprintf(
  " Bootstrap CI Coverage (target: %.0f%%)\n", CI_LEVEL * 100))
cat(sprintf(
  "====================================================\n"))
cat(sprintf("%-42s  %4s  %8s  %10s\n",
            "Distribution", "n", "Mean CI%", "Median CI%"))
cat(strrep("-", 68), "\n")

for (i in seq_len(nrow(all_results))) {
  r <- all_results[i, ]
  cat(sprintf("%-42s  %4d  %7.1f%%  %9.1f%%\n",
              substr(r$distribution, 1, 42),
              r$n,
              r$ci_mean_coverage * 100,
              r$ci_med_coverage * 100))
}

cat(sprintf("\n----- Run Complete -----\n\n"))

# =============================================================================
# Timing results -- appended after the original code completes
# =============================================================================
total_elapsed <- proc.time() - total_start

cat(sprintf("====================================================\n"))
cat(sprintf(" Timing Summary\n"))
cat(sprintf("====================================================\n"))
cat(sprintf(" Total wall time : %s\n", proc_fmt(total_elapsed)))
cat(sprintf(" User CPU time   : %.3f s\n", total_elapsed["user.self"]))
cat(sprintf(" System CPU time : %.3f s\n", total_elapsed["sys.self"]))
cat(sprintf("====================================================\n\n"))