package tasks

import (
    "sync"
    "github.com/prometheus/client_golang/prometheus"
)

var (
    metricsOnce sync.Once
    reconcileRuns = prometheus.NewCounter(prometheus.CounterOpts{
        Name: "flashsale_reconcile_runs_total",
        Help: "Total reconciliation runs",
    })
    reconcileInconsistent = prometheus.NewCounter(prometheus.CounterOpts{
        Name: "flashsale_reconcile_inconsistent_total",
        Help: "Total times inconsistencies were detected",
    })
    reconcileCompSent = prometheus.NewCounter(prometheus.CounterOpts{
        Name: "flashsale_reconcile_compensations_sent_total",
        Help: "Total compensation events sent",
    })
    reconcileCompSkipped = prometheus.NewCounterVec(prometheus.CounterOpts{
        Name: "flashsale_reconcile_compensation_skipped_total",
        Help: "Compensation skipped with reasons",
    }, []string{"reason"})
    reconcileDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
        Name:    "flashsale_reconcile_duration_seconds",
        Help:    "Reconciliation loop duration",
        Buckets: prometheus.DefBuckets,
    })

    stockLogRows = prometheus.NewCounter(prometheus.CounterOpts{
        Name: "flashsale_stocklog_rows_total",
        Help: "Total stock log rows written to DB",
    })
    stockLogErrors = prometheus.NewCounter(prometheus.CounterOpts{
        Name: "flashsale_stocklog_errors_total",
        Help: "Total stock log write errors",
    })
)

func initTaskMetrics() {
    metricsOnce.Do(func() {
        // ignore AlreadyRegistered errors
        _ = prometheus.Register(reconcileRuns)
        _ = prometheus.Register(reconcileInconsistent)
        _ = prometheus.Register(reconcileCompSent)
        _ = prometheus.Register(reconcileCompSkipped)
        _ = prometheus.Register(reconcileDuration)
        _ = prometheus.Register(stockLogRows)
        _ = prometheus.Register(stockLogErrors)
    })
}

