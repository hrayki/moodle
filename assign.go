package moodle

import (
	"bytes"
	"context"
	"encoding/json"
)

type AssignConfig struct {
	Plugin  string `json:"plugin"`
	Subtype string `json:"subtype"`
	Name    string `json:"name"`
	Value   string `json:"value"`
}

type Assignment struct {
	ID               Int            `json:"id"`
	CourseModule     Int            `json:"cmid"`
	Course           Int            `json:"course"`
	Name             string         `json:"name"`
	Intro            string         `json:"intro"`
	IntroFiles       []File         `json:"introfiles"`
	IntroAttachments []File         `json:"introattachments"`
	Activity         string         `json:"activity"`
	Due              Time           `json:"duedate"`
	AllowFrom        Time           `json:"allowsubmissionsfromdate"`
	Cutoff           Time           `json:"cutoffdate"`
	GradingDue       Time           `json:"gradingduedate"`
	Grade            Int            `json:"grade"`
	NoSubmissions    Bool           `json:"nosubmissions"`
	SubmissionDrafts Bool           `json:"submissiondrafts"`
	RequireStatement Bool           `json:"requiresubmissionstatement"`
	TeamSubmission   Bool           `json:"teamsubmission"`
	MaxAttempts      Int            `json:"maxattempts"`
	Configs          []AssignConfig `json:"configs"`
	TimeModified     Time           `json:"timemodified"`
}

func (a *Assignment) Config(plugin, name string) string {
	for _, c := range a.Configs {
		if c.Plugin == plugin && c.Name == name {
			return c.Value
		}
	}
	return ""
}

func (a *Assignment) AcceptsText() bool { return a.Config("onlinetext", "enabled") == "1" }

func (a *Assignment) AcceptsFiles() bool { return a.Config("file", "enabled") == "1" }

type AssignmentCourse struct {
	ID          Int          `json:"id"`
	FullName    string       `json:"fullname"`
	ShortName   string       `json:"shortname"`
	Assignments []Assignment `json:"assignments"`
}

func (c *Client) Assignments(ctx context.Context, courseIDs []Int) ([]AssignmentCourse, error) {
	params := map[string]any{}
	if len(courseIDs) > 0 {
		params["courseids"] = courseIDs
	}
	var out struct {
		Courses []AssignmentCourse `json:"courses"`
	}
	if err := c.Call(ctx, "mod_assign_get_assignments", params, &out); err != nil {
		return nil, err
	}
	return out.Courses, nil
}

type EditorField struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Text        string `json:"text"`
	Format      Int    `json:"format"`
}

type FileArea struct {
	Area  string `json:"area"`
	Files []File `json:"files"`
}

type SubmissionPlugin struct {
	Type         string        `json:"type"`
	Name         string        `json:"name"`
	FileAreas    []FileArea    `json:"fileareas"`
	EditorFields []EditorField `json:"editorfields"`
}

type Submission struct {
	ID           Int                `json:"id"`
	UserID       Int                `json:"userid"`
	AttemptNum   Int                `json:"attemptnumber"`
	Created      Time               `json:"timecreated"`
	Modified     Time               `json:"timemodified"`
	Started      Time               `json:"timestarted"`
	Status       string             `json:"status"`
	GroupID      Int                `json:"groupid"`
	Latest       Bool               `json:"latest"`
	Plugins      []SubmissionPlugin `json:"plugins"`
	GradingState string             `json:"gradingstatus"`
}

type LastAttempt struct {
	Submission         *Submission `json:"submission"`
	TeamSubmission     *Submission `json:"teamsubmission"`
	SubmissionsEnabled Bool        `json:"submissionsenabled"`
	Locked             Bool        `json:"locked"`
	Graded             Bool        `json:"graded"`
	CanEdit            Bool        `json:"canedit"`
	CaneditOwner       Bool        `json:"caneditowner"`
	CanSubmit          Bool        `json:"cansubmit"`
	Extension          Time        `json:"extensionduedate"`
	TimeLimit          Int         `json:"timelimit"`
	BlindMarking       Bool        `json:"blindmarking"`
	GradingStatus      string      `json:"gradingstatus"`
}

type GradeInfo struct {
	Grade    string `json:"grade"`
	Grader   Int    `json:"grader"`
	Modified Time   `json:"timemodified"`
}

type Feedback struct {
	Grade           *GradeInfo         `json:"grade"`
	GradeForDisplay string             `json:"gradefordisplay"`
	GradedDate      Time               `json:"gradeddate"`
	Plugins         []SubmissionPlugin `json:"plugins"`
}

type SubmissionStatus struct {
	LastAttempt *LastAttempt    `json:"lastattempt"`
	Feedback    *Feedback       `json:"feedback"`
	Previous    json.RawMessage `json:"previousattempts"`
	Warnings    []Warning       `json:"warnings"`
}

func (c *Client) SubmissionStatus(ctx context.Context, assignID Int) (*SubmissionStatus, error) {
	var out SubmissionStatus
	if err := c.Call(ctx, "mod_assign_get_submission_status", map[string]any{"assignid": assignID}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

type SubmissionInput struct {
	Text      string
	DraftItem Int
}

func (c *Client) SaveSubmission(ctx context.Context, assignID Int, in SubmissionInput) ([]Warning, error) {
	plugin := map[string]any{}
	if in.Text != "" {
		plugin["onlinetext_editor"] = map[string]any{"text": in.Text, "format": 1, "itemid": 0}
	}
	if in.DraftItem != 0 {
		plugin["files_filemanager"] = in.DraftItem
	}
	return c.callWarnings(ctx, "mod_assign_save_submission", map[string]any{"assignmentid": assignID, "plugindata": plugin})
}

func (c *Client) SubmitForGrading(ctx context.Context, assignID Int, acceptStatement bool) ([]Warning, error) {
	return c.callWarnings(ctx, "mod_assign_submit_for_grading", map[string]any{"assignmentid": assignID, "acceptsubmissionstatement": acceptStatement})
}

func (c *Client) RemoveSubmission(ctx context.Context, assignID, userID Int) ([]Warning, error) {
	return c.callWarnings(ctx, "mod_assign_remove_submission", map[string]any{"userid": userID, "assignid": assignID})
}

func (c *Client) callWarnings(ctx context.Context, function string, params any) ([]Warning, error) {
	raw, err := c.CallRaw(ctx, function, params)
	if err != nil {
		return nil, err
	}
	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 || string(raw) == "null" {
		return nil, nil
	}
	var list []Warning
	if raw[0] == '[' {
		err := json.Unmarshal(raw, &list)
		return list, err
	}
	var obj struct {
		Status   *Bool     `json:"status"`
		Warnings []Warning `json:"warnings"`
	}
	if err := json.Unmarshal(raw, &obj); err != nil {
		return nil, err
	}
	if obj.Status != nil && !*obj.Status && len(obj.Warnings) == 0 {
		obj.Warnings = []Warning{{WarningCode: "failed", Message: function + " didn't succeed"}}
	}
	return obj.Warnings, nil
}
