package redirect

import "strings"

// OS represents a device operating system.
type OS int

const (
	OSDesktop OS = iota
	OSIOS
	OSAndroid
)

// String returns the OS name.
func (o OS) String() string {
	switch o {
	case OSIOS:
		return "ios"
	case OSAndroid:
		return "android"
	default:
		return "desktop"
	}
}

// DeviceInfo contains parsed device information from the User-Agent.
type DeviceInfo struct {
	OS           OS
	InAppBrowser string // non-empty if request is from an in-app browser (e.g. "Instagram", "TikTok")
}

// IsMobile reports whether the device is iOS or Android.
func (d DeviceInfo) IsMobile() bool {
	return d.OS == OSIOS || d.OS == OSAndroid
}

// IsInApp reports whether the request is from an in-app browser.
func (d DeviceInfo) IsInApp() bool {
	return d.InAppBrowser != ""
}

// inAppBrowsers maps UA substrings to in-app browser names.
// Order matters: more specific strings first.
var inAppBrowsers = []struct {
	substr string
	name   string
}{
	{"Instagram", "Instagram"},
	{"FBAN", "Facebook"},
	{"FBAV", "Facebook"},
	{"FB_IAB", "Facebook"},
	{"musical_ly", "TikTok"},
	{"BytedanceWebview", "TikTok"},
	{"TikTok", "TikTok"},
	{"Snapchat", "Snapchat"},
	{"Twitter", "Twitter"},
	{"LinkedInApp", "LinkedIn"},
	{"PinterestBot", "Pinterest"},
	{"Pinterest", "Pinterest"},
	{"Line/", "LINE"},
	{"WeChat", "WeChat"},
	{"MicroMessenger", "WeChat"},
	{"Telegram", "Telegram"},
	{"Reddit", "Reddit"},
	{"Slack", "Slack"},
	{"Discord", "Discord"},
}

// DetectDevice parses a User-Agent string into DeviceInfo.
func DetectDevice(ua string) DeviceInfo {
	var d DeviceInfo

	switch {
	case strings.Contains(ua, "iPhone") ||
		strings.Contains(ua, "iPad") ||
		strings.Contains(ua, "iPod"):
		d.OS = OSIOS
	case strings.Contains(ua, "Android"):
		d.OS = OSAndroid
	default:
		d.OS = OSDesktop
	}

	for _, b := range inAppBrowsers {
		if strings.Contains(ua, b.substr) {
			d.InAppBrowser = b.name
			break
		}
	}

	return d
}
