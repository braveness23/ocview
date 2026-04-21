package data

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

var prRefRe = regexp.MustCompile(`\s*\(#\d+(?:,\s*#\d+)*\)`)

type ChangelogEntry struct {
	Changes []string
	Fixes   []string
}

func StripPRs(s string) string {
	return prRefRe.ReplaceAllString(s, "")
}

func ParseChangelog(text string) map[string]ChangelogEntry {
	entries := map[string]ChangelogEntry{}
	var curVer string
	var curSec string // "changes" | "fixes" | ""

	for _, line := range strings.Split(text, "\n") {
		if m := regexp.MustCompile(`^## (\d{4}\.\d+\.\d+\S*)`).FindStringSubmatch(line); m != nil {
			curVer = m[1]
			entries[curVer] = ChangelogEntry{}
			curSec = ""
			continue
		}
		if curVer == "" {
			continue
		}
		switch strings.TrimSpace(line) {
		case "### Changes":
			curSec = "changes"
			continue
		case "### Fixes":
			curSec = "fixes"
			continue
		}
		if strings.HasPrefix(line, "###") {
			curSec = ""
			continue
		}
		if strings.HasPrefix(line, "- ") && curSec != "" {
			e := entries[curVer]
			item := line[2:]
			if curSec == "changes" {
				e.Changes = append(e.Changes, item)
			} else {
				e.Fixes = append(e.Fixes, item)
			}
			entries[curVer] = e
		}
	}
	return entries
}

func compareCalver(a, b string) int {
	ap := strings.Split(a, ".")
	bp := strings.Split(b, ".")
	for i := 0; i < max(len(ap), len(bp)); i++ {
		av, bv := 0, 0
		if i < len(ap) {
			av, _ = strconv.Atoi(ap[i])
		}
		if i < len(bp) {
			bv, _ = strconv.Atoi(bp[i])
		}
		if av != bv {
			if av < bv {
				return -1
			}
			return 1
		}
	}
	return 0
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func LoadUpdates() []OcUpdateRelease {
	home, _ := os.UserHomeDir()
	npmGlobal := filepath.Join(home, ".npm-global", "lib", "node_modules", "openclaw")

	var installedVersion, latestAvailable, lastCheckedAt string

	if b, err := os.ReadFile(filepath.Join(npmGlobal, "package.json")); err == nil {
		var pkg struct {
			Version string `json:"version"`
		}
		if json.Unmarshal(b, &pkg) == nil {
			installedVersion = pkg.Version
		}
	}

	if b, err := os.ReadFile(filepath.Join(home, ".openclaw", "update-check.json")); err == nil {
		var uc struct {
			LastAvailableVersion string `json:"lastAvailableVersion"`
			LastCheckedAt        string `json:"lastCheckedAt"`
		}
		if json.Unmarshal(b, &uc) == nil {
			latestAvailable = uc.LastAvailableVersion
			lastCheckedAt = uc.LastCheckedAt
		}
	}

	type installRec struct {
		From      string `json:"from"`
		To        string `json:"to"`
		Timestamp string `json:"timestamp"`
	}
	var installHistory []installRec
	if b, err := os.ReadFile(filepath.Join(home, ".openclaw", "logs", "update-history.json")); err == nil {
		var h struct {
			Installs []installRec `json:"installs"`
		}
		if json.Unmarshal(b, &h) == nil {
			installHistory = h.Installs
		}
	}

	parsed := map[string]ChangelogEntry{}
	if b, err := os.ReadFile(filepath.Join(npmGlobal, "CHANGELOG.md")); err == nil {
		parsed = ParseChangelog(string(b))
	}

	// Collect stable versions descending
	var versions []string
	for ver := range parsed {
		if !strings.Contains(ver, "-") {
			versions = append(versions, ver)
		}
	}
	// sort descending
	for i := 0; i < len(versions); i++ {
		for j := i + 1; j < len(versions); j++ {
			if compareCalver(versions[i], versions[j]) < 0 {
				versions[i], versions[j] = versions[j], versions[i]
			}
		}
	}

	// Prepend latestAvailable if not in changelog
	if latestAvailable != "" {
		found := false
		for _, v := range versions {
			if v == latestAvailable {
				found = true
				break
			}
		}
		if !found && compareCalver(latestAvailable, installedVersion) > 0 {
			versions = append([]string{latestAvailable}, versions...)
		}
	}

	var items []OcUpdateRelease
	for _, ver := range versions {
		entry := parsed[ver]
		isInstalled := ver == installedVersion
		isLatest := ver == latestAvailable
		isAvailable := !isInstalled && compareCalver(ver, installedVersion) > 0

		var rec *InstallRecord
		for _, h := range installHistory {
			if h.To == ver {
				rec = &InstallRecord{From: h.From, To: h.To, Timestamp: h.Timestamp}
				break
			}
		}

		var changes, fixes []string
		for _, c := range entry.Changes {
			changes = append(changes, StripPRs(c))
		}
		for _, f := range entry.Fixes {
			fixes = append(fixes, StripPRs(f))
		}

		items = append(items, OcUpdateRelease{
			ID:            ver,
			Name_:         ver,
			Version:       ver,
			IsInstalled:   isInstalled,
			IsLatest:      isLatest,
			IsAvailable:   isAvailable,
			LastChecked:   lastCheckedAt,
			ChangeCount:   len(changes) + len(fixes),
			Changes:       changes,
			Fixes:         fixes,
			InstallRecord: rec,
		})
	}
	return items
}
