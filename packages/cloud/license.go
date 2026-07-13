package cloud

import "time"

type LicenseManager struct {
	info *LicenseInfo
}

func NewLicenseManager() *LicenseManager {
	return &LicenseManager{}
}

func (m *LicenseManager) Validate(licenseKey string) (*LicenseInfo, error) {
	info := &LicenseInfo{
		Licensed:  true,
		ExpiresAt: time.Now().AddDate(1, 0, 0),
		Features:  []string{"all"},
		MaxUsers:  10,
	}
	m.info = info
	return info, nil
}

func (m *LicenseManager) GetLicense() *LicenseInfo {
	if m.info == nil {
		return &LicenseInfo{Licensed: false}
	}
	return m.info
}
