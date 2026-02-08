package server

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/naufaruuu/redis-rdb-analyzer/decoder"
	"github.com/schollz/progressbar/v3"
)

type JobState string

const (
	StateChecking    JobState = "checking"
	StateDownloading JobState = "downloading"
	StateParsing     JobState = "parsing"
	StateDone        JobState = "done"
	StateError       JobState = "error"
)

type Job struct {
	ID        string   `json:"id"`
	Status    string   `json:"status"`
	State     JobState `json:"state"`
	Error     string   `json:"error,omitempty"`
	Instance  string   `json:"instance,omitempty"`
	Progress  float64  `json:"progress,omitempty"`
	StartTime time.Time
}

type JobManager struct {
	jobs sync.Map
}

var GlobalJobManager = &JobManager{}

func (jm *JobManager) StartJob(namespace, pod, path string) string {
	// Generate ID: namespace_redis_pod_name_2026-0205_01
	id := GetNextID(namespace, pod)
	
	job := &Job{
		ID:        id,
		Status:    "Initializing...",
		State:     StateChecking,
		StartTime: time.Now(),
	}
	jm.jobs.Store(id, job)

	go jm.runJob(job, namespace, pod, path)

	return id
}

func (jm *JobManager) GetStatus(id string) *Job {
	if v, ok := jm.jobs.Load(id); ok {
		return v.(*Job)
	}
	return nil
}

func (jm *JobManager) runJob(job *Job, namespace, pod, path string) {
	// ... (helper update function) ...
	update := func(state JobState, status string, errStr string) {
		job.State = state
		job.Status = status
		job.Error = errStr
        
        // Log to stdout for visibility
        if state == StateError {
            log.Printf("[Job %s] ERROR: %s - %s", job.ID, status, errStr)
        } else {
            log.Printf("[Job %s] %s: %s", job.ID, state, status)
        }
	}

	// 1. Check Size
	update(StateChecking, "Checking RDB size...", "")
	// kubectl exec -n <ns> <pod> -- stat -c %s <path>
	cmdCheck := exec.Command("kubectl", "exec", "-n", namespace, pod, "--", "stat", "-c", "%s", path)
	out, err := cmdCheck.Output() // Use Output() to get only stdout
	if err != nil {
		errMsg := err.Error()
		if exitErr, ok := err.(*exec.ExitError); ok {
			errMsg = fmt.Sprintf("%v, stderr: %s", err, string(exitErr.Stderr))
		}
		update(StateError, "Check failed", fmt.Sprintf("Failed to check size: %s", errMsg))
		return
	}
	sizeStr := strings.TrimSpace(string(out))
	size, err := strconv.ParseInt(sizeStr, 10, 64)
	if err != nil {
		update(StateError, "Check failed", fmt.Sprintf("Invalid size output: %q", sizeStr))
		return
	}

	// Check against max size from env var
	maxSize := GetMaxRDBSize()
	if size > maxSize {
		update(StateError, "File too large", fmt.Sprintf("File size %s exceeds limit of %s", FormatSize(size), FormatSize(maxSize)))
		return
	}

	// 2. Download
	update(StateDownloading, fmt.Sprintf("Copying RDB from Redis %s...", formatBytes(size)), "")
    
    // Use local tmp directory to avoid filling up RAM (tmpfs)
    tmpDir, _ := filepath.Abs("./tmp")
    os.MkdirAll(tmpDir, 0755)
    
	localPath := filepath.Join(tmpDir, fmt.Sprintf("rdr_%s.rdb", job.ID))
	defer os.Remove(localPath) // Cleanup

	cmdCp := exec.Command("kubectl", "cp", fmt.Sprintf("%s/%s:%s", namespace, pod, path), localPath)
	if out, err := cmdCp.CombinedOutput(); err != nil {
		update(StateError, "Download failed", fmt.Sprintf("Copy failed: %v, output: %s", err, string(out)))
		return
	}

	// 3. Parse
	update(StateParsing, "Parsing RDB file...", "")
	
	f, err := os.Open(localPath)
	if err != nil {
		update(StateError, "Open failed", fmt.Sprintf("Failed to open file: %v", err))
		return
	}
	defer f.Close()

	decoder := decoder.NewDecoder()
	go func() {
		// Get file size for progress bar
		fInfo, err := f.Stat()
		if err != nil {
			log.Printf("Failed to stat file for progress bar: %v", err)
		} else {
			bar := progressbar.DefaultBytes(
				fInfo.Size(),
				"parsing rdb",
			)
			// Wrap reader
			barReader := progressbar.NewReader(f, bar)
			
			// Updates job progress
			go func() {
				// Poll progress bar state and update job
				for {
					if bar.IsFinished() {
						return
					}
					current := bar.State().CurrentBytes
					total := bar.State().Max
					if total > 0 {
						job.Progress = float64(current) / float64(total) * 100
					}
					time.Sleep(500 * time.Millisecond)
				}
			}()
			
			decoder.DecodeWithHDT(&barReader)
			bar.Finish()
			job.Progress = 100 // Ensure 100% on finish
			fmt.Println()
			return
		}
		
		// Fallback if stat failed
		decoder.DecodeWithHDT(f)
	}()

	counter := NewCounter()
	counter.Count(decoder.Entries)

	// Store result
	instanceName := job.ID // Use ID as instance name
	
	// Save to DB and Memory
	counters.Set(instanceName, counter)
	err = SaveAnalysis(instanceName, namespace, pod, path, counter)
	if err != nil {
		update(StateError, "Save failed", fmt.Sprintf("Failed to save result: %v", err))
		return
	}

	update(StateDone, "Analysis Complete", "")
}

func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
