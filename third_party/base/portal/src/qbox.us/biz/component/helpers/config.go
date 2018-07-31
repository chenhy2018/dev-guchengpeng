package helpers

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/teapots/config"
	"github.com/teapots/teapot"
	"qbox.us/biz/utils.v2/log"
)

type SliceValue []string

func (s *SliceValue) Set(v string) error {
	for _, p := range strings.Split(v, ",") {
		p = strings.TrimSpace(p)
		if p != "" {
			*s = append(*s, v)
		}
	}
	return nil
}

func (s *SliceValue) String() string {
	return strings.Join(*s, ",")
}

func (s *SliceValue) Unique() {
	v := make(SliceValue, 0, len(*s))
	m := make(map[string]bool, len(*s))
	for _, p := range *s {
		if m[p] {
			continue
		}
		m[p] = true
		v = append(v, p)
	}
	*s = v
}

func LoadClassicEnv(tea *teapot.Teapot, app string, env interface{}, dir string, addFiles ...string) {
	defer func() {
		// 非请求的 logger 使用全局的
		UseGlobalLogger(tea)
	}()

	UseLongLogLevelTag()

	LoadConfigFiles(tea, app, env, dir, addFiles...)
}

func LoadConfigFiles(tea *teapot.Teapot, app string, env interface{}, dir string, addFiles ...string) (conf config.Configer) {
	directory, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	wd, _ := os.Getwd()
	wd, _ = filepath.Abs(wd)

	directories := []string{directory, wd}
	if dir != "" {
		dir, _ = filepath.Abs(dir)
		directories = append(directories, dir)
	}

	files := []string{}
	for _, d := range directories {
		files = append(files, []string{
			filepath.Join(d, app+".ini"),
			filepath.Join(d, app+".dev.ini"),
			filepath.Join(d, app+".test.ini"),
			filepath.Join(d, app+".prod.ini"),
		}...)
	}
	skipNum := len(files)
	files = append(files, addFiles...)

	maps := make(map[string]bool, len(files))
	var last config.Configer
	for n, path := range files {
		if n < skipNum && maps[path] {
			continue
		}
		maps[path] = true
		conf, err := config.LoadIniFile(path)
		if err != nil {
			if !os.IsNotExist(err) || n >= skipNum {
				tea.Logger().Errorf("%s load err: %v", path, err)
			}
		} else {
			if last != nil {
				conf.SetParent(last)
			}
			last = conf
			tea.Logger().Infof("%s load success", path)
		}
	}

	if last != nil {
		conf = last
		tea.ImportConfig(last)
		config.Decode(tea.Config, env)
	}
	return
}

func UseGlobalLogger(tea *teapot.Teapot) {
	log.X.SetFlatLine(tea.Config.RunMode.IsProd())
	log.X.SetColorMode(!tea.Config.RunMode.IsProd())
	log.X.SetShortLine(tea.Config.RunMode.IsProd())
	tea.SetLogger(log.X)
}

func UseLongLogLevelTag() {
	teapot.DefaultLevelTag = [9]string{
		"[EMERGENCY]",
		"[ALERT]",
		"[CRITCIAL]",
		"[ERROR]",
		"[WARN]",
		"[NOTICE]",
		"[INFO]",
		"[DEBUG]",
	}
}
