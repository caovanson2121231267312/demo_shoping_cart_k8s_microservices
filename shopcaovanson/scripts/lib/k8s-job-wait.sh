# shellcheck shell=bash
# Wait for a Kubernetes Job with live progress output.
# Usage: source scripts/lib/k8s-job-wait.sh && wait_for_k8s_job <job-name> [namespace] [timeout_secs]

wait_for_k8s_job() {
  local job_name="$1"
  local ns="${2:-${NAMESPACE:-shop}}"
  local timeout="${3:-${JOB_TIMEOUT:-300}}"
  local interval="${JOB_POLL_INTERVAL:-5}"
  local elapsed=0
  local prefix="${JOB_LOG_PREFIX:-[k8s-job]}"

  while [[ "${elapsed}" -lt "${timeout}" ]]; do
    local succeeded failed
    succeeded=$(kubectl get job "${job_name}" -n "${ns}" -o jsonpath='{.status.succeeded}' 2>/dev/null || echo "")
    failed=$(kubectl get job "${job_name}" -n "${ns}" -o jsonpath='{.status.failed}' 2>/dev/null || echo "0")

    if [[ "${succeeded}" == "1" ]]; then
      echo "${prefix} ✓ job/${job_name} completed (${elapsed}s)"
      return 0
    fi

    local backoff
    backoff=$(kubectl get job "${job_name}" -n "${ns}" -o jsonpath='{.status.conditions[?(@.type=="Failed")].status}' 2>/dev/null || echo "")
    if [[ "${backoff}" == "True" ]] || [[ "${failed}" =~ ^[1-9] ]]; then
      echo "${prefix} ERROR: job/${job_name} failed (failed=${failed})"
      kubectl get pods -n "${ns}" -l "job-name=${job_name}" -o wide 2>/dev/null || true
      echo "${prefix} --- logs ---"
      kubectl logs -n "${ns}" -l "job-name=${job_name}" --tail=80 2>/dev/null || true
      echo "${prefix} --- events ---"
      kubectl describe job "${job_name}" -n "${ns}" 2>/dev/null | sed -n '/Events:/,$p' | head -15 || true
      return 1
    fi

    local pod phase reason restarts
    pod=$(kubectl get pods -n "${ns}" -l "job-name=${job_name}" -o jsonpath='{.items[0].metadata.name}' 2>/dev/null || true)
    if [[ -n "${pod}" ]]; then
      phase=$(kubectl get pod "${pod}" -n "${ns}" -o jsonpath='{.status.phase}' 2>/dev/null || echo "?")
      reason=$(kubectl get pod "${pod}" -n "${ns}" -o jsonpath='{.status.containerStatuses[0].state.waiting.reason}' 2>/dev/null || true)
      restarts=$(kubectl get pod "${pod}" -n "${ns}" -o jsonpath='{.status.containerStatuses[0].restartCount}' 2>/dev/null || echo "0")
      echo "${prefix}   ... ${job_name}: pod=${pod} phase=${phase} restarts=${restarts} ${reason:+wait=${reason}} (${elapsed}s/${timeout}s)"
    else
      echo "${prefix}   ... ${job_name}: chờ pod được tạo (${elapsed}s/${timeout}s)"
    fi

    sleep "${interval}"
    elapsed=$((elapsed + interval))
  done

  echo "${prefix} ERROR: timeout job/${job_name} sau ${timeout}s"
  kubectl get pods -n "${ns}" -l "job-name=${job_name}" -o wide 2>/dev/null || true
  kubectl logs -n "${ns}" -l "job-name=${job_name}" --tail=80 2>/dev/null || true
  return 1
}
