package safety

import "testing"

func TestIsForbidden(t *testing.T) {
	for _, p := range []string{
		"/System/Library/foo",
		"/usr/bin/x",
		"/bin/ls",
		"/sbin/ping",
		"/private/var/db/blah",
		"/System",
	} {
		if !IsForbidden(p) {
			t.Errorf("IsForbidden(%q) = false, want true", p)
		}
	}
	for _, p := range []string{
		"/Applications/Slack.app",
		"/Users/me/Library/Caches/com.example.app",
		"/Library/LaunchAgents/com.example.x.plist",
		"/Users/me/Library",
	} {
		if IsForbidden(p) {
			t.Errorf("IsForbidden(%q) = true, want false", p)
		}
	}
}

func TestIsAppleBundleID(t *testing.T) {
	for _, bid := range []string{"com.apple", "com.apple.Safari", "com.apple.foundation"} {
		if !IsAppleBundleID(bid) {
			t.Errorf("IsAppleBundleID(%q) = false, want true", bid)
		}
	}
	for _, bid := range []string{"com.example.app", "org.foo.bar", ""} {
		if IsAppleBundleID(bid) {
			t.Errorf("IsAppleBundleID(%q) = true, want false", bid)
		}
	}
}
