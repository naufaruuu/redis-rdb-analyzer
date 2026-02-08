package server

import (
    "encoding/json"
    "fmt"
    "os/exec"
    "strings"
)

type DiscoveredRedis struct {
    Namespace string   `json:"namespace"`
    Pods      []string `json:"pods"`
}

func DiscoverRedisResources() ([]DiscoveredRedis, error) {
    // 0. Check Cache
    if cachedBytes, err := GetDiscoveryCache(); err == nil {
         fmt.Println("Serving Redis resources from Cache (SQLite)...")
         var cachedRes []DiscoveredRedis
         if err := json.Unmarshal(cachedBytes, &cachedRes); err == nil {
             return cachedRes, nil
         }
    }

    fmt.Println("Discovering Redis resources via kubectl (Optimized)...")
    
    // 1. Get all StatefulSets
    cmdSts := exec.Command("kubectl", "get", "sts", "-A", "-o", "json")
    outSts, err := cmdSts.Output()
    if err != nil {
        return nil, fmt.Errorf("failed to get sts: %v", err)
    }

    // We need a struct to parse the Label Selector
    type StsItem struct {
        Metadata struct {
            Name      string `json:"name"`
            Namespace string `json:"namespace"`
        } `json:"metadata"`
        Spec struct {
            Selector struct {
                MatchLabels map[string]string `json:"matchLabels"`
            } `json:"selector"`
        } `json:"spec"`
    }
    
    type StsList struct {
        Items []StsItem `json:"items"`
    }

    var stsList StsList
    if err := json.Unmarshal(outSts, &stsList); err != nil {
        return nil, err
    }

    results := []DiscoveredRedis{}
    namespaces := make(map[string]bool) // just for logging count
    
    // 2. Iterate each Redis STS and fetch its pods directly using Selector
    for _, sts := range stsList.Items {
        if !strings.Contains(sts.Metadata.Name, "redis") {
            continue
        }
        namespaces[sts.Metadata.Namespace] = true

        // Build Label Selector string
        var selectors []string
        for k, v := range sts.Spec.Selector.MatchLabels {
            selectors = append(selectors, fmt.Sprintf("%s=%s", k, v))
        }
        selectorStr := strings.Join(selectors, ",")
        
        if selectorStr == "" {
            continue 
        }

        // Get pods for this specific STS
        cmdPods := exec.Command("kubectl", "get", "pods", "-n", sts.Metadata.Namespace, "-l", selectorStr, "-o", "json")
        outPods, err := cmdPods.Output()
        if err != nil {
            fmt.Printf("Error getting pods for STS %s: %v\n", sts.Metadata.Name, err)
            continue
        }

        // Parse pods (reusing a simple struct)
        type PodItem struct {
            Metadata struct {
                Name string `json:"name"`
            } `json:"metadata"`
        }
        type PodList struct {
            Items []PodItem `json:"items"`
        }

        var podList PodList 
        if err := json.Unmarshal(outPods, &podList); err != nil {
            continue
        }

        pods := []string{}
        for _, pod := range podList.Items {
             pods = append(pods, pod.Metadata.Name)
        }
        fmt.Printf("STS %s/%s: Found %d pods.\n", sts.Metadata.Namespace, sts.Metadata.Name, len(pods))

        if len(pods) > 0 {
            // Check if we already have an entry for this namespace in results?
            // The existing structure is []DiscoveredRedis{ Namespace, Pods[] }.
            // Accessing/Merging is O(N).
            // Let's optimize: map keys first.
            found := false
            for i := range results {
                if results[i].Namespace == sts.Metadata.Namespace {
                    results[i].Pods = append(results[i].Pods, pods...)
                    found = true
                    break
                }
            }
            if !found {
               results = append(results, DiscoveredRedis{
                   Namespace: sts.Metadata.Namespace,
                   Pods:      pods,
               }) 
            }
        }
    }
    fmt.Printf("Total unique namespaces with Redis: %d\n", len(namespaces))

    // Save to Cache
    if jsonBytes, err := json.Marshal(results); err == nil {
        SaveDiscoveryCache(jsonBytes)
    }

    return results, nil
}
