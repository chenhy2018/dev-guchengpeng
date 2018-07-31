package account

import (
	"time"

	"qbox.us/admin_api/account.v2"
)

func convertFromAccInfo(info account.Info) *Info {
	vendors := make([]*Vendor, 0, len(info.Vendors))
	for _, v := range info.Vendors {
		vendors = append(vendors, &Vendor{
			Vendor:      v.Vendor,
			VendorId:    v.VendorId,
			VendorEmail: v.VendorEmail,
			CreatedAt:   v.CreatedAt,
		})
	}

	finalInfo := &Info{
		UserInfo: UserInfo{
			Uid:                   info.Uid,
			Username:              info.Username,
			Email:                 info.Email,
			UserType:              UserType(info.Utype),
			ParentUid:             info.ParentUid,
			IsActivated:           info.Activated,
			IsDisabled:            info.IsDisabled(),
			LastParentOperationAt: time.Time{}, // TODO admin account api should return this value
		},

		DisabledType:     DisabledType(info.DisabledType),
		DisabledReason:   info.DisabledReason,
		Vendors:          vendors,
		ChildEmailDomain: info.ChildEmailDomain,
		CanGetChildKey:   info.CanGetChildKey,
		CreatedAt:        info.CreatedAt.Time(),
		UpdatedAt:        info.UpdatedAt.Time(),
		LastLoginAt:      info.LastLoginAt.Time(),
		DisabledAt:       info.DisabledAt,
	}

	return finalInfo
}
