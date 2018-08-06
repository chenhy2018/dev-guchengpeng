package idomain

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"

	"qbox.us/api/one/domain"
)

func TestCheckIdomain(t *testing.T) {

	domainSuffix := []string{
		".com0.glb.qiniucdn.com",
		".com1.glb.clouddn.com",
		".com0.glb.clouddn.com",

		".com0.z0.glb.qiniucdn.com",
		".com1.z0.glb.clouddn.com",
		".com0.z0.glb.clouddn.com",

		".com0.z1.glb.qiniucdn.com",
		".com1.z1.glb.clouddn.com",
		".com0.z1.glb.clouddn.com",
	}

	itbl := uint32(4000000)
	sitbl := strconv.FormatUint(uint64(itbl), 36)

	idomain := Idomain{domainSuffix}

	type ret struct {
		B       bool
		Channel string
		Itbl    uint32
		Region  string
		Global  bool
	}

	testCase := map[string]ret{
		"":                                            ret{B: false},
		"helloqini":                                   ret{B: false},
		"hello.qini":                                  ret{B: false},
		"hello.qini.aksjdf":                           ret{B: false},
		"_hell.com0.z0.glb.qiniucdn.com":              ret{B: false},
		".com1.z0.glb.qiniucdn.com":                   ret{B: false},
		"xxx." + sitbl + "..com0.z0.glb.qiniucdn.com": ret{B: false},
		"xxx." + sitbl + ".com0.z0.glb.qiniucdn.com":  ret{B: true, Channel: "com0", Itbl: itbl, Region: "z0", Global: false},
		sitbl + ".com0.z0.glb.qiniucdn.com":           ret{B: true, Channel: "com0", Itbl: itbl, Region: "z0", Global: false},
		"xxx." + sitbl + ".com0.glb.qiniucdn.com":     ret{B: true, Channel: "com0", Itbl: itbl, Region: "", Global: true},
		sitbl + ".com0.glb.qiniucdn.com":              ret{B: true, Channel: "com0", Itbl: itbl, Region: "", Global: true},
	}

	for k, v := range testCase {
		itbl, channel, region, global, ok := idomain.Split(k)
		assert.Equal(t, v.B, ok, "checkIDomain: %v, %+v", k, v)
		if ok {
			assert.Equal(t, v.Itbl, itbl, "checkIDomain")
			assert.Equal(t, v.Channel, channel, "checkIDomain")
			assert.Equal(t, v.Region, region, "checkIDomain")
			assert.Equal(t, v.Global, global, "checkIDomain")
		}
	}

	{
		domains, cnames := idomain.DomainsAndCnames(itbl, "com0", "z1")
		assert.Equal(t, []string{
			sitbl + ".com0.z1.glb.qiniucdn.com",
			sitbl + ".com0.z1.glb.clouddn.com",
		}, domains)
		assert.Equal(t, []string{
			sitbl + ".src.com.z1.glb.qiniudns.com.",
			sitbl + ".src.com.z1.glb.qiniudns.com.",
		}, cnames)

		domains = idomain.Domains(itbl, "com1", "z0")
		assert.Equal(t, []string{
			sitbl + ".com1.z0.glb.clouddn.com",
		}, domains)
	}

	{
		domains, cnames := idomain.DomainsAndCnamesAll(itbl, "com0")
		assert.Equal(t, []domain.Entry{
			{Domain: sitbl + ".com0.glb.qiniucdn.com", Global: true},
			{Domain: sitbl + ".com0.glb.clouddn.com", Global: true},
			{Domain: sitbl + ".com0.z0.glb.qiniucdn.com", Global: false},
			{Domain: sitbl + ".com0.z0.glb.clouddn.com", Global: false},
			{Domain: sitbl + ".com0.z1.glb.qiniucdn.com", Global: false},
			{Domain: sitbl + ".com0.z1.glb.clouddn.com", Global: false},
		}, domains)
		assert.Equal(t, []string{
			sitbl + ".src.com.glb.qiniudns.com.",
			sitbl + ".src.com.glb.qiniudns.com.",
			sitbl + ".src.com.z0.glb.qiniudns.com.",
			sitbl + ".src.com.z0.glb.qiniudns.com.",
			sitbl + ".src.com.z1.glb.qiniudns.com.",
			sitbl + ".src.com.z1.glb.qiniudns.com.",
		}, cnames)

		domains = idomain.DomainsAll(itbl, "com1")
		assert.Equal(t, []domain.Entry{
			{Domain: sitbl + ".com1.glb.clouddn.com", Global: true},
			{Domain: sitbl + ".com1.z0.glb.clouddn.com", Global: false},
			{Domain: sitbl + ".com1.z1.glb.clouddn.com", Global: false},
		}, domains)
	}
}

func TestDomain2Cname(t *testing.T) {
	{
		domain := "77fl0b.com2.z0.glb.qiniucdn.com"
		cname := Domain2Cname(domain)
		assert.Equal(t, "77fl0b.v2.com.z0.glb.qiniudns.com.", cname)
	}
	{
		domain := "77fl0b.com0.z0.glb.qiniucdn.com"
		cname := Domain2Cname(domain)
		assert.Equal(t, "77fl0b.src.com.z0.glb.qiniudns.com.", cname)
	}

	{
		domain := "77fl0b.com2.glb.qiniucdn.com"
		cname := Domain2Cname(domain)
		assert.Equal(t, "77fl0b.v2.com.glb.qiniudns.com.", cname)
	}
	{
		domain := "77fl0b.com0.glb.qiniucdn.com"
		cname := Domain2Cname(domain)
		assert.Equal(t, "77fl0b.src.com.glb.qiniudns.com.", cname)
	}
}
