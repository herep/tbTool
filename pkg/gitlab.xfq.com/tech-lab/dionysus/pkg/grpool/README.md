# pkg说明

### grpool

在gapi框架中不建议使用go goroutine开go协程，因为如果go协程如果没有recover()的话，程序panic会无法捕获。
要求使用此pkg下的```grpool.Submit(task func())```来运行go协程，此方法自带了recover()，可以捕获panic。