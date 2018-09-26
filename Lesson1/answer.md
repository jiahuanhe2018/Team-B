# 第一题
zk-snark:（参照教学版）
constraints | inputsize | time |
:---|:---|:---
1000|100| 0.2s
1000|200| 0.31s
1000|300| 0.21s
1000|400| 0.24s
2000|100| 0.29s
2000|200| 0.31s
2000|300| 0.32s
2000|400| 0.35s

目前我还不清楚constraints和inputsize的作用。

ring-signature我是跑的golang版本。不知道该怎么对比。

# 第二题


名字 | 交易 | 交易数 | 交易属性 | 块大小 
:---|:---|:---|:---|:---
btc | 转账 | 以填满区块文件为上限, 最近2000左右 | inputs ouputs | 1 Mb|
eth | 转账、智能合约的调用 | 以耗尽区块的gas limit为上限 | From to value | 无固定上限，目前25k左右
monero | 转账 | 以填满区块文件为上限, 最近7笔左右 | Inputs Outputs | 最大值是60K或者最近100块大小中位数的2倍
zcash | 转账 | 以填满区块文件为上限 | Inputs Outputs | 最大2MB
eos | trace | 以填满区块文件为上限, | receipt, act  | 1MB




