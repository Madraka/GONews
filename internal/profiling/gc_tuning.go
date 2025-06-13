package profiling

import (
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof" // This automatically registers pprof handlers
	"os"
	"runtime"
	"runtime/debug"
	"time"

	"github.com/gin-gonic/gin"
)

// SetupPprof adds pprof endpoints to the router for performance profiling
// Hot reload test: This endpoint enables comprehensive Go profiling capabilities
func SetupPprof(router *gin.Engine) {
	// Only enable pprof in development or when explicitly enabled
	if gin.Mode() == gin.DebugMode || os.Getenv("ENABLE_PPROF") == "true" {
		log.Println("🔍 pprof profiling endpoints enabled - Hot reload working!")

		// Mount pprof endpoints under /debug/pprof/
		pprofGroup := router.Group("/debug/pprof")
		{
			pprofGroup.GET("/", gin.WrapF(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				http.Redirect(w, r, "/debug/pprof/", http.StatusMovedPermanently)
			})))
			pprofGroup.GET("/cmdline", gin.WrapF(http.DefaultServeMux.ServeHTTP))
			pprofGroup.GET("/profile", gin.WrapF(http.DefaultServeMux.ServeHTTP))
			pprofGroup.GET("/symbol", gin.WrapF(http.DefaultServeMux.ServeHTTP))
			pprofGroup.GET("/trace", gin.WrapF(http.DefaultServeMux.ServeHTTP))
			pprofGroup.GET("/allocs", gin.WrapF(http.DefaultServeMux.ServeHTTP))
			pprofGroup.GET("/block", gin.WrapF(http.DefaultServeMux.ServeHTTP))
			pprofGroup.GET("/goroutine", gin.WrapF(http.DefaultServeMux.ServeHTTP))
			pprofGroup.GET("/heap", gin.WrapF(http.DefaultServeMux.ServeHTTP))
			pprofGroup.GET("/mutex", gin.WrapF(http.DefaultServeMux.ServeHTTP))
			pprofGroup.GET("/threadcreate", gin.WrapF(http.DefaultServeMux.ServeHTTP))
		}
	}
}

// GCStats provides detailed garbage collection statistics
type GCStats struct {
	// GC frequency and timing
	NumGC      uint32        `json:"num_gc"`      // Number of GC cycles completed
	PauseTotal time.Duration `json:"pause_total"` // Total pause time
	PauseNs    []uint64      `json:"pause_ns"`    // Recent pause times (nanoseconds)
	LastGC     time.Time     `json:"last_gc"`     // Time of last GC
	NextGC     uint64        `json:"next_gc"`     // Target heap size for next GC

	// Memory statistics
	Alloc      uint64 `json:"alloc"`       // Bytes allocated and in use
	TotalAlloc uint64 `json:"total_alloc"` // Total bytes allocated (cumulative)
	Sys        uint64 `json:"sys"`         // Bytes obtained from system
	Lookups    uint64 `json:"lookups"`     // Number of pointer lookups
	Mallocs    uint64 `json:"mallocs"`     // Number of allocations
	Frees      uint64 `json:"frees"`       // Number of frees

	// Heap statistics
	HeapAlloc    uint64 `json:"heap_alloc"`    // Bytes allocated and in use
	HeapSys      uint64 `json:"heap_sys"`      // Bytes obtained from system
	HeapIdle     uint64 `json:"heap_idle"`     // Bytes in idle spans
	HeapInuse    uint64 `json:"heap_inuse"`    // Bytes in non-idle spans
	HeapReleased uint64 `json:"heap_released"` // Bytes released to the OS
	HeapObjects  uint64 `json:"heap_objects"`  // Number of allocated objects

	// GC performance metrics
	GCCPUFraction float64 `json:"gc_cpu_fraction"` // Fraction of CPU time used by GC
	GOGC          string  `json:"gogc"`            // Current GOGC setting
	GOMemLimit    string  `json:"gomemlimit"`      // Current GOMEMLIMIT setting
}

// GetGCStats returns detailed garbage collection statistics
func GetGCStats() GCStats {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	var gcStats debug.GCStats
	debug.ReadGCStats(&gcStats)

	// Convert pause durations to uint64 nanoseconds
	pauseNs := make([]uint64, len(gcStats.Pause))
	for i, pause := range gcStats.Pause {
		pauseNs[i] = uint64(pause.Nanoseconds())
	}

	return GCStats{
		// GC timing
		NumGC:      m.NumGC,
		PauseTotal: gcStats.PauseTotal,
		PauseNs:    pauseNs,
		LastGC:     time.Unix(0, int64(m.LastGC)),
		NextGC:     m.NextGC,

		// Memory stats
		Alloc:      m.Alloc,
		TotalAlloc: m.TotalAlloc,
		Sys:        m.Sys,
		Lookups:    m.Lookups,
		Mallocs:    m.Mallocs,
		Frees:      m.Frees,

		// Heap stats
		HeapAlloc:    m.HeapAlloc,
		HeapSys:      m.HeapSys,
		HeapIdle:     m.HeapIdle,
		HeapInuse:    m.HeapInuse,
		HeapReleased: m.HeapReleased,
		HeapObjects:  m.HeapObjects,

		// Performance metrics
		GCCPUFraction: m.GCCPUFraction,
		GOGC:          os.Getenv("GOGC"),
		GOMemLimit:    os.Getenv("GOMEMLIMIT"),
	}
}

