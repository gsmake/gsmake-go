# gsmake2.0 测试

##安装

1. go get github.com/gsdocker/gsmake
2. git checkout release/v2.0
3. ./setup ${安装目录}
4. 添加 GSMAKE_HOME=${安装目录}
5. 将${安装目录}/bin加入PATH环境变量

##使用gsmake创建gsweb项目

1. 通过archtype创建项目模板：gsmake create -o test -v v2.0 "github.com/gsdocker/gsweb:basic"
2. 启动gsweb自动构建: cd test & gsmake gsweb gsweb.basic 
