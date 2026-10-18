package moodle

import "context"

type CourseGrade struct {
	CourseID Int    `json:"courseid"`
	Grade    string `json:"grade"`
	RawGrade string `json:"rawgrade"`
	Rank     Int    `json:"rank"`
}

func (c *Client) CourseGrades(ctx context.Context, userID Int) ([]CourseGrade, error) {
	var out struct {
		Grades []CourseGrade `json:"grades"`
	}
	if err := c.Call(ctx, "gradereport_overview_get_course_grades", map[string]any{"userid": userID}, &out); err != nil {
		return nil, err
	}
	return out.Grades, nil
}

type GradeItem struct {
	ID           Int    `json:"id"`
	Name         string `json:"itemname"`
	Type         string `json:"itemtype"`
	Module       string `json:"itemmodule"`
	Instance     Int    `json:"iteminstance"`
	CourseModule Int    `json:"cmid"`
	Weight       string `json:"weightformatted"`
	Grade        string `json:"gradeformatted"`
	Raw          *Float `json:"graderaw"`
	Min          Float  `json:"grademin"`
	Max          Float  `json:"grademax"`
	Range        string `json:"rangeformatted"`
	Percentage   string `json:"percentageformatted"`
	Letter       string `json:"lettergradeformatted"`
	Feedback     string `json:"feedback"`
	Graded       Time   `json:"gradedategraded"`
	Hidden       Bool   `json:"gradehiddenbydate"`
	NeedsUpdate  Bool   `json:"gradeneedsupdate"`
	IsHidden     Bool   `json:"gradeishidden"`
	Locked       Bool   `json:"locked"`
}

type UserGrades struct {
	CourseID Int         `json:"courseid"`
	UserID   Int         `json:"userid"`
	FullName string      `json:"userfullname"`
	Items    []GradeItem `json:"gradeitems"`
}

func (c *Client) GradeItems(ctx context.Context, courseID, userID Int) ([]GradeItem, error) {
	var out struct {
		UserGrades []UserGrades `json:"usergrades"`
	}
	if err := c.Call(ctx, "gradereport_user_get_grade_items", map[string]any{"courseid": courseID, "userid": userID}, &out); err != nil {
		return nil, err
	}
	var items []GradeItem
	for _, u := range out.UserGrades {
		items = append(items, u.Items...)
	}
	return items, nil
}
