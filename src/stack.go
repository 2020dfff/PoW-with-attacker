package main

import "fmt"

type Stack struct {
	list []*BlockChainNode
}

// Initialize
func NewStack() *Stack {
	return &Stack{
		list: make([]*BlockChainNode, 0),
	}
}

// Length
func (s *Stack) Len() int {
	return len(s.list)
}

// IsEmpty?
func (s *Stack) IsEmpty() bool {
	if len(s.list) == 0 {
		return true
	} else {
		return false
	}
}

// push
func (s *Stack) Push(x *BlockChainNode) {
	s.list = append(s.list, x)
}

// push in sequence
func (s *Stack) PushList(x []*BlockChainNode) {
	s.list = append(s.list, x...)
}

// pop
func (s *Stack) Pop() *BlockChainNode {
	if len(s.list) <= 0 {
		fmt.Println("Stack is Empty")
		return nil
	} else {
		ret := s.list[len(s.list)-1]
		s.list = s.list[:len(s.list)-1]
		return ret
	}
}

// return top (nil)
func (s *Stack) Top() *BlockChainNode {
	if s.IsEmpty() == true {
		fmt.Println("Stack is Empty")
		return nil
	} else {
		return s.list[len(s.list)-1]
	}
}

// Clear stack
func (s *Stack) Clear() bool {
	if len(s.list) == 0 {
		return true
	}
	for i := 0; i < s.Len(); i++ {
		s.list[i] = nil
	}
	s.list = make([]*BlockChainNode, 0)
	return true
}

// Print stack
func (s *Stack) Show() {
	len := len(s.list)
	for i := 0; i != len; i++ {
		fmt.Println(s.Pop())
	}
}
