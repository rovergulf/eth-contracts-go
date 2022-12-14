package watcher

import (
	"context"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/rovergulf/eth-contracts-go/pkg/logutils"
	"go.uber.org/zap"
)

type EventHandler struct {
	logger *zap.SugaredLogger
	client *ethclient.Client
}

func New(ctx context.Context, providerUrl string) (*EventHandler, error) {
	logger, err := logutils.NewLogger()
	if err != nil {
		return nil, err
	}

	client, err := ethclient.DialContext(ctx, providerUrl)
	if err != nil {
		return nil, err
	}

	return &EventHandler{
		logger: logger,
		client: client,
	}, nil
}

func (e *EventHandler) Shutdown() {
	e.client.Close()
}

func (e *EventHandler) WatchEvents(ctx context.Context, addresses []common.Address, topics [][]common.Hash, callback func(ctx context.Context, eventLog types.Log)) error {
	logs := make(chan types.Log)
	sub, err := e.client.SubscribeFilterLogs(ctx, ethereum.FilterQuery{
		Addresses: addresses,
		Topics:    topics,
	}, logs)
	if err != nil {
		e.logger.Errorw("Unable to subscribe chain logs", "err", err)
		return err
	} else {
		e.logger.Infow("Subscribed to contract events", "addresses", addresses)
	}

	var breakErr error
	go func() {
	watchingLoop:
		for {
			select {
			case <-ctx.Done():
				close(logs)
				sub.Unsubscribe()
				break watchingLoop
			// handle errors
			case err := <-sub.Err():
				if err != nil {
					e.logger.Errorw("Events subscription error", "err", err)
					close(logs)
					sub.Unsubscribe()
					breakErr = err
					break watchingLoop
				}
			// handle logs
			case log, ok := <-logs:
				if !ok {
					break watchingLoop
				}
				e.logger.Infow("Event received", "address", log.Address, "tx_hash", log.TxHash)
				callback(ctx, log)
			}
		}
	}()

	return breakErr
}
