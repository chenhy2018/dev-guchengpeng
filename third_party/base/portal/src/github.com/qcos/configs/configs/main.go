package main

import (
	"flag"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"

	"github.com/qiniu/http/bsonrpc.v1"
	"github.com/qiniu/http/webroute.v1"

	"github.com/teapots/render"
	. "github.com/teapots/teapot"

	"github.com/qcos/configs"

	"qbox.us/biz/component/helpers"
)

var signalCh = make(chan os.Signal, 5)

func init() {
	signal.Notify(signalCh, syscall.SIGHUP, syscall.SIGTERM)
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	tea := New()

	var (
		app      string
		confPath string
		files    helpers.SliceValue
	)
	flagSet := flag.NewFlagSet("app", flag.ExitOnError)
	flagSet.StringVar(&confPath, "conf", "", "")
	flagSet.StringVar(&app, "app", "app", "")
	flagSet.Var(&files, "f", "")
	flagSet.Parse(os.Args[1:])

	helpers.LoadClassicEnv(tea, app, &configs.Env, confPath, files...)

	configs.Env.Teapot = tea

	ConfigRepo(tea)
	ConfigProviders(tea)
	ConfigFilters(tea)
	ConfigRoutes(tea)

	tea.Run()
}

func ConfigRepo(tea *Teapot) {
	configs.Env.RepoPath, _ = filepath.Abs(configs.Env.RepoPath)
	tea.Logger().Info("RepoPath:", configs.Env.RepoPath)

	err := configs.Repo.Init(tea.Logger())
	if err != nil {
		tea.Logger().Error("Repo.Init failed", err)
		os.Exit(2)
	}
}

func ConfigProviders(tea *Teapot) {
	helpers.LoadClassicProviders(tea)

	tea.Provide(render.Renderer())
}

func ConfigFilters(tea *Teapot) {
	helpers.LoadClassicFilters(tea)
}

func ConfigRoutes(tea *Teapot) {
	brt := webroute.Router{Factory: bsonrpc.Factory}
	getb := &configs.GetB{}

	tea.Routers(
		Router("v1",
			Filter(configs.GracefuleWait(signalCh, 570, syscall.SIGTERM)),
			Filter(configs.StubTokenRequired),

			Router("/config/batch",
				Post(&configs.Service{}).Action("BatchConfig"),
			),
		),

		Router("/ping", Get(func(rw http.ResponseWriter) {
			rw.WriteHeader(200)
			rw.Write([]byte("pong"))
		})),

		Router("/getb", Any(brt.Register(getb).ServeHTTP)),
	)
}
