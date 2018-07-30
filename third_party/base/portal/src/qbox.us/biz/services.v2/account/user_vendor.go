package account

import (
	"time"
)

type Vendor struct {
	Vendor      string    `json:"vendor"`
	VendorId    string    `json:"vendor_id"`
	VendorEmail string    `json:"vendor_email"`
	CreatedAt   time.Time `json:"created_at"`
}

func (a *adminAccountService) UserCreateByVendor(vendor, vendorId, vendorEmail string) (info *Info, err error) {
	userInfo, err := a.service.UserCreateByVendor(vendor, vendorId, vendorEmail, a.rpcLogger)
	if err != nil {
		return
	}
	info = convertFromAccInfo(userInfo)
	return
}

func (a *adminAccountService) UserBindAccount(uid uint32, vendor, vendorId, vendorEmail string) (info *Info, err error) {
	userInfo, err := a.service.UserBindAccount(uid, vendor, vendorId, vendorEmail, a.rpcLogger)
	if err != nil {
		return
	}
	info = convertFromAccInfo(userInfo)
	return
}

func (a *adminAccountService) UserUnbindAccount(uid uint32, vendor string) (info *Info, err error) {
	userInfo, err := a.service.UserUnbindAccount(uid, vendor, a.rpcLogger)
	if err != nil {
		return
	}
	info = convertFromAccInfo(userInfo)
	return
}
