package helpers

import (
	"log"
	"os"

	"github.com/teapots/request-logger"
	"github.com/teapots/teapot"

	"qbox.us/biz/component/filters"
)

func LoadClassicFilters(tea *teapot.Teapot) {
	logOut := log.New(os.Stderr, "", log.LstdFlags|log.Lmicroseconds)

	loggerOption := reqlogger.LoggerOption{
		ColorMode:     !tea.Config.RunMode.IsProd(),
		LineInfo:      true,
		ShortLine:     tea.Config.RunMode.IsProd(),
		FlatLine:      tea.Config.RunMode.IsProd(),
		LogStackLevel: teapot.LevelCritical,
	}

	tea.Filter(
		// 所有过滤器之前抓取 panic
		teapot.RecoveryFilter(),
	)

	if tea.Config.RunMode.IsDev() || tea.Config.RunMode.IsTest() {
		// 因为 chrome  的限制
		// 开发模式下删除 Postman Header 请求前缀
		tea.Filter(
			filters.HeaderRemovePrefixFilter("Postman-"),
		)
	}

	tea.Filter(
		// 在静态文件之后加入，跳过静态文件请求
		reqlogger.ReqLoggerFilter(logOut, loggerOption),

		// 在 action 里直接返回一般请求结果
		teapot.GenericOutFilter(),
	)
}
