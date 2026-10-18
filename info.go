package moodle

import (
	"context"
	"slices"
)

type Function struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type SiteInfo struct {
	SiteName        string     `json:"sitename"`
	Username        string     `json:"username"`
	FirstName       string     `json:"firstname"`
	LastName        string     `json:"lastname"`
	FullName        string     `json:"fullname"`
	Lang            string     `json:"lang"`
	UserID          Int        `json:"userid"`
	SiteURL         string     `json:"siteurl"`
	UserPictureURL  string     `json:"userpictureurl"`
	Functions       []Function `json:"functions"`
	DownloadFiles   Bool       `json:"downloadfiles"`
	UploadFiles     Bool       `json:"uploadfiles"`
	Release         string     `json:"release"`
	Version         string     `json:"version"`
	MobileCSSURL    string     `json:"mobilecssurl"`
	UserCanManage   Bool       `json:"usercanmanageownfiles"`
	UserQuota       Int        `json:"userquota"`
	UserMaxUpload   Int        `json:"usermaxuploadfilesize"`
	UserHomePage    Int        `json:"userhomepage"`
	UserPrivateKey  string     `json:"userprivateaccesskey"`
	SiteID          Int        `json:"siteid"`
	SiteCalendar    string     `json:"sitecalendartype"`
	UserCalendar    string     `json:"usercalendartype"`
	UserIsSiteAdmin Bool       `json:"userissiteadmin"`
	Theme           string     `json:"theme"`
	TimeZone        string     `json:"usertimezone"`
}

func (i *SiteInfo) Has(function string) bool {
	return slices.ContainsFunc(i.Functions, func(f Function) bool { return f.Name == function })
}

func (c *Client) SiteInfo(ctx context.Context) (*SiteInfo, error) {
	var out SiteInfo
	if err := c.Call(ctx, "core_webservice_get_site_info", nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
