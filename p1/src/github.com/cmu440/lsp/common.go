package lsp

import (
	"github.com/cmu440/lspnet"
	"sync"
)

type Activity struct {
	msg  *Message
	sig  chan struct{}
	done bool
}

type SafeNum struct {
	sync.Mutex
	n int
}
