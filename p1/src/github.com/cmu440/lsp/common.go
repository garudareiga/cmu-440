package lsp

import (
//	"sync"
)

type Activity struct {
	message *Message
	signal  chan struct{}
	done    bool
}

type SlideWindow struct {
	seqNumLower int
	seqNumUpper int
}