// GCTuningRecommendations provides GC tuning suggestions based on current metrics
type GCTuningRecommendations struct {
	CurrentSettings map[string]string `json:"current_settings"`
	Issues          []string          `json:"issues"`
	Recommendations []string          `json:"recommendations"`
	OptimalSettings map[string]string `json:"optimal_settings"`
}

// AnalyzeGCPerformance analyzes current GC performance and provides tuning recommendations
func AnalyzeGCPerformance() GCTuningRecommendations {
	stats := GetGCStats()

	recommendations := GCTuningRecommendations{
		CurrentSettings: map[string]string{
			"GOGC":         stats.GOGC,
			"GOMEMLIMIT":   stats.GOMemLimit,
			"NumGoroutine": string(rune(runtime.NumGoroutine())),
		},
		Issues:          []string{},
		Recommendations: []string{},
		OptimalSettings: map[string]string{},
	}

	// Analyze GC frequency
	avgPause := float64(stats.PauseTotal) / float64(stats.NumGC) / 1e6 // Convert to milliseconds

	if avgPause > 10 {
		recommendations.Issues = append(recommendations.Issues, "High GC pause times detected (>10ms average)")
		recommendations.Recommendations = append(recommendations.Recommendations, "Consider increasing GOGC to reduce GC frequency")
		recommendations.OptimalSettings["GOGC"] = "200"
	}

	if stats.GCCPUFraction > 0.05 {
		recommendations.Issues = append(recommendations.Issues, "High GC CPU usage (>5%)")
		recommendations.Recommendations = append(recommendations.Recommendations, "Consider tuning memory allocation patterns")
	}

	// Analyze heap utilization
	heapUtilization := float64(stats.HeapInuse) / float64(stats.HeapSys)
	if heapUtilization < 0.5 {
		recommendations.Issues = append(recommendations.Issues, "Low heap utilization (<50%)")
		recommendations.Recommendations = append(recommendations.Recommendations, "Consider reducing GOGC for more frequent cleanup")
		recommendations.OptimalSettings["GOGC"] = "50"
	}

	// Memory limit recommendations
	if stats.GOMemLimit == "" {
		recommendations.Recommendations = append(recommendations.Recommendations, "Set GOMEMLIMIT to prevent OOM situations")
		recommendations.OptimalSettings["GOMEMLIMIT"] = "2GiB"
	}

	return recommendations
}

// SetupGCTuningEndpoints adds GC monitoring and tuning endpoints
func SetupGCTuningEndpoints(router *gin.Engine) {
	gcGroup := router.Group("/debug/gc")
	{
		// GC statistics endpoint
		gcGroup.GET("/stats", func(c *gin.Context) {
			stats := GetGCStats()
			c.JSON(http.StatusOK, gin.H{
				"gc_stats":  stats,
				"timestamp": time.Now(),
			})
		})

		// GC tuning analysis endpoint
		gcGroup.GET("/analyze", func(c *gin.Context) {
			analysis := AnalyzeGCPerformance()
			c.JSON(http.StatusOK, gin.H{
				"analysis":  analysis,
				"timestamp": time.Now(),
			})
		})

		// Force GC endpoint (for testing)
		gcGroup.POST("/force", func(c *gin.Context) {
			before := runtime.NumGoroutine()
			start := time.Now()

			runtime.GC()

			after := runtime.NumGoroutine()
			duration := time.Since(start)

			c.JSON(http.StatusOK, gin.H{
				"message":           "Garbage collection forced",
				"goroutines_before": before,
				"goroutines_after":  after,
				"gc_duration_ms":    duration.Milliseconds(),
				"timestamp":         time.Now(),
			})
		})

		// Runtime tuning endpoint
		gcGroup.POST("/tune", func(c *gin.Context) {
			var tuning struct {
				GOGC     *int    `json:"gogc"`
				MaxProcs *int    `json:"max_procs"`
				MemLimit *string `json:"mem_limit"`
			}

			if err := c.ShouldBindJSON(&tuning); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

			changes := []string{}

			if tuning.GOGC != nil {
				oldGOGC := runtime.GOMAXPROCS(0)
				if *tuning.GOGC >= 10 && *tuning.GOGC <= 500 {
					if err := os.Setenv("GOGC", fmt.Sprintf("%d", *tuning.GOGC)); err != nil {
						fmt.Printf("Warning: Failed to set GOGC environment variable: %v\n", err)
					} else {
						changes = append(changes, fmt.Sprintf("GOGC: %d", *tuning.GOGC))
					}
				}
				_ = oldGOGC // Prevent unused variable error
			}

			if tuning.MaxProcs != nil && *tuning.MaxProcs > 0 {
				oldMaxProcs := runtime.GOMAXPROCS(*tuning.MaxProcs)
				changes = append(changes, fmt.Sprintf("GOMAXPROCS: %d -> %d", oldMaxProcs, *tuning.MaxProcs))
			}

			if tuning.MemLimit != nil && *tuning.MemLimit != "" {
				if err := os.Setenv("GOMEMLIMIT", *tuning.MemLimit); err != nil {
					log.Printf("Warning: Failed to set GOMEMLIMIT: %v", err)
				}
				changes = append(changes, fmt.Sprintf("GOMEMLIMIT: %s", *tuning.MemLimit))
			}

			c.JSON(http.StatusOK, gin.H{
				"message":   "Runtime settings updated",
				"changes":   changes,
				"timestamp": time.Now(),
			})
		})
	}
}
