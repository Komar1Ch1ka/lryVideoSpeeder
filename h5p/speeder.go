package h5p

import (
	"fmt"
	"sync"
	"time"
)

type Speeder struct {
	client   *Client
	settings Settings
}

func NewSpeeder(client *Client, settings Settings) *Speeder {
	return &Speeder{
		client:   client,
		settings: settings,
	}
}

func (s *Speeder) Run(course Course) error {
	cmid, err := ExtractCmidFromURL(course.URL)
	if err != nil {
		return fmt.Errorf("解析课程ID失败: %v", err)
	}

	duration, err := s.client.GetVideoDuration(course.URL)
	if err != nil {
		duration = 600
	}

	targetProgress := float64(s.settings.TargetProgress)
	stepProgress := float64(s.settings.StepProgress)
	interval := time.Duration(s.settings.IntervalSeconds*1000) * time.Millisecond

	currentProgress := 0.0
	for currentProgress < targetProgress {
		nextProgress := currentProgress + stepProgress
		if nextProgress > targetProgress {
			nextProgress = targetProgress
		}

		timeSpent := int(float64(duration) * nextProgress / 100)
		finish := 0
		if nextProgress >= 100 {
			finish = 1
		}

		resp, err := s.client.SendProgress(cmid, duration, nextProgress, timeSpent, finish)
		if err != nil {
			time.Sleep(interval)
			continue
		}

		if resp.Error {
			msg := "未知错误"
			if resp.Exception != nil {
				msg = resp.Exception.Message
			}
			return fmt.Errorf("server error: %s", msg)
		}

		currentProgress = nextProgress

		if currentProgress < targetProgress {
			time.Sleep(interval)
		}
	}

	if targetProgress < 100 {
		timeSpent := duration
		s.client.SendProgress(cmid, duration, 100, timeSpent, 1)
	}

	return nil
}

func (s *Speeder) RunAll(courses []Course) error {
	if len(courses) == 0 {
		return nil
	}

	maxConcurrent := s.settings.MaxConcurrent
	if maxConcurrent <= 0 {
		maxConcurrent = 10
	}

	startTime := time.Now()
	fmt.Printf("开始并行执行 %d 个课程，最大并发数: %d\n\n", len(courses), maxConcurrent)

	sem := make(chan struct{}, maxConcurrent)
	var wg sync.WaitGroup
	var mu sync.Mutex
	successCount := 0
	failCount := 0

	for _, course := range courses {
		wg.Add(1)
		go func(c Course) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			courseStart := time.Now()
			if err := s.Run(c); err != nil {
				mu.Lock()
				failCount++
				elapsed := time.Since(courseStart)
				fmt.Printf("[失败] %s (%.1fs) - %v\n", c.Name, elapsed.Seconds(), err)
				mu.Unlock()
			} else {
				mu.Lock()
				successCount++
				elapsed := time.Since(courseStart)
				fmt.Printf("[完成] %s (%.1fs)\n", c.Name, elapsed.Seconds())
				mu.Unlock()
			}
		}(course)
	}

	wg.Wait()

	totalElapsed := time.Since(startTime)
	fmt.Printf("\n执行完成！总耗时: %.1fs, 成功: %d, 失败: %d\n", totalElapsed.Seconds(), successCount, failCount)

	if failCount > 0 {
		return fmt.Errorf("%d 个课程执行失败", failCount)
	}

	return nil
}
