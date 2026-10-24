package launcher

import (
	"os"
	"os/exec"
	"path/filepath"
)

// OpenUI shows the interface. mode "app" opens a chromeless app window in
// an installed Chromium-based browser (Chrome, Edge, Chromium, Brave,
// Vivaldi) with its own profile, falling back to the default browser;
// "browser" always uses the default browser; "none" does nothing.
func OpenUI(mode, url, profileDir string) error {
	switch mode {
	case "none":
		return nil
	case "browser":
		return openDefault(url)
	}
	if exe := findAppBrowser(); exe != "" {
		cmd := exec.Command(exe,
			"--app="+url,
			"--user-data-dir="+profileDir,
			"--no-first-run",
			"--no-default-browser-check",
			"--window-size=1440,900",
			"--class=Cryptomon",
		)
		if err := cmd.Start(); err == nil {
			go func() { _ = cmd.Wait() }()
			return nil
		}
	}
	return openDefault(url)
}

func openDefault(url string) error {
	var cmd *exec.Cmd
	switch goos() {
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	if err := cmd.Start(); err != nil {
		return err
	}
	go func() { _ = cmd.Wait() }()
	return nil
}

func findAppBrowser() string {
	switch goos() {
	case "windows":
		var candidates []string
		for _, env := range []string{"ProgramFiles", "ProgramFiles(x86)", "LocalAppData"} {
			base := os.Getenv(env)
			if base == "" {
				continue
			}
			candidates = append(candidates,
				filepath.Join(base, `Google\Chrome\Application\chrome.exe`),
				filepath.Join(base, `Microsoft\Edge\Application\msedge.exe`),
				filepath.Join(base, `BraveSoftware\Brave-Browser\Application\brave.exe`),
				filepath.Join(base, `Chromium\Application\chrome.exe`),
				filepath.Join(base, `Vivaldi\Application\vivaldi.exe`),
			)
		}
		return firstExisting(candidates)
	case "darwin":
		return firstExisting([]string{
			"/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
			"/Applications/Chromium.app/Contents/MacOS/Chromium",
			"/Applications/Microsoft Edge.app/Contents/MacOS/Microsoft Edge",
			"/Applications/Brave Browser.app/Contents/MacOS/Brave Browser",
			"/Applications/Vivaldi.app/Contents/MacOS/Vivaldi",
		})
	default:
		for _, name := range []string{"chromium", "chromium-browser", "google-chrome", "google-chrome-stable",
			"brave", "brave-browser", "microsoft-edge", "microsoft-edge-stable", "vivaldi", "vivaldi-stable"} {
			if p, err := exec.LookPath(name); err == nil {
				return p
			}
		}
	}
	return ""
}

func firstExisting(paths []string) string {
	for _, p := range paths {
		if st, err := os.Stat(p); err == nil && !st.IsDir() {
			return p
		}
	}
	return ""
}
