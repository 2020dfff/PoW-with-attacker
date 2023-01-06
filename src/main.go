package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"math/big"
	"math/rand"
	"time"
)

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

//// 用于存放区块,以便连接成区块链
//var blockchain []block
//// 初始块挖矿难度值(在此修改即可）
//var diffNum uint = 17

// 区块挖矿（通过自身递增nonce值计算hash）
//func mine(data string) block {
//	if len(blockchain) < 1 {
//		log.Panic("还未生成创世区块！")
//	}
//	lastBlock := blockchain[len(blockchain)-1]
//	//制造一个新的区块
//	newBlock := new(block)
//	newBlock.Lasthash = lastBlock.Hash
//	newBlock.Timestamp = time.Now().UnixMilli()
//	newBlock.Height = lastBlock.Height + 1
//	newBlock.DiffNum = diffNum
//	newBlock.Data = data
//	var nonce int64 = 0
//	//根据挖矿难度值计算的一个大数
//	newBigint := big.NewInt(1)
//	newBigint.Lsh(newBigint, 256-diffNum) //相当于左移 1<<256-diffNum
//	for {
//		newBlock.Nonce = nonce
//		newBlock.getHash()
//		hashInt := big.Int{}
//		hashBytes, _ := hex.DecodeString(newBlock.Hash)
//		hashInt.SetBytes(hashBytes) //把本区块hash值转换为一串数字
//		//如果hash小于挖矿难度值计算的一个大数，则代表挖矿成功
//		if hashInt.Cmp(newBigint) == -1 {
//			break
//		} else {
//			nonce++ //不满足条件，则不断递增随机数，直到本区块的散列值小于指定的大数
//		}
//	}
//	return *newBlock
//}

func (b *block) serialize() []byte {
	bytes, err := json.Marshal(b)
	if err != nil {
		log.Panic(err)
	}
	return bytes
}

func (b *block) getHash() {
	result := sha256.Sum256(b.serialize())
	b.Hash = hex.EncodeToString(result[:])
}

func (b *block) IsValid() bool {
	target := big.NewInt(1)
	target.Lsh(target, 256-b.DiffNum)
	hashBytes, _ := hex.DecodeString(b.Hash)
	hashInt := big.Int{}
	hashInt.SetBytes(hashBytes)
	if hashInt.Cmp(target) == -1 {
		return true
	} else {
		return false
	}
}

// Tree: Data Structure For Blockchain Storage

type BlockChainNode struct {
	block  *block
	childs []*BlockChainNode
	parent *BlockChainNode
}
type BlockChain struct {
	root      *BlockChainNode
	workspace *BlockChainNode
	maxHeight uint
	index     map[uint][]*BlockChainNode
}

// relevant functions

func NewBlockChain(genesisBlock *block) *BlockChain {
	root := new(BlockChainNode)
	root.block = genesisBlock
	root.childs = nil
	root.parent = nil
	newBlockChain := new(BlockChain)
	newBlockChain.root = root
	newBlockChain.workspace = root
	newBlockChain.maxHeight = 1
	newBlockChain.index = make(map[uint][]*BlockChainNode)
	newBlockChain.index[1] = append(newBlockChain.index[1], newBlockChain.root)
	return newBlockChain
}

func (bc *BlockChain) append(b *block) bool {
	if !b.IsValid() {
		fmt.Print("APPEND NOT VALID!\n")
		return false
	} // 判断有效性
	if bc.index[b.Height-1] == nil {
		fmt.Printf("b.Height = %d, NIL!\n", b.Height)
		return false
	} else {
		for i := 0; i < len(bc.index[b.Height]); i++ {
			if b.Hash == bc.index[b.Height][i].block.Hash {
				fmt.Print("SAME\n")
				return false
			}
		}
		for i := 0; i < len(bc.index[b.Height-1]); i++ {
			if b.Lasthash == bc.index[b.Height-1][i].block.Hash {
				newBCN := new(BlockChainNode)
				newBCN.block = b
				newBCN.childs = nil
				newBCN.parent = bc.index[b.Height-1][i]
				bc.index[b.Height-1][i].childs = append(bc.index[b.Height-1][i].childs, newBCN)
				bc.index[b.Height] = append(bc.index[b.Height], newBCN)
				if b.Height > bc.maxHeight {
					// switch to the longest chain
					bc.maxHeight = b.Height
					bc.workspace = newBCN
					return true
				}
				return false
			}
		}
		return false
	}
}

