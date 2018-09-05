Mocking Network
===========

## 规格

```go
import (
	"net"
	"sync"
)

var Mocking bool 		// 网络是否处于mocking环境下
var MockingIPs []string	// 在mocking环境下，本Node的IP列表

type SpeedItem struct {
	Duration int `json:"duration"` // Duration in millisecond, 0 表示永久
	Bps      int `json:"Bps"`      // Bytes per second
}

type Speed []SpeedItem // 在len(Speed)为0的情况下表示两个节点不连通

// MockingIPs (i.e. localIPs) => "0.0.0.1" "0.0.0.2"
//  {
//    "0.0.0.3": [ # 访问0.0.0.3的时候
//      [ # 首个IP(0.0.0.1)
//        {
//          "Bps": 1024 # 速度为1K/s
//        },
//        {
//          "duration": 1000, # 1s后速度变成2K/s
//          "Bps": 2048
//        }
//      ],
//      [ # 第二个IP(0.0.0.2)
//        ...
//      ]
//    ],
//    "0.0.0.4": [ # 访问0.0.0.4的时候
//      ...
//    ]
//  }
var speeds map[string][]Speed

var DialTCP = net.DialTCP // 不推荐使用该函数，建议用 (*net.Dialer).Dial
var Dial    = net.Dial
var Listen  = net.Listen

// 逻辑地址 logicIP:logicPort => 物理地址 phyIP:phyPort 算法：
//	1) logicIP 必须是 "0.0.0.xxx"，且 xxx 在 1..255 之间
//	2) logicPort 必须在 1..255 之间
//	3) phyIP = "127.0.0.1"
//	4) phyPort = xxx * 256 + logicPort
//	5) logicIP = "0.0.0.0" 时，需要将 logicIP 改为 MockingIPs 后进行地址转换（需要Listen多个端口）
//

// 负责: 1) raddr 的地址转换; 2) laddr 的地址转换; 3) 限速
func MockDialTCP(net string, laddr, raddr *TCPAddr) (*net.TCPConn, error) {}

type Dialer net.Dialer

func (p *Dialer) Dial(network, address string) (net.Conn, error) {
	if !Mocking {
		return ((*net.Dialer)(p)).Dail(network, address)
	}
	// 1) remote address 的地址转换; 2) LocalAddr 的地址转换; 3) 限速
}

// 负责: 1) remote address的地址转换; 2) 限速
//
func MockDial(network, address string) (net.Conn, error) {}

// 负责：laddr(logicIP:logicPort)地址转换
// 当 logicIP = "0.0.0.0" 时，需要将 logicIP 改为 MockingIPs 后进行地址转换（需要Listen多个端口）
//
func MockListen(net, laddr string) (net.Listener, error) {}

type InitConfig struct {
	MockingIPs []string
	Speeds     map[string][]Speed
}

// 初始化整个Mocking Network系统
//
func Init(cfg *InitConfig) {}
```
