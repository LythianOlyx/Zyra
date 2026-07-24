package render

import "github.com/prometheus/client_golang/prometheus"

// ssrFallbackTotal counts every "ssr" render that fell back to serving the
// CSR shell instead of failing the request, per
// 03-RENDERING-ENGINE.md ("Fallback wajib"). Labeled by route and whether
// the cause was a render timeout, so timeouts (usually a sign of slow
// upstream data or a runaway bundle) can be distinguished from generic
// JavaScript errors in dashboards/alerts.
var ssrFallbackTotal = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "zyra_render_ssr_fallback_total",
		Help: "Total number of SSR renders that fell back to the CSR shell instead of failing the request.",
	},
	[]string{"route", "timeout"},
)

func init() {
	prometheus.MustRegister(ssrFallbackTotal)
}
