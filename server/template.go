// Copyright 2017 XUEQIU.COM
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package server

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/naufaruuu/redis-rdb-analyzer/views"
	"github.com/dustin/go-humanize"
)

var (
	tmpl          *template.Template
	tplCommonData = map[string]interface{}{}
	tplFuncMap    = make(template.FuncMap)
	isFirst       = true
)

func InitHTMLTmpl(includeEmbedded bool, scanDirs []string) {
	// init function maps
	tplFuncMap["isFirst"] = func() bool { res := isFirst; isFirst = false; return res }
	tplFuncMap["clearFirst"] = func() bool { isFirst = true; return isFirst }
	tplFuncMap["hash"] = func(str string) string { return fmt.Sprintf("%x", md5.Sum([]byte(str))) }
	tplFuncMap["humanizeBytes"] = humanize.Bytes
	tplFuncMap["humanizeComma"] = func(i uint64) string { return humanize.Comma(int64(i)) }
    tplFuncMap["timestampFormat"] = func(expiry int64) string {
		if expiry > 0 {
			return time.Unix(0, expiry*int64(time.Millisecond)).Format("2006-01-02 15:04:05")
		}else{
			return ""
		}
	}

	// Gather all unique template names from both assets and local file system
	// We map the template name (e.g. "home.html") to the actual file path that should be loaded
	// This allows "views/legacy/home.html" to be registered as "home.html".
	templatePaths := make(map[string]string)
	
	// 1. From embedded assets
	if includeEmbedded {
		for _, name := range views.AssetNames() {
			if strings.HasSuffix(name, ".html") {
				// Embedded assets are just filenames usually
				templatePaths[name] = "" // Empty path means use Asset()
			}
		}
	}
	
	// 2. From local views directories
	for _, dir := range scanDirs {
		entries, err := os.ReadDir(dir)
		if err == nil {
			for _, entry := range entries {
				if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".html") {
					// Map "home.html" -> "views/legacy/home.html"
					// This effectively overrides embedded asset if name matches
					templatePaths[entry.Name()] = dir + "/" + entry.Name()
				}
			}
		} else {
			// It is okay if directory doesn't exist (e.g. if running in prod without local views)
			// But for debugging print it
			// fmt.Printf("DEBUG: Could not scan %s: %v\n", dir, err)
		}
	}

	// init views html template
	for name, path := range templatePaths {
		var content []byte
		var tmplErr error
		
		loadedFromDisk := false
		
		// If we have a local path mapped, try to load it
		if path != "" {
			if data, err := os.ReadFile(path); err == nil {
				content = data
				loadedFromDisk = true
				fmt.Printf("DEBUG: Loaded template from disk: %s\n", path)
			}
		} else {
			// Fallback: If no explicit path, and we allow legacy handling,
			// check for file in "views/" + name etc.
			// However, if we are in strict mode (scanDirs provided), we shouldn't randomly scan other places
			// UNLESS we want to maintain the old behavior of "override embedded if local exists".
			// Since scanDirs covers the new locations, this fallback is mostly for "development in root views/".
			// Let's keep it but respect scanDirs context if possible?
			// Actually, let's keep it simple: if includeEmbedded is true, we check local overrides.
			
			if includeEmbedded {
				localPaths := []string{
					"views/" + name,
					"../views/" + name,
					name,
				}
				for _, p := range localPaths {
					if data, err := os.ReadFile(p); err == nil {
						content = data
						loadedFromDisk = true
						fmt.Printf("DEBUG: Loaded template from disk: %s\n", p)
						break
					}
				}
			}
		}

		// Priority 2: Use embedded assets (only if not loaded from disk and exists in assets)
		// And only if includeEmbedded is true
		if !loadedFromDisk && includeEmbedded {
			// Check if it exists in assets
			if _, err := views.AssetInfo(name); err == nil {
				content, tmplErr = views.Asset(name)
				if tmplErr != nil {
					log.Printf("|ERROR|asset %v err %v", name, tmplErr)
					continue
				}
			} else {
				// Should have been loaded from disk if it was in our map from Scan
				log.Printf("|ERROR| Template %s found in scan but failed to load from disk and not in assets", name)
				continue
			}
		}
		
		// If neither loaded from disk nor embedded (and embedded allowed), skip
		if len(content) == 0 {
			continue
		}

		if tmpl == nil {
			tmpl, tmplErr = template.New(name).Funcs(tplFuncMap).Parse(string(content))
		} else {
			tmpl, tmplErr = tmpl.New(name).Funcs(tplFuncMap).Parse(string(content))
		}

		if tmplErr != nil {
			log.Printf("|ERROR|parse template err %v", tmplErr)
		}
	}
}

// ServeHTML generate and write html to client
func ServeHTML(w http.ResponseWriter, layout string, content string, data map[string]interface{}) {
	var buf bytes.Buffer
	bodyTmplErr := tmpl.ExecuteTemplate(&buf, content, data)
	if bodyTmplErr != nil {
		log.Printf("|ERROR|ServeHTML bodyTmplErr ERROR %v", bodyTmplErr)
	}
	bodyHTML := template.HTML(buf.String())
	if len(data) == 0 {
		data = map[string]interface{}{}
	}
	data["LayoutContent"] = bodyHTML
	tmplErr := tmpl.ExecuteTemplate(w,
		layout,
		data)

	if tmplErr != nil {
		log.Printf("|ERROR|ServeHTML LayoutTmplErr ERROR %v", tmplErr)
	}
}
