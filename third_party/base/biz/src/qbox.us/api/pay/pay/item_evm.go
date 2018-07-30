package pay

const (
	PRODUCT_EVM Product = "evm"
)

const (
	GROUP_EVM_COMPUTE  Group = "evm_compute"
	GROUP_EVM_NETWORK  Group = "evm_network"
	GROUP_EVM_VOLUME   Group = "evm_volume"
	GROUP_EVM_SNAPSHOT Group = "evm_snapshot"
	GROUP_EVM_LISTENER Group = "evm_listener"
)

const (
	// 按需

	EVM_COMPUTE_COMPUTE_2   Item = "evm:compute:compute2"
	EVM_COMPUTE_COMPUTE_4   Item = "evm:compute:compute4"
	EVM_COMPUTE_COMPUTE_8   Item = "evm:compute:compute8"
	EVM_COMPUTE_COMPUTE_12  Item = "evm:compute:compute12"
	EVM_COMPUTE_MEMORY_1    Item = "evm:compute:memory1"
	EVM_COMPUTE_MEMORY_2    Item = "evm:compute:memory2"
	EVM_COMPUTE_MEMORY_4    Item = "evm:compute:memory4"
	EVM_COMPUTE_MEMORY_8    Item = "evm:compute:memory8"
	EVM_COMPUTE_MEMORY_12   Item = "evm:compute:memory12"
	EVM_COMPUTE_MEMORY_16   Item = "evm:compute:memory16"
	EVM_COMPUTE_STD_1       Item = "evm:compute:std1"
	EVM_COMPUTE_STD_2       Item = "evm:compute:std2"
	EVM_COMPUTE_STD_4       Item = "evm:compute:std4"
	EVM_COMPUTE_STD_8       Item = "evm:compute:std8"
	EVM_COMPUTE_STD_12      Item = "evm:compute:std12"
	EVM_COMPUTE_STD_16      Item = "evm:compute:std16"
	EVM_COMPUTE_MICRO_1     Item = "evm:compute:micro1"
	EVM_COMPUTE_MICRO_0     Item = "evm:compute:micro0"
	EVM_NETWORK_BANDWIDTH_1 Item = "evm:network:bandwidth1"
	EVM_NETWORK_BANDWIDTH_2 Item = "evm:network:bandwidth2"
	EVM_NETWORK_BANDWIDTH_3 Item = "evm:network:bandwidth3"
	EVM_NETWORK_BANDWIDTH_4 Item = "evm:network:bandwidth4"
	EVM_NETWORK_BANDWIDTH_5 Item = "evm:network:bandwidth5"
	EVM_NETWORK_BANDWIDTH_6 Item = "evm:network:bandwidth6"
	EVM_VOLUME_QUICK        Item = "evm:volume:quick"
	EVM_VOLUME_CAPACIOUS    Item = "evm:volume:capacious"
	EVM_SNAPSHOT_VOLUME     Item = "evm:snapshot:volume"
	EVM_LISTENER_5K         Item = "evm:listener:5k"
	EVM_LISTENER_10K        Item = "evm:listener:10k"
	EVM_LISTENER_20K        Item = "evm:listener:20k"
	EVM_LISTENER_40K        Item = "evm:listener:40k"

	EVM_COMPUTE_12C12G_40GB Item = "evm:compute:12c12g_40gb"
	EVM_COMPUTE_12C24G_40GB Item = "evm:compute:12c24g_40gb"
	EVM_COMPUTE_12C48G_40GB Item = "evm:compute:12c48g_40gb"
	EVM_COMPUTE_16C32G_40GB Item = "evm:compute:16c32g_40gb"
	EVM_COMPUTE_16C64G_40GB Item = "evm:compute:16c64g_40gb"
	EVM_COMPUTE_1C1G_40GB   Item = "evm:compute:1c1g_40gb"
	EVM_COMPUTE_1C2G_40GB   Item = "evm:compute:1c2g_40gb"
	EVM_COMPUTE_1C4G_40GB   Item = "evm:compute:1c4g_40gb"
	EVM_COMPUTE_24C48G_20GB Item = "evm:compute:24c48g_20gb"
	EVM_COMPUTE_24C48G_40GB Item = "evm:compute:24c48g_40gb"
	EVM_COMPUTE_2C2G_40GB   Item = "evm:compute:2c2g_40gb"
	EVM_COMPUTE_2C4G_40GB   Item = "evm:compute:2c4g_40gb"
	EVM_COMPUTE_2C8G_40GB   Item = "evm:compute:2c8g_40gb"
	EVM_COMPUTE_32C64G_20GB Item = "evm:compute:32c64g_20gb"
	EVM_COMPUTE_32C64G_40GB Item = "evm:compute:32c64g_40gb"
	EVM_COMPUTE_4C16G_40GB  Item = "evm:compute:4c16g_40gb"
	EVM_COMPUTE_4C4G_40GB   Item = "evm:compute:4c4g_40gb"
	EVM_COMPUTE_4C8G_40GB   Item = "evm:compute:4c8g_40gb"
	EVM_COMPUTE_8C16G_40GB  Item = "evm:compute:8c16g_40gb"
	EVM_COMPUTE_8C32G_40GB  Item = "evm:compute:8c32g_40gb"
	EVM_COMPUTE_8C8G_40GB   Item = "evm:compute:8c8g_40gb"

	// 包月
	EVM_COMPUTE_COMPUTE_2_MONTH   Item = "evm:compute:compute2:month"
	EVM_COMPUTE_COMPUTE_4_MONTH   Item = "evm:compute:compute4:month"
	EVM_COMPUTE_COMPUTE_8_MONTH   Item = "evm:compute:compute8:month"
	EVM_COMPUTE_COMPUTE_12_MONTH  Item = "evm:compute:compute12:month"
	EVM_COMPUTE_MEMORY_1_MONTH    Item = "evm:compute:memory1:month"
	EVM_COMPUTE_MEMORY_2_MONTH    Item = "evm:compute:memory2:month"
	EVM_COMPUTE_MEMORY_4_MONTH    Item = "evm:compute:memory4:month"
	EVM_COMPUTE_MEMORY_8_MONTH    Item = "evm:compute:memory8:month"
	EVM_COMPUTE_MEMORY_12_MONTH   Item = "evm:compute:memory12:month"
	EVM_COMPUTE_MEMORY_16_MONTH   Item = "evm:compute:memory16:month"
	EVM_COMPUTE_STD_1_MONTH       Item = "evm:compute:std1:month"
	EVM_COMPUTE_STD_2_MONTH       Item = "evm:compute:std2:month"
	EVM_COMPUTE_STD_4_MONTH       Item = "evm:compute:std4:month"
	EVM_COMPUTE_STD_8_MONTH       Item = "evm:compute:std8:month"
	EVM_COMPUTE_STD_12_MONTH      Item = "evm:compute:std12:month"
	EVM_COMPUTE_STD_16_MONTH      Item = "evm:compute:std16:month"
	EVM_COMPUTE_MICRO_1_MONTH     Item = "evm:compute:micro1:month"
	EVM_COMPUTE_MICRO_0_MONTH     Item = "evm:compute:micro0:month"
	EVM_NETWORK_BANDWIDTH_1_MONTH Item = "evm:network:bandwidth1:month"
	EVM_NETWORK_BANDWIDTH_2_MONTH Item = "evm:network:bandwidth2:month"
	EVM_NETWORK_BANDWIDTH_3_MONTH Item = "evm:network:bandwidth3:month"
	EVM_NETWORK_BANDWIDTH_4_MONTH Item = "evm:network:bandwidth4:month"
	EVM_NETWORK_BANDWIDTH_5_MONTH Item = "evm:network:bandwidth5:month"
	EVM_NETWORK_BANDWIDTH_6_MONTH Item = "evm:network:bandwidth6:month"
	EVM_VOLUME_QUICK_MONTH        Item = "evm:volume:quick:month"
	EVM_VOLUME_CAPACIOUS_MONTH    Item = "evm:volume:capacious:month"
	EVM_SNAPSHOT_VOLUME_MONTH     Item = "evm:snapshot:volume:month"

	EVM_COMPUTE_12C12G_40GB_MONTH Item = "evm:compute:12c12g_40gb:month"
	EVM_COMPUTE_12C24G_40GB_MONTH Item = "evm:compute:12c24g_40gb:month"
	EVM_COMPUTE_12C48G_40GB_MONTH Item = "evm:compute:12c48g_40gb:month"
	EVM_COMPUTE_16C32G_40GB_MONTH Item = "evm:compute:16c32g_40gb:month"
	EVM_COMPUTE_16C64G_40GB_MONTH Item = "evm:compute:16c64g_40gb:month"
	EVM_COMPUTE_1C1G_40GB_MONTH   Item = "evm:compute:1c1g_40gb:month"
	EVM_COMPUTE_1C2G_40GB_MONTH   Item = "evm:compute:1c2g_40gb:month"
	EVM_COMPUTE_1C4G_40GB_MONTH   Item = "evm:compute:1c4g_40gb:month"
	EVM_COMPUTE_24C48G_20GB_MONTH Item = "evm:compute:24c48g_20gb:month"
	EVM_COMPUTE_24C48G_40GB_MONTH Item = "evm:compute:24c48g_40gb:month"
	EVM_COMPUTE_2C2G_40GB_MONTH   Item = "evm:compute:2c2g_40gb:month"
	EVM_COMPUTE_2C4G_40GB_MONTH   Item = "evm:compute:2c4g_40gb:month"
	EVM_COMPUTE_2C8G_40GB_MONTH   Item = "evm:compute:2c8g_40gb:month"
	EVM_COMPUTE_32C64G_20GB_MONTH Item = "evm:compute:32c64g_20gb:month"
	EVM_COMPUTE_32C64G_40GB_MONTH Item = "evm:compute:32c64g_40gb:month"
	EVM_COMPUTE_4C16G_40GB_MONTH  Item = "evm:compute:4c16g_40gb:month"
	EVM_COMPUTE_4C4G_40GB_MONTH   Item = "evm:compute:4c4g_40gb:month"
	EVM_COMPUTE_4C8G_40GB_MONTH   Item = "evm:compute:4c8g_40gb:month"
	EVM_COMPUTE_8C16G_40GB_MONTH  Item = "evm:compute:8c16g_40gb:month"
	EVM_COMPUTE_8C32G_40GB_MONTH  Item = "evm:compute:8c32g_40gb:month"
	EVM_COMPUTE_8C8G_40GB_MONTH   Item = "evm:compute:8c8g_40gb:month"

	// 包年
	EVM_COMPUTE_COMPUTE_2_YEAR   Item = "evm:compute:compute2:year"
	EVM_COMPUTE_COMPUTE_4_YEAR   Item = "evm:compute:compute4:year"
	EVM_COMPUTE_COMPUTE_8_YEAR   Item = "evm:compute:compute8:year"
	EVM_COMPUTE_COMPUTE_12_YEAR  Item = "evm:compute:compute12:year"
	EVM_COMPUTE_MEMORY_1_YEAR    Item = "evm:compute:memory1:year"
	EVM_COMPUTE_MEMORY_2_YEAR    Item = "evm:compute:memory2:year"
	EVM_COMPUTE_MEMORY_4_YEAR    Item = "evm:compute:memory4:year"
	EVM_COMPUTE_MEMORY_8_YEAR    Item = "evm:compute:memory8:year"
	EVM_COMPUTE_MEMORY_12_YEAR   Item = "evm:compute:memory12:year"
	EVM_COMPUTE_MEMORY_16_YEAR   Item = "evm:compute:memory16:year"
	EVM_COMPUTE_STD_1_YEAR       Item = "evm:compute:std1:year"
	EVM_COMPUTE_STD_2_YEAR       Item = "evm:compute:std2:year"
	EVM_COMPUTE_STD_4_YEAR       Item = "evm:compute:std4:year"
	EVM_COMPUTE_STD_8_YEAR       Item = "evm:compute:std8:year"
	EVM_COMPUTE_STD_12_YEAR      Item = "evm:compute:std12:year"
	EVM_COMPUTE_STD_16_YEAR      Item = "evm:compute:std16:year"
	EVM_COMPUTE_MICRO_1_YEAR     Item = "evm:compute:micro1:year"
	EVM_COMPUTE_MICRO_0_YEAR     Item = "evm:compute:micro0:year"
	EVM_NETWORK_BANDWIDTH_1_YEAR Item = "evm:network:bandwidth1:year"
	EVM_NETWORK_BANDWIDTH_2_YEAR Item = "evm:network:bandwidth2:year"
	EVM_NETWORK_BANDWIDTH_3_YEAR Item = "evm:network:bandwidth3:year"
	EVM_NETWORK_BANDWIDTH_4_YEAR Item = "evm:network:bandwidth4:year"
	EVM_NETWORK_BANDWIDTH_5_YEAR Item = "evm:network:bandwidth5:year"
	EVM_NETWORK_BANDWIDTH_6_YEAR Item = "evm:network:bandwidth6:year"
	EVM_VOLUME_QUICK_YEAR        Item = "evm:volume:quick:year"
	EVM_VOLUME_CAPACIOUS_YEAR    Item = "evm:volume:capacious:year"
	EVM_SNAPSHOT_VOLUME_YEAR     Item = "evm:snapshot:volume:year"

	EVM_COMPUTE_12C12G_40GB_YEAR Item = "evm:compute:12c12g_40gb:year"
	EVM_COMPUTE_12C24G_40GB_YEAR Item = "evm:compute:12c24g_40gb:year"
	EVM_COMPUTE_12C48G_40GB_YEAR Item = "evm:compute:12c48g_40gb:year"
	EVM_COMPUTE_16C32G_40GB_YEAR Item = "evm:compute:16c32g_40gb:year"
	EVM_COMPUTE_16C64G_40GB_YEAR Item = "evm:compute:16c64g_40gb:year"
	EVM_COMPUTE_1C1G_40GB_YEAR   Item = "evm:compute:1c1g_40gb:year"
	EVM_COMPUTE_1C2G_40GB_YEAR   Item = "evm:compute:1c2g_40gb:year"
	EVM_COMPUTE_1C4G_40GB_YEAR   Item = "evm:compute:1c4g_40gb:year"
	EVM_COMPUTE_24C48G_20GB_YEAR Item = "evm:compute:24c48g_20gb:year"
	EVM_COMPUTE_24C48G_40GB_YEAR Item = "evm:compute:24c48g_40gb:year"
	EVM_COMPUTE_2C2G_40GB_YEAR   Item = "evm:compute:2c2g_40gb:year"
	EVM_COMPUTE_2C4G_40GB_YEAR   Item = "evm:compute:2c4g_40gb:year"
	EVM_COMPUTE_2C8G_40GB_YEAR   Item = "evm:compute:2c8g_40gb:year"
	EVM_COMPUTE_32C64G_20GB_YEAR Item = "evm:compute:32c64g_20gb:year"
	EVM_COMPUTE_32C64G_40GB_YEAR Item = "evm:compute:32c64g_40gb:year"
	EVM_COMPUTE_4C16G_40GB_YEAR  Item = "evm:compute:4c16g_40gb:year"
	EVM_COMPUTE_4C4G_40GB_YEAR   Item = "evm:compute:4c4g_40gb:year"
	EVM_COMPUTE_4C8G_40GB_YEAR   Item = "evm:compute:4c8g_40gb:year"
	EVM_COMPUTE_8C16G_40GB_YEAR  Item = "evm:compute:8c16g_40gb:year"
	EVM_COMPUTE_8C32G_40GB_YEAR  Item = "evm:compute:8c32g_40gb:year"
	EVM_COMPUTE_8C8G_40GB_YEAR   Item = "evm:compute:8c8g_40gb:year"
)
