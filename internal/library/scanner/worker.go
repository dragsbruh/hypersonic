package scanner

import (
	"context"
	"fmt"
	"sync"
)

type WorkerPool[Job, Result any] struct {
	Jobs        []Job
	WaitGroup   *sync.WaitGroup
	JobChan     chan Job
	ResultChan  chan Result
	Concurrency int
	ErrChan     chan error
	Func        func(context.Context, Job) (Result, error)
}

func NewWorkerPool[Job, Result any](jobs []Job, concurrency int, fn func(context.Context, Job) (Result, error)) *WorkerPool[Job, Result] {
	sp := WorkerPool[Job, Result]{
		Jobs:        jobs,
		Concurrency: concurrency,
		WaitGroup:   &sync.WaitGroup{},
		JobChan:     make(chan Job, len(jobs)),
		ResultChan:  make(chan Result, len(jobs)),
		ErrChan:     make(chan error, concurrency),
		Func:        fn,
	}

	return &sp
}

func (sp *WorkerPool[Job, Result]) Worker(ctx context.Context) {
	defer sp.WaitGroup.Done()
	for job := range sp.JobChan {
		select {
		case <-ctx.Done():
			return
		default:
		}

		result, err := sp.Func(ctx, job)
		if err != nil {
			sp.ErrChan <- err
			return
		}
		sp.ResultChan <- result
	}
}

func (sp *WorkerPool[Job, Result]) Run(ctx context.Context, cb func(context.Context, Result) error) error {
	poolCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	for range sp.Concurrency {
		sp.WaitGroup.Add(1)
		go sp.Worker(poolCtx)
	}

	for _, job := range sp.Jobs {
		sp.JobChan <- job
	}
	close(sp.JobChan)

	successCtx, successCancel := context.WithCancel(ctx)
	go func() {
		sp.WaitGroup.Wait()
		successCancel()
	}()

	go func() {
		for {
			select {
			case track := <-sp.ResultChan:
				if err := cb(ctx, track); err != nil {
					sp.ErrChan <- err
					return
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	err := func() error {
		select {
		case <-successCtx.Done():
			return nil
		case <-poolCtx.Done():
			return context.Canceled
		case err := <-sp.ErrChan:
			return err
		}
	}()
	cancel()

	sp.WaitGroup.Wait()
	close(sp.ResultChan)

	return err
}

func (sp *WorkerPool[Job, Result]) RunAndCollect(ctx context.Context) ([]Result, error) {
	results := []Result{}
	err := sp.Run(ctx, func(ctx context.Context, r Result) error {
		results = append(results, r)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("run pool: %w", err)
	}

	return results, nil
}
