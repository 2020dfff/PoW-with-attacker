package main

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"math/rand"
	"strconv"
	"sync"
	"time"
)

type Node struct {
	id              uint64
	blockChain      BlockChain
	blockChainMutex sync.RWMutex // 锁
	peers           map[uint64]chan block
	receiveChan     chan block
	update          bool
	flagMutex       sync.RWMutex
}

func NewNode(id uint64, genesisBlock block, peers map[uint64]chan block, recvChan chan block) *Node {
	return &Node{
		id:          id,
		blockChain:  *NewBlockChain(&genesisBlock),
		peers:       peers,
		receiveChan: recvChan,
		update:      false,
	}
}

// Run start one consensus node, remember to start one thread for receiving messages (n.Receive())
func (n *Node) Run() {
	fmt.Println("start node : ", n.id)
	go n.Receive()

	for {
		// n.Broadcast(Block{MinerId: n.id})
		newBlock := n.Mine()
		if newBlock == nil {
			continue
		}
		n.blockChainMutex.Lock()
		n.blockChain.append(newBlock)
		n.blockChainMutex.Unlock()
		n.Broadcast(*newBlock)
	}
}

func (n *Node) Receive() {
	for {
		select {
		case msg := <-n.receiveChan:
			n.handler(msg)
		}
	}
}

func (n *Node) handler(msg block) {
	n.blockChainMutex.Lock()
	flag := n.blockChain.append(&msg)
	n.blockChainMutex.Unlock()
	if flag {
		n.flagMutex.Lock()
		n.update = true
		n.flagMutex.Unlock()
	}
}

func (n *Node) Broadcast(msg block) {
	for id, ch := range n.peers {
		if id == n.id {
			continue
		}
		ch <- msg
	}
}

func (n *Node) Mine() *block {
	// New block
	newBlock := new(block)
	newBlock.Timestamp = time.Now().UnixMilli()
	newBlock.MinerId = n.id
	var nonce int64 = rand.Int63()
	n.blockChainMutex.RLock()
	lastBlock := n.blockChain.workspace.block
	newBlock.Lasthash = lastBlock.Hash
	newBlock.Height = lastBlock.Height + 1
	newBlock.Data = "第" + strconv.Itoa(int(lastBlock.Height)) + "次打造区块链，被矿工" + strconv.FormatUint(n.id, 10) + "记录"
	// calculate difficulty
	newBlock.DiffNum = lastBlock.DiffNum
	n.blockChainMutex.RUnlock()
	if ((newBlock.Height-1)%IntervalNum == 0) && (newBlock.Height != 0) {
		n.blockChainMutex.RLock()
		interval := n.blockChain.statistics(n.blockChain.workspace)
		n.blockChainMutex.RUnlock()
		if interval < 0.9*IntervalNum*Interval {
			newBlock.DiffNum += 1
		} else if interval > 1.1*IntervalNum*Interval {
			newBlock.DiffNum -= 1
		}
	}
	// Create a big integer by diffnum
	newInt := big.NewInt(1)
	newInt.Lsh(newInt, 256-newBlock.DiffNum) // 相当于左移 1<<256-diffNum
	for {
		n.flagMutex.RLock()
		if n.update {
			n.flagMutex.RUnlock()
			n.flagMutex.Lock()
			n.update = false
			n.flagMutex.Unlock()
			return nil
		}
		n.flagMutex.RUnlock()
		newBlock.Nonce = nonce
		newBlock.getHash()
		hashInt := big.Int{}
		hashBytes, _ := hex.DecodeString(newBlock.Hash)
		hashInt.SetBytes(hashBytes) // 把本区块 hash 值转换为一串数字
		// 如果 hash 小于挖矿难度值计算的一个大数，则代表挖矿成功
		if hashInt.Cmp(newInt) == -1 {
			break
		} else {
			nonce++ // 不满足条件，则不断递增随机数，直到本区块的散列值小于指定的大数
		}
	}
	return newBlock
}

type MsgType uint

const (
	INIT MsgType = iota
	CONTINUE
	ACKNOWLEGGE
	RECLINE
)

