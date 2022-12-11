# env 

env 包是为了程识别当前处于我们生产流程中哪个环境设计的(开发环境 -> 测试环境 -> 灰度环境 -> 线上环境)

## 用法

```go
d:=Environment("develop")
env.Set(d) // 设置当前环境为开发环境

if env.IsDevelop(){ // true
    print("现在是开发环境")
}

if env.IsGray(){ // false
    print("现在是灰度环境")
}

print(env.Get()) // 输出 develop


```