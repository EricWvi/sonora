package service

import (
	"context"
	"sync"
)

var (
	_, WorkerCancel = context.WithCancel(context.Background())
	WorkerWg        = sync.WaitGroup{}
)