type ConspiratorialTarget struct {
	target  *block
	msgType MsgType
}

type Attacker struct {
	id                uint64
	secret_id         uint64
	blockChain        BlockChain
	blockChainMutex   sync.RWMutex // 锁
	peers             map[uint64]chan block
	receiveChan       chan block
	update            bool
	flagMutex         sync.RWMutex
	cahoot            map[uint64]chan ConspiratorialTarget
	secretReceiveChan chan ConspiratorialTarget
	request           bool
	requestMutex      sync.RWMutex
	target            *BlockChainNode
	targetMutex       sync.RWMutex
	votes             uint64
}

func NewAttacker(id uint64, secret_id uint64, genesisBlock block, peers map[uint64]chan block, recvChan chan block, cahoot map[uint64]chan ConspiratorialTarget, secretRecvChan chan ConspiratorialTarget) *Attacker {
	bc := NewBlockChain(&genesisBlock)
	return &Attacker{
		id:                id,
		secret_id:         secret_id,
		blockChain:        *bc,
		peers:             peers,
		receiveChan:       recvChan,
		update:            false,
		cahoot:            cahoot,
		secretReceiveChan: secretRecvChan,
		request:           false,
		target:            bc.root,
		votes:             0,
	}
}

func (a *Attacker) Run() {
	fmt.Println("start attacker : ", a.id)
	go a.Receive()

	// time.Sleep(time.Millisecond * 1500)
	for {
		a.blockChainMutex.RLock()
		if a.blockChain.maxHeight >= 100 {
			a.blockChainMutex.RUnlock()
			break
		}
		a.blockChainMutex.RUnlock()
	}
	for {
		newBlock := a.Attack()
		if newBlock == nil {
			continue
		}
		a.blockChainMutex.Lock()
		a.blockChain.append(newBlock)
		a.blockChainMutex.Unlock()
		a.Broadcast(*newBlock)
		a.targetMutex.Lock()
		a.target = a.blockChain.search(newBlock)
		a.targetMutex.Unlock()
		a.polt(ConspiratorialTarget{a.target.block, CONTINUE})
		// fmt.Println("Attacker find : ", newBlock)
	}
}

func (a *Attacker) Receive() {
	for {
		select {
		case msg := <-a.receiveChan:
			a.handler(msg)
			break
		case msg := <-a.secretReceiveChan:
			a.shandler(msg)
			break
		}
	}
}

func (a *Attacker) Broadcast(msg block) {
	for id, ch := range a.peers {
		if id == a.id {
			continue
		}
		ch <- msg
	}
}

func (a *Attacker) handler(msg block) {
	a.blockChainMutex.Lock()
	flag := a.blockChain.append(&msg)
	a.blockChainMutex.Unlock()
	if flag {
		a.flagMutex.Lock()
		a.update = true
		// fmt.Printf("node %d swtich to longest chain!\n", n.id)
		a.flagMutex.Unlock()
	}
}

func (a *Attacker) shandler(msg ConspiratorialTarget) {
	if (msg.msgType == INIT) || (msg.msgType == CONTINUE) {
		a.blockChainMutex.RLock()
		a.targetMutex.Lock()
		newTarget := a.blockChain.search(msg.target)
		if newTarget == nil {
			a.targetMutex.Unlock()
			a.blockChainMutex.RUnlock()
			return
		} else {
			// fmt.Print("GOT TARGET!!!!!!!!!!!!\n")
		}
		var flag bool = false
		if a.target.block.Hash != newTarget.block.Hash {
			a.target = newTarget
			flag = true
		}
		a.targetMutex.Unlock()
		a.blockChainMutex.RUnlock()
		if flag {
			a.requestMutex.Lock()
			a.request = true
			a.requestMutex.Unlock()
		}
	} else if msg.msgType == RECLINE {
		a.votes++
		if a.votes >= 1/2*AttackerNumber {
			a.blockChainMutex.RLock()
			a.targetMutex.Lock()
			a.target = a.blockChain.workspace.parent
			a.targetMutex.Unlock()
			a.blockChainMutex.RUnlock()
			a.requestMutex.Lock()
			a.request = true
			a.requestMutex.Unlock()
		}
	}
}

