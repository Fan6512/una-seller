package publication

import (
	"fmt"
	"strings"
	"unicode"
)

type TitleMode string

const (
	PlatformAuto TitleMode = "platform_auto"
	Localized    TitleMode = "localized"
)

// LocalizedText comes from server-loaded, NFC-normalized saved listing fields.
// No translation or language detection is performed during publication.
type LocalizedText struct {
	English, Spanish, Portuguese string
}

type Copy struct {
	Mode                            TitleMode
	FamilyName, Description         string
	SiteFamilyName, SiteDescription string
}

// SelectCopy maps a confirmed user choice and saved copy to one destination.
// Missing localized titles are errors, not permission to fall back to English.
// A missing localized description means no description override, independently
// of the title choice; the caller must show this choice before submission.
func SelectCopy(site string, mode TitleMode, titles, descriptions LocalizedText) (Copy, error) {
	c := Copy{Mode: mode, FamilyName: titles.English, Description: descriptions.English}
	if !sitePattern.MatchString(site) {
		return Copy{}, fmt.Errorf("unsupported copy destination")
	}
	if mode == Localized {
		c.SiteFamilyName, c.SiteDescription = titles.Spanish, descriptions.Spanish
		if site == "MLB" {
			c.SiteFamilyName, c.SiteDescription = titles.Portuguese, descriptions.Portuguese
		}
	}
	if err := c.validate(); err != nil {
		return Copy{}, err
	}
	return c, nil
}

func titleOK(s string, max int) bool {
	return textOK(s, max) && strings.IndexFunc(s, unicode.IsControl) < 0
}

func (c Copy) validate() error {
	if !titleOK(c.FamilyName, 500) || !textOK(c.Description, 20000) {
		return fmt.Errorf("English family name and description are required")
	}
	switch c.Mode {
	case PlatformAuto:
		if c.SiteFamilyName != "" || c.SiteDescription != "" {
			return fmt.Errorf("automatic copy must not include site overrides")
		}
	case Localized:
		if !titleOK(c.SiteFamilyName, 60) {
			return fmt.Errorf("localized family name is required and must not exceed 60 characters")
		}
		if c.SiteDescription != "" && !textOK(c.SiteDescription, 20000) {
			return fmt.Errorf("invalid localized description")
		}
	default:
		return fmt.Errorf("publication title mode must be explicitly selected")
	}
	return nil
}