func (bc *BlockChain) statistics(n *BlockChainNode) int64 {
	if n.block.Height < IntervalNum+1 {
		return IntervalNum * Interval
	}
	cur := n.block.Timestamp
	lstPtr := n
	for i := 0; i < IntervalNum; i++ {
		lstPtr = lstPtr.parent
	}
	lst := lstPtr.block.Timestamp
	return cur - lst
}

func (bc *BlockChain) search(b *block) *BlockChainNode {
	if b == nil {
		return nil
	}
	for _, bcn := range bc.index[b.Height] {
		if bcn.block.Hash == b.Hash {
			return bcn
		}
	}
	return nil
}

func (bc *BlockChain) print() {
	tmp := []*BlockChainNode{}
	tmp = append(tmp, bc.workspace)
	for tmp[len(tmp)-1].parent != nil {
		tmp = append(tmp, tmp[len(tmp)-1].parent)
	}
	fmt.Print("*******************************\n")
	for i := len(tmp) - 1; i >= 0; i-- {
		fmt.Println(tmp[i].block)
	}
	fmt.Print("*******************************\n")
}

func (bc *BlockChain) formatPrintAll() {
	fmt.Print("*******************************\n")
	stack := NewStack()
	stack.Push(bc.root)
	for !stack.IsEmpty() {
		cbcn := stack.Pop()
		for i := uint(1); i < cbcn.block.Height; i++ {
			fmt.Print("\t")
		}
		fmt.Println("└──", cbcn.block.Data)
		for _, bcn := range cbcn.childs {
			stack.Push(bcn)
		}
	}
	fmt.Print("===============================\n")
}

func CreateGenesisBlock(data string) *block {
	//制造一个创世区块
	genesisBlock := new(block)
	genesisBlock.Timestamp = time.Now().UnixMilli()
	genesisBlock.Data = data
	genesisBlock.MinerId = math.MaxUint64
	genesisBlock.Lasthash = "0000000000000000000000000000000000000000000000000000000000000000"
	genesisBlock.Height = 1
	genesisBlock.Nonce = 0
	genesisBlock.DiffNum = 0
	newInt := big.NewInt(1)
	newInt.Lsh(newInt, 256-genesisBlock.DiffNum)
	for {
		genesisBlock.getHash()
		hashInt := big.Int{}
		hashBytes, _ := hex.DecodeString(genesisBlock.Hash)
		hashInt.SetBytes(hashBytes)
		if hashInt.Cmp(newInt) == -1 {
			break
		} else {
			genesisBlock.Nonce++
		}
	}
	return genesisBlock
}

//	genesisBlock.getHash()
//	fmt.Println(*genesisBlock)
//	//将创世区块添加进区块链
//	blockchain = append(blockchain, *genesisBlock)
//	for i := 0; i < 10; i++ {
//		newBlock := mine("天气不错" + strconv.Itoa(i))
//		blockchain = append(blockchain, newBlock)
//		fmt.Println(newBlock)
//	}
//}

func main() {
	// random seed
	rand.Seed(time.Now().Unix())

	nodes := make([]*Node, MinersNumber)
	attackers := make([]*Attacker, AttackerNumber)
	peers := make(map[uint64]chan block)
	for i := 0; i < MinersNumber+AttackerNumber+MonitorNumber; i++ {
		peers[uint64(i)] = make(chan block, MaxChannelSize)
	}
	genesisBlock := CreateGenesisBlock("IS416 HHJ")
	for i := uint64(0); i < MinersNumber; i++ {
		nodes[i] = NewNode(i, *genesisBlock, peers, peers[i])
	}
	cahoots := make(map[uint64]chan ConspiratorialTarget)
	for i := uint64(0); i < AttackerNumber; i++ {
		cahoots[uint64(i)] = make(chan ConspiratorialTarget, MaxChannelSize)
	}
	for i := uint64(0); i < AttackerNumber; i++ {
		attackers[i] = NewAttacker(i+MinersNumber, i, *genesisBlock, peers, peers[i+MinersNumber], cahoots, cahoots[i])
	}
	monitor := NewMonitor(0, *genesisBlock, peers, peers[MinersNumber+AttackerNumber])
	go monitor.Run()

	// start all nodes
	for i := 0; i < MinersNumber; i++ {
		go nodes[i].Run()
	}
	for i := 0; i < AttackerNumber; i++ {
		go attackers[i].Run()
	}

	// block to wait for all nodes' threads
	<-make(chan int)
}