func (a *Attacker) polt(msg ConspiratorialTarget) {
	for id, ch := range a.cahoot {
		if id == a.secret_id {
			continue
		}
		ch <- msg
	}
}

func (a *Attacker) Attack() *block {
	newBlock := new(block)
	newBlock.Timestamp = time.Now().UnixMilli()
	newBlock.MinerId = a.id
	var nonce int64 = rand.Int63()
	a.targetMutex.RLock()
	lastBlock := a.target.block
	a.targetMutex.RUnlock()
	newBlock.Lasthash = lastBlock.Hash
	newBlock.Height = lastBlock.Height + 1
	newBlock.Data = "攻击者" + strconv.FormatUint(a.secret_id, 10) + "尝试分叉攻击!"
	// calculate difficulty
	newBlock.DiffNum = lastBlock.DiffNum
	if ((newBlock.Height-1)%2 == 0) && (newBlock.Height >= 4) {
		a.blockChainMutex.RLock()
		a.targetMutex.RLock()
		interval := a.blockChain.statistics(a.target)
		a.targetMutex.RUnlock()
		a.blockChainMutex.RUnlock()
		if interval < 0.9*IntervalNum*Interval {
			newBlock.DiffNum += 1
		} else if interval > 1.1*IntervalNum*Interval {
			newBlock.DiffNum -= 1
		}
	}
	// Create a big integer by diffnum
	newInt := big.NewInt(1)
	newInt.Lsh(newInt, 256-newBlock.DiffNum) // 相当于左移 1<<256-diffNum
	for {
		a.flagMutex.RLock()
		if a.update {
			a.flagMutex.RUnlock()
			a.flagMutex.Lock()
			a.update = false
			a.flagMutex.Unlock()
			a.blockChainMutex.RLock()
			if a.blockChain.maxHeight > newBlock.Height+1 {
				a.targetMutex.Lock()
				a.target = a.blockChain.workspace.parent
				a.polt(ConspiratorialTarget{a.target.block, INIT})
				a.targetMutex.Unlock()
				a.blockChainMutex.RUnlock()
				return nil
			}
			a.blockChainMutex.RUnlock()
		} else {
			a.flagMutex.RUnlock()
		}
		a.requestMutex.RLock()
		if a.request {
			a.requestMutex.RUnlock()
			a.requestMutex.Lock()
			a.request = false
			a.requestMutex.Unlock()
			return nil
		}
		a.requestMutex.RUnlock()
		newBlock.Nonce = nonce
		newBlock.getHash()
		hashInt := big.Int{}
		hashBytes, _ := hex.DecodeString(newBlock.Hash)
		hashInt.SetBytes(hashBytes)
		// hash < bigint --> success
		if hashInt.Cmp(newInt) == -1 {
			break
		} else {
			nonce++ // not satisfied, make nonce bigger
		}
	}
	if !newBlock.IsValid() {
		fmt.Println("NewBlock is invalid")
	}
	return newBlock
}

type Monitor struct {
	id          uint64
	blockChain  BlockChain
	peers       map[uint64]chan block
	receiveChan chan block
}

func NewMonitor(id uint64, genesisBlock block, peers map[uint64]chan block, recvChan chan block) *Monitor {
	return &Monitor{
		id:          id,
		blockChain:  *NewBlockChain(&genesisBlock),
		peers:       peers,
		receiveChan: recvChan,
	}
}

func (m *Monitor) Run() {
	fmt.Println("start Monitor : ", m.id)
	go m.Receive()
}

func (m *Monitor) Receive() {
	for {
		select {
		case msg := <-m.receiveChan:
			m.handler(msg)
		}
	}
}

func (m *Monitor) handler(msg block) {
	m.blockChain.append(&msg)
	fmt.Printf("%f\t%d\n", float64(m.blockChain.workspace.block.Timestamp-m.blockChain.root.block.Timestamp)/float64(m.blockChain.maxHeight-1),
		m.blockChain.workspace.block.DiffNum) // Adding rate, and diffnum
}
