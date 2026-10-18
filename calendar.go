package moodle

import (
	"context"
	"time"
)

type EventAction struct {
	Name       string `json:"name"`
	URL        string `json:"url"`
	ItemCount  Int    `json:"itemcount"`
	Actionable Bool   `json:"actionable"`
}

type EventCourse struct {
	ID        Int    `json:"id"`
	FullName  string `json:"fullname"`
	ShortName string `json:"shortname"`
	ViewURL   string `json:"viewurl"`
}

type Event struct {
	ID           Int          `json:"id"`
	Name         string       `json:"name"`
	Description  string       `json:"description"`
	Location     string       `json:"location"`
	Component    string       `json:"component"`
	ModuleName   string       `json:"modulename"`
	ActivityName string       `json:"activityname"`
	ActivityStr  string       `json:"activitystr"`
	Instance     Int          `json:"instance"`
	EventType    string       `json:"eventtype"`
	Start        Time         `json:"timestart"`
	Duration     Int          `json:"timeduration"`
	Sort         Time         `json:"timesort"`
	Modified     Time         `json:"timemodified"`
	Overdue      Bool         `json:"overdue"`
	Course       *EventCourse `json:"course"`
	Action       *EventAction `json:"action"`
	URL          string       `json:"url"`
	ViewURL      string       `json:"viewurl"`
	CourseModule Int          `json:"cmid"`
}

func (e *Event) End() time.Time {
	return e.Start.Time().Add(time.Duration(e.Duration) * time.Second)
}

func (c *Client) Deadlines(ctx context.Context, from, to time.Time, limit int) ([]Event, error) {
	params := map[string]any{"timesortfrom": from.Unix(), "limitnum": limit}
	if !to.IsZero() {
		params["timesortto"] = to.Unix()
	}
	var out struct {
		Events []Event `json:"events"`
	}
	if err := c.Call(ctx, "core_calendar_get_action_events_by_timesort", params, &out); err != nil {
		return nil, err
	}
	return out.Events, nil
}

func (c *Client) Upcoming(ctx context.Context) ([]Event, error) {
	var out struct {
		Events []Event `json:"events"`
	}
	if err := c.Call(ctx, "core_calendar_get_calendar_upcoming_view", map[string]any{"courseid": 1, "categoryid": 0}, &out); err != nil {
		return nil, err
	}
	return out.Events, nil
}

func (c *Client) Events(ctx context.Context, from, to time.Time, courseIDs []Int) ([]Event, error) {
	params := map[string]any{
		"options": map[string]any{"userevents": true, "siteevents": true, "timestart": from.Unix(), "timeend": to.Unix()},
	}
	if len(courseIDs) > 0 {
		params["events"] = map[string]any{"courseids": courseIDs}
	}
	var out struct {
		Events []Event `json:"events"`
	}
	if err := c.Call(ctx, "core_calendar_get_calendar_events", params, &out); err != nil {
		return nil, err
	}
	return out.Events, nil
}
