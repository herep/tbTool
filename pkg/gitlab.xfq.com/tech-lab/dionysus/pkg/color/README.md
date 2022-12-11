# color

color 包是为了在开发过程中 std.out/std.err 输出更加容易识别而设计的

## 使用方法

1. 直接在 print 里使用
```go
fmt.Println(string(color.Red), "这行是红色的", string(color.White)) // 这里要设置一下恢复白色，要不下一行也是红色
```

2. 包装 std.out

```go
var w io.Writer

w = color.Writer(os.Stdout, color.Yellow)
w.Write([]byte("这行是黄色的 \n"))
```