package proxy

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

func getEnvAny(names ...string) string {
	for _, n := range names {
		if v := os.Getenv(n); v != "" {
			return v
		}
	}
	return ""
}

// windowsProxy reads system proxy from Windows registry.
func windowsProxy() string {
	if runtime.GOOS != "windows" {
		return ""
	}

	// Check if proxy is enabled.
	enable, err := exec.Command("reg", "query",
		`HKCU\Software\Microsoft\Windows\CurrentVersion\Internet Settings`,
		"/v", "ProxyEnable").Output()
	if err != nil || !strings.Contains(string(enable), "0x1") {
		return ""
	}

	// Read proxy server address.
	server, err := exec.Command("reg", "query",
		`HKCU\Software\Microsoft\Windows\CurrentVersion\Internet Settings`,
		"/v", "ProxyServer").CombinedOutput()
	if err != nil {
		return ""
	}

	for _, line := range strings.Split(string(server), "\n") {
		line = strings.TrimSpace(line)
		// Match lines like: "ProxyServer    REG_SZ    http://127.0.0.1:7890"
		if strings.Contains(line, "REG_SZ") || strings.Contains(line, "REG_EXPAND_SZ") {
			fields := strings.Fields(line)
			if len(fields) >= 3 {
				return fields[len(fields)-1]
			}
		}
	}

	return ""
}

// ProxyFromEnvironment returns a proxy function that respects:
// - Windows: system proxy from registry (Internet Options)
// - Linux/macOS: ALL_PROXY / HTTP_PROXY / HTTPS_PROXY env vars
func ProxyFromEnvironment() func(*http.Request) (*url.URL, error) {
	var proxyURL string

	if runtime.GOOS == "windows" {
		proxyURL = windowsProxy()
		if proxyURL == "" {
			fmt.Fprintf(os.Stderr, "[proxy] no system proxy found in Windows settings\n")
		}
	}

	if proxyURL == "" {
		proxyURL = getEnvAny(
			"ALL_PROXY", "all_proxy",
			"HTTP_PROXY", "http_proxy",
			"HTTPS_PROXY", "https_proxy",
		)
	}

	if proxyURL != "" {
		if !strings.HasPrefix(proxyURL, "http://") &&
			!strings.HasPrefix(proxyURL, "https://") &&
			!strings.HasPrefix(proxyURL, "socks5://") {
			proxyURL = "http://" + proxyURL
		}

		parsed, err := url.Parse(proxyURL)
		if err == nil {
			fmt.Fprintf(os.Stderr, "[proxy] using %s\n", proxyURL)
			return http.ProxyURL(parsed)
		}
	}

	// Report what was checked (visible at -v / verbose).
	fmt.Fprintf(os.Stderr, "[proxy] no proxy found (checked env vars and Windows registry)\n")
	return http.ProxyFromEnvironment
}
