package idomain

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/qiniu/ctype"
	"qbox.us/api/one/domain"
)

/*
https://pm.qbox.me/attachments/download/2457/七牛DNS&域名规划.pdf

domain: <index>.<itbl>.<channel>.<zone>.glb.<qiniucdn|clouddn>.com
cname:  <itbl>.<channel.vip>.<channel.name>.<zone>.glb.qiniudns.com
*/
type Idomain struct {
	// TODO: IDomainSuffix这个列表未来会很长
	// .<channel>.<zone>.glb.<qiniucdn|clouddn>.com
	// .<channel>.glb.<qiniucdn|clouddn>.com
	IDomainSuffix []string
}

func NewIdomain(idomainSuffix []string) Idomain {

	return Idomain{IDomainSuffix: idomainSuffix}
}

func (self Idomain) Split(domain string) (itbl uint32, channel, region string, global, ok bool) {
	for _, suffix := range self.IDomainSuffix {
		if strings.HasSuffix(domain, suffix) {
			prefix := domain[:len(domain)-len(suffix)]
			lastDotPos := strings.LastIndex(prefix, ".")
			itbl36Str := prefix[lastDotPos+1:]
			itbl64, err := strconv.ParseUint(itbl36Str, 36, 64)
			if err != nil {
				return
			}
			suffixBlock := strings.Split(suffix, ".")
			itbl, channel, ok = uint32(itbl64), suffixBlock[1], true
			if suffixBlock[2] == "glb" {
				global = true
				return
			}
			region = suffixBlock[2]
			return
		}
	}
	return
}

func (self Idomain) Domains(itbl uint32, channel, zone string) (domains []string) {

	itblStr := strconv.FormatUint(uint64(itbl), 36)
	for _, suffix := range self.IDomainSuffix {
		if strings.HasPrefix(suffix, "."+channel+"."+zone) {
			domains = append(domains, itblStr+suffix)
		}
	}
	return
}

func (self Idomain) DomainsAndCnames(itbl uint32, channel, zone string) (domains, cnames []string) {

	itblStr := strconv.FormatUint(uint64(itbl), 36)
	for _, suffix := range self.IDomainSuffix {
		if strings.HasPrefix(suffix, "."+channel+"."+zone) {
			domains = append(domains, itblStr+suffix)
			cnameSuffix := cnameSuffix(channel, zone, false)
			cnames = append(cnames, itblStr+cnameSuffix)
		}
	}
	return
}

func (self Idomain) DomainsAll(itbl uint32, channel string) (domains []domain.Entry) {

	itblStr := strconv.FormatUint(uint64(itbl), 36)
	for _, suffix := range self.IDomainSuffix {
		suffixBlock := strings.Split(suffix, ".")
		if suffixBlock[1] == channel {
			global := len(suffixBlock) == 5
			domains = append(domains, domain.Entry{Domain: itblStr + suffix, Global: global})
		}
	}
	return
}

func (self Idomain) DomainsAndCnamesAll(itbl uint32, channel string) (domains []domain.Entry, cnames []string) {

	itblStr := strconv.FormatUint(uint64(itbl), 36)
	for _, suffix := range self.IDomainSuffix {
		suffixBlock := strings.Split(suffix, ".")
		if suffixBlock[1] == channel {
			global := len(suffixBlock) == 5
			zone := ""
			if !global {
				zone = suffixBlock[2]
			}
			domains = append(domains, domain.Entry{Domain: itblStr + suffix, Global: global})

			cnameSuffix := cnameSuffix(channel, zone, global)
			cnames = append(cnames, itblStr+cnameSuffix)
		}
	}
	return
}

func cnameSuffix(channel, zone string, global bool) (cnameSuffix string) {
	channelName, channelVip := channel[:len(channel)-1], string(channel[len(channel)-1])
	channelVip = "v" + channelVip
	if channelVip == "v0" {
		channelVip = "src"
	}
	if global {
		cnameSuffix = fmt.Sprintf(".%v.%v.glb.qiniudns.com.", channelVip, channelName)
	} else {
		cnameSuffix = fmt.Sprintf(".%v.%v.%v.glb.qiniudns.com.", channelVip, channelName, zone)
	}
	return
}

// z[0-9]+
func isValidRegion(region string) bool {
	if !strings.HasPrefix(region, "z") {
		return false
	}
	return ctype.IsType(ctype.DIGIT, region[1:])
}

func Domain2Cname(domain string) (cname string) {
	domainBlock := strings.Split(domain, ".")
	N := len(domainBlock)
	if isValidRegion(domainBlock[N-4]) { // not global
		itblStr, channel, zone := domainBlock[N-6], domainBlock[N-5], domainBlock[N-4]
		cname = itblStr + cnameSuffix(channel, zone, false)
	} else { // global
		itblStr, channel := domainBlock[N-5], domainBlock[N-4]
		cname = itblStr + cnameSuffix(channel, "", true)
	}
	return
}
