package moodle

import "context"

type Course struct {
	ID           Int    `json:"id"`
	ShortName    string `json:"shortname"`
	FullName     string `json:"fullname"`
	DisplayName  string `json:"displayname"`
	IDNumber     string `json:"idnumber"`
	Summary      string `json:"summary"`
	Format       string `json:"format"`
	Category     Int    `json:"category"`
	CourseImage  string `json:"courseimage"`
	Progress     *Float `json:"progress"`
	Completed    Bool   `json:"completed"`
	StartDate    Time   `json:"startdate"`
	EndDate      Time   `json:"enddate"`
	LastAccess   Time   `json:"lastaccess"`
	Favourite    Bool   `json:"isfavourite"`
	Hidden       Bool   `json:"hidden"`
	Visible      Bool   `json:"visible"`
	ShowGrades   Bool   `json:"showgrades"`
	TimeModified Time   `json:"timemodified"`
	Overview     []File `json:"overviewfiles"`
}

func (c *Client) Courses(ctx context.Context, userID Int) ([]Course, error) {
	var out []Course
	if err := c.Call(ctx, "core_enrol_get_users_courses", map[string]any{"userid": userID, "returnusercount": false}, &out); err != nil {
		return nil, err
	}
	return out, nil
}

type ModuleDate struct {
	Label     string `json:"label"`
	Timestamp Time   `json:"timestamp"`
	DataID    string `json:"dataid"`
}

type Completion struct {
	State         Int  `json:"state"`
	TimeCompleted Time `json:"timecompleted"`
	HasCompletion Bool `json:"hascompletion"`
	IsAutomatic   Bool `json:"isautomatic"`
}

type Module struct {
	ID          Int          `json:"id"`
	URL         string       `json:"url"`
	Name        string       `json:"name"`
	Instance    Int          `json:"instance"`
	ContextID   Int          `json:"contextid"`
	Description string       `json:"description"`
	Visible     Bool         `json:"visible"`
	UserVisible Bool         `json:"uservisible"`
	Available   string       `json:"availabilityinfo"`
	ModName     string       `json:"modname"`
	ModPlural   string       `json:"modplural"`
	Dates       []ModuleDate `json:"dates"`
	Completion  *Completion  `json:"completiondata"`
	Contents    []File       `json:"contents"`
	NoViewLink  Bool         `json:"noviewlink"`
}

type Section struct {
	ID          Int      `json:"id"`
	Name        string   `json:"name"`
	Number      Int      `json:"section"`
	Summary     string   `json:"summary"`
	Visible     Bool     `json:"visible"`
	UserVisible Bool     `json:"uservisible"`
	Available   string   `json:"availabilityinfo"`
	Modules     []Module `json:"modules"`
}

func (c *Client) CourseContents(ctx context.Context, courseID Int) ([]Section, error) {
	var out []Section
	if err := c.Call(ctx, "core_course_get_contents", map[string]any{"courseid": courseID}, &out); err != nil {
		return nil, err
	}
	return out, nil
}

type CourseResult struct {
	ID           Int    `json:"id"`
	FullName     string `json:"fullname"`
	DisplayName  string `json:"displayname"`
	ShortName    string `json:"shortname"`
	Summary      string `json:"summary"`
	CategoryName string `json:"categoryname"`
}

func (c *Client) SearchCourses(ctx context.Context, query string, limit int) ([]CourseResult, error) {
	var out struct {
		Total   Int            `json:"total"`
		Courses []CourseResult `json:"courses"`
	}
	params := map[string]any{"criterianame": "search", "criteriavalue": query, "page": 0, "perpage": limit}
	if err := c.Call(ctx, "core_course_search_courses", params, &out); err != nil {
		return nil, err
	}
	return out.Courses, nil
}
