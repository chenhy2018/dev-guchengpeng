1. 搭建GO编程环境
2. clone qiniu base库到本地(https://github.com/qbox/base)
3. source base/env.sh base/env-mock.sh
4. 安装dep curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh
5. make 在bin/目录下会生成一个main的可执行文件
6. ./bin/main 可以运行这个文件
7. 浏览器打开指定的url就可以测试了

dep的日常使用参考 https://golang.github.io/dep/

添加新包之后记得更新push Gopkg.lock, 这个可以保证大家使用的外部版本都是一样的
