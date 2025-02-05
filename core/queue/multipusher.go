package queue

import "github.com/brucewang585/go-zero/core/errorx"

type MultiPusher struct {
	name    string
	pushers []Pusher
}

func NewMultiPusher(pushers []Pusher) Pusher {
	return &MultiPusher{
		name:    generateName(pushers),
		pushers: pushers,
	}
}

func (pusher *MultiPusher) Name() string {
	return pusher.name
}

func (pusher *MultiPusher) Push(message string) error {
	var batchError errorx.BatchError

	for _, each := range pusher.pushers {
		if err := each.Push(message); err != nil {
			batchError.Add(err)
		}
	}

	return batchError.Err()
}
