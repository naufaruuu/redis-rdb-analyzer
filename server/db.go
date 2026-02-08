package server

import (
    "database/sql"
    "encoding/json"
    "fmt"
    "log"
    "time"

    _ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

func InitDB() {
    var err error
    log.Println("Initializing SQLite database at ./data/rdr.db...")
    db, err = sql.Open("sqlite3", "./data/rdr.db")
    if err != nil {
        log.Fatal(err)
    }

    createTableSQL := `CREATE TABLE IF NOT EXISTS history (
        "id" TEXT PRIMARY KEY,
        "namespace" TEXT,
        "pod" TEXT,
        "path" TEXT,
        "date" TEXT,
        "number" INTEGER,
        "data" BLOB,
        "timestamp" DATETIME DEFAULT CURRENT_TIMESTAMP
    );`

    createCacheTableSQL := `CREATE TABLE IF NOT EXISTS discovery_cache (
        key TEXT PRIMARY KEY,
        data BLOB,
        updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
    );`
    
    _, err = db.Exec(createTableSQL)
    if err != nil {
        log.Fatal(err)
    }
    _, err = db.Exec(createCacheTableSQL)
    if err != nil {
        log.Fatal(err)
    }
    log.Println("Database initialized successfully.")
}

func SaveAnalysis(id string, ns, pod, path string, counter *Counter) error {
    log.Printf("Saving analysis to DB: %s (ns=%s, pod=%s, path=%s)", id, ns, pod, path)
    dto := counter.ToDTO()
    data, err := json.Marshal(dto)
    if err != nil {
        return err
    }

    stmt, err := db.Prepare("INSERT OR REPLACE INTO history(id, namespace, pod, path, data, date) values(?,?,?,?,?,?)")
    if err != nil {
        return err
    }
    defer stmt.Close()
    
    dateStr := time.Now().Format("2006-0102")

    _, err = stmt.Exec(id, ns, pod, path, data, dateStr)
    if err == nil {
        log.Println("Analysis saved to DB successfully.")
    }
    return err
}

func LoadHistory() {
    log.Println("Loading history from DB...")
    rows, err := db.Query("SELECT id, data FROM history")
    if err != nil {
        log.Println("Error loading history:", err)
        return
    }
    defer rows.Close()

    count := 0
    for rows.Next() {
        var id string
        var data []byte
        err = rows.Scan(&id, &data)
        if err != nil {
            log.Println(err)
            continue
        }

        var dto CounterDTO
        if err := json.Unmarshal(data, &dto); err != nil {
            log.Printf("Failed to unmarshal data for %s: %v", id, err)
            continue
        }

        counter := dto.ToCounter()
        counters.Set(id, counter)
        count++
    }
    log.Printf("Loaded %d analysis records from history.", count)
}

func GetNextID(ns, pod string) string {
    dateStr := time.Now().Format("2006-0102")
    // Pattern: ns_pod_date_%
    prefix := fmt.Sprintf("%s_%s_%s", ns, pod, dateStr)
    
    // We want to find max number for this prefix.
    // Querying by ID LIKE prefix + '_' + '%'
    // But checking suffix is hard in SQL.
    // We can store 'number' as separate column?
    // Let's parse ID in code or simpler: store a sequence in a separate table?
    // Or just SELECT id FROM history WHERE id LIKE ?
    
    rows, err := db.Query("SELECT id FROM history WHERE id LIKE ?", prefix+"_%")
    if err != nil {
        return prefix + "_01"
    }
    defer rows.Close()

    maxN := 0
    for rows.Next() {
        var id string
        rows.Scan(&id)
        // Extract suffix
        // id is 'prefix_NN'
        if len(id) > len(prefix)+1 {
            suffix := id[len(prefix)+1:]
            var n int
            fmt.Sscanf(suffix, "%d", &n)
            if n > maxN {
                maxN = n
            }
        }
    }

    return fmt.Sprintf("%s_%02d", prefix, maxN+1)
}

func GetDiscoveryCache() ([]byte, error) {
    var data []byte
    var updatedAt time.Time
    // check if data exists
    row := db.QueryRow("SELECT data, updated_at FROM discovery_cache WHERE key = 'k8s_resources'")
    if err := row.Scan(&data, &updatedAt); err != nil {
        return nil, err
    }
    // check TTL (from POD_CACHE_DURATION env var)
    cacheDuration := GetPodCacheDuration()
    if time.Since(updatedAt) > cacheDuration {
        return nil, fmt.Errorf("cache expired (TTL: %v)", cacheDuration)
    }
    return data, nil
}

func SaveDiscoveryCache(data []byte) error {
    stmt, err := db.Prepare("INSERT OR REPLACE INTO discovery_cache(key, data, updated_at) values('k8s_resources', ?, CURRENT_TIMESTAMP)")
    if err != nil {
         return err
    }
    defer stmt.Close()
    _, err = stmt.Exec(data)
    return err
}
