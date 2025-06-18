package worker

import (
	"sync"
	"time"

	"github.com/hollgett/shortener.git/internal/logger"
	"github.com/hollgett/shortener.git/internal/models"
	"go.uber.org/zap"
)

const (
	lenBuf     = 10
	timePush   = 1 * time.Second
	limitQueue = 100
)

type StoreDeleteURLs interface {
	DeleteURLs(URLs []models.DeleteURL) error
}

type DeleteWorker struct {
	logger   *logger.Logger
	store    StoreDeleteURLs
	DeleteCh chan models.DeleteURL
	DoneCh   chan struct{}
	queue    []models.DeleteURL
	ticker   *time.Ticker
	mu       *sync.Mutex
	wg       *sync.WaitGroup
}

func NewDeleteWorker(logger *logger.Logger, store StoreDeleteURLs) *DeleteWorker {
	deleteCh := make(chan models.DeleteURL, lenBuf)

	return &DeleteWorker{
		logger:   logger,
		store:    store,
		DeleteCh: deleteCh,
		DoneCh:   make(chan struct{}),
		queue:    make([]models.DeleteURL, 0),
		ticker:   time.NewTicker(timePush),
		mu:       &sync.Mutex{},
		wg:       &sync.WaitGroup{},
	}
}

func (d *DeleteWorker) Run() {
	for {
		select {
		case <-d.DoneCh:
			return
		case url := <-d.DeleteCh:
			d.add(url)
		case <-d.ticker.C:
			d.flush()
		}
	}
}

func (d *DeleteWorker) add(url models.DeleteURL) {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.queue = append(d.queue, url)
	if len(d.queue) >= limitQueue {
		d.flush()
	}
}

func (d *DeleteWorker) flush() {
	if len(d.queue) == 0 {
		return
	}
	d.flushLocker()
}

func (d *DeleteWorker) flushLocker() {
	d.mu.Lock()
	defer d.mu.Unlock()
	toDelete := make([]models.DeleteURL, len(d.queue))
	copy(toDelete, d.queue)
	d.queue = d.queue[:0]

	d.wg.Add(1)
	go func() {
		defer d.wg.Done()
		err := d.store.DeleteURLs(toDelete)
		if err != nil {
			d.logger.Info("delete flush", zap.Error(err))
		}
	}()
}

func (d *DeleteWorker) ShutDown() {
	d.ticker.Stop()

	d.flush()
	d.wg.Wait()
	close(d.DeleteCh)
	close(d.DoneCh)
}
