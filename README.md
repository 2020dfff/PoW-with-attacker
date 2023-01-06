# PoW with Attacker

> 费扬 519021910917
>
> 2023.01.04

## 1 Architecture

```go
--/src
├── main.go  // Main Function, Tree Structure
├── attack.go  // Miner, Attacker, Monitor
└── stack.go  // Stack Data Structure

0 directories, 3 files
```

**Language: Golang**

## 2 Main.go

```Go
const (
   MinersNumber   = 5
   AttackerNumber = 2
   MonitorNumber  = 1
   MaxChannelSize = 1000
   Interval       = 1000 // ms
   IntervalNum    = 5
)

type block struct {
   Lasthash  string //上一个区块的Hash
   Hash      string //本区块Hash
   Data      string //区块存储的数据
   Timestamp int64  //时间戳
   Height    uint   //区块高度
   MinerId   uint64 //矿工 ID
   DiffNum   uint   //难度值
   Nonce     int64  //随机数
}
```

区块的结构如上，同时定义了攻击和挖矿过程中的常量

```Go
func (b *Block) serialize() []byte  // 序列化
func (b *Block) getHash() // 计算hash
func (b *Block) IsValid() bool  // 验证

func NewBlockChain(genesisBlock *Block) *BlockChain  // 初始化
func (bc *BlockChain) append(b *Block) bool  // 添加新块
func (bc *BlockChain) search(b *Block) *BlockChainNode  // 根据 hash 查找区块 
func (bc *BlockChain) print()  // 打印最长链
```

## 3 Attack.go

```Go
//矿工
type Node struct {
	id              uint64
	blockChain      BlockChain
	blockChainMutex sync.RWMutex
	peers           map[uint64]chan Block
	receiveChan     chan Block
	update          bool
	flagMutex       sync.RWMutex
}
//攻击者
type Attacker struct {
	id                uint64  // 攻击者对外的 uid
	secret_id         uint64  // 攻击者内的 uid
	blockChain        BlockChain
	blockChainMutex   sync.RWMutex
	peers             map[uint64]chan Block
	receiveChan       chan Block
	update            bool
	flagMutex         sync.RWMutex
	cahoot            map[uint64]chan ConspiratorialTarget  // 攻击者秘密信道
	secretReceiveChan chan ConspiratorialTarget  // 攻击者秘密接收信道
	request           bool  // 目标父区块更新标志位
	requestMutex      sync.RWMutex
	target            *BlockChainNode  // 目标父区块
	targetMutex       sync.RWMutex  // 目标父区块读写锁
	votes             uint64
}
//监视者
type Monitor struct {
	id          uint64
	blockChain  BlockChain
	peers       map[uint64]chan Block
	receiveChan chan Block
}
```

- 矿工需要异步接受其他节点的新区块，让自己尽可能地在最长的上挖掘，还需要循环不断的挖矿；

- 监听员的任务是获取系统中完整的区块链，进行输出和统计信息；
- 攻击者选择当前最长链的倒数第二个区块，或在区块链最大高度不大于正在解决的区块的高度加 1 时仍将原目标区块作为目标区块，实行分叉攻击；

## 4 Test

![img](https://raw.githubusercontent.com/2020dfff/md_fig/master/data_win11/goland.png?token=GHSAT0AAAAAABV5CJYF2NUCMWKMQYGWHA6OY5YH4EQ)

无论节点是否恶意都会遵守工作量证明机制。当系统区域稳定后，时间的变化幅度在 10% 以内，结果符合预期。