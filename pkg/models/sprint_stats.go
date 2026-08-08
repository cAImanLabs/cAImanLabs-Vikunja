// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-present Vikunja and contributors. All rights reserved.

package models

import (
	"time"

	"xorm.io/builder"
	"xorm.io/xorm"
)

// SprintBurndownPoint is the remaining story points as of one day of a sprint.
type SprintBurndownPoint struct {
	// The date this point covers.
	Date time.Time `json:"date" doc:"The date of this point in the burndown."`
	// The total story points not yet done as of this date.
	Remaining int `json:"remaining" doc:"The total story points across the sprint's tasks that were not yet done as of this date."`
}

// SprintBurndown is the burndown chart data for a single sprint: one point
// per day from its start date through today (or its end date, whichever is
// earlier), plus what's needed to draw the ideal line alongside it.
type SprintBurndown struct {
	SprintID int64 `json:"sprint_id" doc:"The id of the sprint this burndown belongs to."`
	// The total story points assigned to the sprint's tasks.
	TotalPoints int `json:"total_points" doc:"The total story points across all of the sprint's tasks, regardless of done state."`
	// The sprint's start date.
	StartDate time.Time `json:"start_date" doc:"The sprint's start date."`
	// The sprint's end date.
	EndDate time.Time `json:"end_date" doc:"The sprint's end date."`
	// One point per day, in chronological order.
	Series []*SprintBurndownPoint `json:"series" doc:"One point per day from the sprint's start date through today (or its end date, whichever is earlier)."`
}

// GetSprintBurndown computes the burndown series for a sprint: for every day
// from its start date through today (capped at its end date), how many story
// points are still not done. This is derived entirely from each task's
// story_points and done_at — no separate daily snapshot table is needed,
// since a task's done_at is a permanent record of when it left the remaining set.
func GetSprintBurndown(s *xorm.Session, projectID, sprintID int64) (*SprintBurndown, error) {
	sprint, err := getSprintByID(s, sprintID)
	if err != nil {
		return nil, err
	}
	if sprint.ProjectID != projectID {
		return nil, ErrSprintDoesNotExist{SprintID: sprintID}
	}
	if sprint.StartDate.IsZero() || sprint.EndDate.IsZero() {
		return nil, ErrSprintMissingDates{SprintID: sprintID}
	}

	tasks := []*Task{}
	err = s.
		Where("sprint_id = ?", sprintID).
		Cols("story_points", "done", "done_at").
		Find(&tasks)
	if err != nil {
		return nil, err
	}

	total := 0
	for _, t := range tasks {
		total += t.StoryPoints
	}

	today := truncateToDay(time.Now())
	end := truncateToDay(sprint.EndDate)
	if today.Before(end) {
		end = today
	}
	start := truncateToDay(sprint.StartDate)

	series := []*SprintBurndownPoint{}
	for day := start; !day.After(end); day = day.AddDate(0, 0, 1) {
		dayEnd := day.AddDate(0, 0, 1)
		completed := 0
		for _, t := range tasks {
			if t.Done && !t.DoneAt.IsZero() && t.DoneAt.Before(dayEnd) {
				completed += t.StoryPoints
			}
		}
		series = append(series, &SprintBurndownPoint{
			Date:      day,
			Remaining: total - completed,
		})
	}

	return &SprintBurndown{
		SprintID:    sprintID,
		TotalPoints: total,
		StartDate:   sprint.StartDate,
		EndDate:     sprint.EndDate,
		Series:      series,
	}, nil
}

func truncateToDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

// SprintVelocityPoint is the story-point totals for one sprint, used to plot
// a project's velocity across its sprints.
type SprintVelocityPoint struct {
	SprintID int64  `json:"sprint_id" doc:"The id of this sprint."`
	Title    string `json:"title" doc:"The title of this sprint."`
	// The total story points across the sprint's tasks.
	TotalPoints int `json:"total_points" doc:"The total story points across all of the sprint's tasks."`
	// The story points of tasks marked done.
	CompletedPoints int `json:"completed_points" doc:"The story points of the sprint's tasks that are marked done."`
}

// GetProjectVelocity returns one point per sprint in a project, ordered by
// start date, with the total and completed story points of each. Requires
// read access to the project (checked by the caller).
func GetProjectVelocity(s *xorm.Session, projectID int64) ([]*SprintVelocityPoint, error) {
	sprints := []*Sprint{}
	err := s.
		Where("project_id = ?", projectID).
		OrderBy("start_date asc, id asc").
		Find(&sprints)
	if err != nil {
		return nil, err
	}
	if len(sprints) == 0 {
		return []*SprintVelocityPoint{}, nil
	}

	sprintIDs := make([]int64, 0, len(sprints))
	for _, sp := range sprints {
		sprintIDs = append(sprintIDs, sp.ID)
	}

	tasks := []*Task{}
	err = s.
		Where(builder.In("sprint_id", sprintIDs)).
		Cols("sprint_id", "story_points", "done").
		Find(&tasks)
	if err != nil {
		return nil, err
	}

	totals := map[int64]int{}
	completed := map[int64]int{}
	for _, t := range tasks {
		totals[t.SprintID] += t.StoryPoints
		if t.Done {
			completed[t.SprintID] += t.StoryPoints
		}
	}

	result := make([]*SprintVelocityPoint, 0, len(sprints))
	for _, sp := range sprints {
		result = append(result, &SprintVelocityPoint{
			SprintID:        sp.ID,
			Title:           sp.Title,
			TotalPoints:     totals[sp.ID],
			CompletedPoints: completed[sp.ID],
		})
	}
	return result, nil
}
