package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strconv"

	"github.com/julienschmidt/httprouter"
	"github.com/urfave/cli"
)

var counters = NewSafeMap()

func getInstances() []string {
	keys := []string{}
	for k := range counters.Items() {
		keys = append(keys, k.(string))
	}
	sort.Strings(keys)
	return keys
}

// Show parse rdbfile(s) and show statistical information by MODERN html
func Show(c *cli.Context) {
	fmt.Fprintln(c.App.Writer, "Starting Redis Analyzer Server...")

	// top N bigkey 
	topN := 100
	
	// Filter logic
	var sizeFilter int64 = 0

	// Initialize DB
	InitDB()
	LoadHistory()

	// Init templates
	InitHTMLTmpl(false, []string{"views"})
	
	// Set common data for templates
	tplCommonData["TopN"] = strconv.Itoa(topN)
	tplCommonData["sizeFilter"] = strconv.FormatInt(sizeFilter, 10)

	// start http server
	router := httprouter.New()
	
	// Serve the new UI
	router.GET("/", index)
	
	// API to return JSON data for the SPA
	router.GET("/api/analysis", apiAnalysis)
	
	// Keep existing APIs for compatibility/Jobs
	router.POST("/api/job/start", startJobHandler)
	router.GET("/api/job/status", statusJobHandler)
    router.GET("/api/discovery", discoveryHandler)

	// Get port from env var (RDR_PORT) or CLI flag or default
	port := GetPort()
	if c.IsSet("port") {
		port = int(c.Uint("port"))
	}
	portStr := strconv.Itoa(port)

	fmt.Fprintln(c.App.Writer, "Server started. Access at http://{$IP}:"+portStr)
	listenErr := http.ListenAndServe(":"+portStr, router)
	if listenErr != nil {
		fmt.Fprintf(c.App.ErrWriter, "Listen port err: %v\n", listenErr)
	}
}

func index(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	data := map[string]interface{}{}
	data["Instances"] = getInstances()
	
	// Serve layout.html which includes other components
	err := tmpl.ExecuteTemplate(w, "layout.html", data)
	if err != nil {
		fmt.Printf("Error serving layout.html: %v\n", err)
		http.Error(w, "Internal Server Error", 500)
	}
}

func apiAnalysis(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	path := r.URL.Query().Get("path")
	if path == "" {
		http.Error(w, "Missing path parameter", 400)
		return
	}

	c := counters.Get(path)
	if c == nil {
		http.Error(w, "Instance not found", 404)
		return
	}
	counter := c.(*Counter)

	// Get params from global/context
	topN := 100
	if val, ok := tplCommonData["TopN"]; ok {
		if s, ok := val.(string); ok {
			if n, err := strconv.Atoi(s); err == nil {
				topN = n
			}
		}
	}
	var sizeFilter int64 = 0
	if val, ok := tplCommonData["sizeFilter"]; ok {
		if s, ok := val.(string); ok {
			if n, err := strconv.ParseInt(s, 10, 64); err == nil {
				sizeFilter = n
			}
		}
	}

	data := map[string]interface{}{}
	data["CurrentInstance"] = path
	data["LargestKeys"] = counter.GetLargestEntries(topN, sizeFilter)
	
	// Prefixes logic
	largestKeyPrefixesByType := map[string][]*PrefixEntry{}
	for _, entry := range counter.GetLargestKeyPrefixes() {
		if entry.Bytes < 1000*1000 && len(largestKeyPrefixesByType[entry.Type]) > 50 {
			continue
		}
		largestKeyPrefixesByType[entry.Type] = append(largestKeyPrefixesByType[entry.Type], entry)
	}
	data["LargestKeyPrefixes"] = largestKeyPrefixesByType

	data["TypeBytes"] = counter.typeBytes
	data["TypeNum"] = counter.typeNum
	
	totalNum := uint64(0)
	for _, v := range counter.typeNum {
		totalNum += v
	}
	totalBytes := uint64(0)
	for _, v := range counter.typeBytes {
		totalBytes += v
	}
	data["TotalNum"] = totalNum
	data["TotalBytes"] = totalBytes

	// LenLevelCount
	lenLevelCount := map[string][]*PrefixEntry{}
	for _, entry := range counter.GetLenLevelCount() {
		lenLevelCount[entry.Type] = append(lenLevelCount[entry.Type], entry)
	}
	data["LenLevelCount"] = lenLevelCount

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func startJobHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	type Request struct {
		Namespace string `json:"namespace"`
		Pod       string `json:"pod"`
		Path      string `json:"path"`
	}
	var req Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	id := GlobalJobManager.StartJob(req.Namespace, req.Pod, req.Path)
	json.NewEncoder(w).Encode(map[string]string{"job_id": id})
}

func statusJobHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	id := r.URL.Query().Get("id")
	job := GlobalJobManager.GetStatus(id)
	if job == nil {
		http.Error(w, "Job not found", http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(job)
}

func discoveryHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
    fmt.Println("DEBUG: discoveryHandler reached")
    res, err := DiscoverRedisResources()
    if err != nil {
        fmt.Printf("DEBUG: Discovery failed: %v\n", err)
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(res)
}


