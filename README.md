# go-gp
[![Go Report Card](https://goreportcard.com/badge/github.com/Ficoto/go-gp)](https://goreportcard.com/report/github.com/Ficoto/go-gp)
[![MIT license](http://img.shields.io/badge/license-MIT-9d1f14)](http://opensource.org/licenses/MIT)

golang实现的协程池

## Usage
```
go get github.com/Ficoto/go-gp
```
Create a Pool to use
```go
package main

import "github.com/Ficoto/go-gp"

func main() {
	p := gp.New()
	p.Go(func() {
		// doing something
	})
	p.Close()
}
```

## 已知问题
 - 在调用Close方法时，待提交的新任务可能会直接被抛弃，仅仅只有执行中的任务会保障被执行完

## 特性
 - 允许设置异步任务重试
 - 允许设置异步任务执行完成回调方法
 - 调用close方法会等待执行中的任务执行完且回调完才会结束
 - 并发协程到设置的最大协程数时，提交新任务会被堵塞