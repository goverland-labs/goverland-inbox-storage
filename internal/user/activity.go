package user

import (
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	activityWindow       = 15 * time.Minute    // for continuous session
	historyWindow        = 30 * 24 * time.Hour // 30 days
	maxActivityIntervals = 2
	lastActivityWindow   = 60 * time.Minute
	activityStep         = time.Minute * 15
)

var (
	// describes which indexes of 15'm contains that hours
	systemIntervals = map[string]systemActivityInterval{
		"0-3":   {0, 12},
		"3-6":   {13, 24},
		"6-9":   {25, 36},
		"9-12":  {37, 48},
		"12-15": {49, 60},
		"15-18": {61, 72},
		"18-21": {73, 84},
		"21-24": {85, 96},
	}
)

func (s *Service) TrackActivity(userID uuid.UUID) error {
	activity, err := s.repo.GetLastActivityInPeriod(userID, activityWindow)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("s.repo.GetLastActivityInPeriod: %w", err)
	}

	if activity != nil {
		activity.FinishedAt = time.Now()
		return s.repo.UpdateUserActivity(activity)
	}

	activity = &Activity{
		UserID:     userID,
		FinishedAt: time.Now(),
	}

	return s.repo.AddUserActivity(activity)
}

func (s *Service) GetLastActivity(userID uuid.UUID) (*Activity, error) {
	activity, err := s.repo.GetLastActivity(userID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("s.repo.GetLastActivity: %w", err)
	}

	return activity, nil
}

type (
	PushInterval struct {
		From time.Time
		To   time.Time
	}

	shortInterval struct {
		id       int
		duration time.Duration
	}

	systemActivityInterval struct {
		From int // 15'm interval idx
		To   int // 15'm interval idx
	}

	// splitting by N hours
	activityIntervalPeriods struct {
		// active short periods
		periods []int
		// total duration based on user activity
		duration time.Duration
		// median period
		median int
	}
)

func (s *Service) AllowSendingPush(userID uuid.UUID) (bool, error) {
	// check if user activity less than N minutes
	last, err := s.GetLastActivity(userID)
	if err != nil {
		return false, fmt.Errorf("s.GetLastActivity: %w", err)
	}
	if last != nil && time.Since(last.FinishedAt) < lastActivityWindow {
		return true, nil
	}

	// add caching for calculating activity windows to 3 hour?
	intervals, err := s.getUserActivityIntervals(userID)
	if err != nil {
		return false, fmt.Errorf("s.calculateActivePeriods: %w", err)
	}
	if intervals == nil || len(intervals) == 0 {
		return true, nil
	}

	// let's find if now is suitable for sending a push
	for _, info := range intervals {
		if info.median == convertHourToMinuteInterval(time.Now()) {
			return true, nil
		}
	}

	return false, nil
}

func (s *Service) getUserActivityIntervals(userID uuid.UUID) ([]activityIntervalPeriods, error) {
	data, ok := s.activityCache.get(userID)
	if ok {
		return data.([]activityIntervalPeriods), nil
	}

	intervals, err := s.calculateActivePeriods(userID)
	if err != nil {
		return nil, err
	}

	s.activityCache.set(userID, intervals, 3*time.Hour)

	return intervals, nil
}

func (s *Service) calculateActivePeriods(userID uuid.UUID) ([]activityIntervalPeriods, error) {
	// get activity list
	list, err := s.repo.GetByFilters([]Filter{
		ActivityFilterUserID{UserID: userID},
		ActivityFilterBetween{From: time.Now().Add(-1 * historyWindow)},
		ActivityFilterUserIDOrderBy{Field: "id", Direction: "desc"},
	})
	if err != nil {
		return nil, fmt.Errorf("s.repo.GetByFilters: %w", err)
	}

	// no activity, let's send it right now
	if len(list) == 0 {
		return nil, nil
	}

	// let's calculate short intervals
	shortIntervals := calculateShortIntervals(list)

	// let's combine for larger intervals
	intervals := make(map[string]activityIntervalPeriods)
	for _, interval := range shortIntervals {
		for idx, window := range systemIntervals {
			if interval.id < window.From || interval.id > window.To {
				continue
			}

			data := intervals[idx]
			data.periods = append(data.periods, interval.id)
			data.duration += interval.duration
			data.median = calcMedian(data.periods)
			intervals[idx] = data
		}
	}

	// converting to slice to sort it
	results := make([]activityIntervalPeriods, 0, len(intervals))
	for _, v := range intervals {
		results = append(results, v)
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].duration > results[j].duration
	})

	end := maxActivityIntervals
	if end > len(results) {
		end = len(results)
	}

	return results[:end], nil
}

// calculateShortIntervals calculate 15'm intervals based on user activity
func calculateShortIntervals(list []Activity) map[int]shortInterval {
	shortIntervals := make(map[int]shortInterval)
	for _, activity := range list {
		cursor := activity.CreatedAt

		for {
			id := convertHourToMinuteInterval(cursor)
			interval := shortIntervals[id]

			interval.id = id
			end := cursor.Add(activityStep).Truncate(15 * time.Minute)

			if end.Before(activity.FinishedAt) {
				interval.duration += end.Sub(cursor)
			} else {
				interval.duration += activity.FinishedAt.Sub(cursor)
			}

			shortIntervals[id] = interval

			if end.After(activity.FinishedAt) {
				break
			}

			cursor = end
		}
	}

	return shortIntervals
}

// todo: replace logic here if change activityStep. It's fast solution for calculating intervals with 15'm
func convertHourToMinuteInterval(point time.Time) int {
	return point.Hour()*4 + point.Minute()/15 + 1
}

func calcMedian(data []int) int {
	dataCopy := make([]int, len(data))
	copy(dataCopy, data)

	sort.Ints(dataCopy)

	var median int
	l := len(dataCopy)
	if l == 0 {
		return 0
	} else if l%2 == 0 {
		median = (dataCopy[l/2-1] + dataCopy[l/2]) / 2
	} else {
		median = dataCopy[l/2]
	}

	return median
}
