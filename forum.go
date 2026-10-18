package moodle

import "context"

type Forum struct {
	ID           Int    `json:"id"`
	Course       Int    `json:"course"`
	Type         string `json:"type"`
	Name         string `json:"name"`
	Intro        string `json:"intro"`
	CourseModule Int    `json:"cmid"`
	Discussions  Int    `json:"numdiscussions"`
	Due          Time   `json:"duedate"`
	Cutoff       Time   `json:"cutoffdate"`
	CanCreate    Bool   `json:"cancreatediscussions"`
	Modified     Time   `json:"timemodified"`
}

func (c *Client) Forums(ctx context.Context, courseIDs []Int) ([]Forum, error) {
	params := map[string]any{}
	if len(courseIDs) > 0 {
		params["courseids"] = courseIDs
	}
	var out []Forum
	if err := c.Call(ctx, "mod_forum_get_forums_by_courses", params, &out); err != nil {
		return nil, err
	}
	return out, nil
}

type Discussion struct {
	ID           Int    `json:"id"`
	Discussion   Int    `json:"discussion"`
	Name         string `json:"name"`
	Subject      string `json:"subject"`
	Message      string `json:"message"`
	UserFullName string `json:"userfullname"`
	UserID       Int    `json:"userid"`
	Created      Time   `json:"created"`
	Modified     Time   `json:"timemodified"`
	Replies      Int    `json:"numreplies"`
	Unread       Int    `json:"numunread"`
	Pinned       Bool   `json:"pinned"`
	Locked       Bool   `json:"locked"`
}

func (c *Client) Discussions(ctx context.Context, forumID Int, page, perPage int) ([]Discussion, error) {
	params := map[string]any{"forumid": forumID, "sortorder": 1, "page": page, "perpage": perPage}
	var out struct {
		Discussions []Discussion `json:"discussions"`
	}
	if err := c.Call(ctx, "mod_forum_get_forum_discussions", params, &out); err != nil {
		return nil, err
	}
	return out.Discussions, nil
}

type PostAuthor struct {
	ID       Int    `json:"id"`
	FullName string `json:"fullname"`
}

type PostCapabilities struct {
	Reply  Bool `json:"reply"`
	Edit   Bool `json:"edit"`
	Delete Bool `json:"delete"`
}

type Post struct {
	ID           Int               `json:"id"`
	Subject      string            `json:"subject"`
	Message      string            `json:"message"`
	Author       *PostAuthor       `json:"author"`
	Discussion   Int               `json:"discussionid"`
	ParentID     Int               `json:"parentid"`
	HasParent    Bool              `json:"hasparent"`
	Created      Time              `json:"timecreated"`
	Unread       Bool              `json:"unread"`
	Deleted      Bool              `json:"isdeleted"`
	Attachments  []File            `json:"attachments"`
	Capabilities *PostCapabilities `json:"capabilities"`
}

func (c *Client) DiscussionPosts(ctx context.Context, discussionID Int) ([]Post, error) {
	var out struct {
		Posts []Post `json:"posts"`
	}
	params := map[string]any{"discussionid": discussionID, "sortby": "created", "sortdirection": "ASC"}
	if err := c.Call(ctx, "mod_forum_get_discussion_posts", params, &out); err != nil {
		return nil, err
	}
	return out.Posts, nil
}

func (c *Client) Reply(ctx context.Context, postID Int, subject, message string) (Int, error) {
	var out struct {
		PostID Int `json:"postid"`
	}
	params := map[string]any{"postid": postID, "subject": subject, "message": message, "messageformat": 1}
	if err := c.Call(ctx, "mod_forum_add_discussion_post", params, &out); err != nil {
		return 0, err
	}
	return out.PostID, nil
}

func (c *Client) StartDiscussion(ctx context.Context, forumID Int, subject, message string) (Int, error) {
	var out struct {
		DiscussionID Int `json:"discussionid"`
	}
	params := map[string]any{"forumid": forumID, "subject": subject, "message": message, "messageformat": 1}
	if err := c.Call(ctx, "mod_forum_add_discussion", params, &out); err != nil {
		return 0, err
	}
	return out.DiscussionID, nil
}

func (c *Client) DeletePost(ctx context.Context, postID Int) error {
	return c.Call(ctx, "mod_forum_delete_post", map[string]any{"postid": postID}, nil)
}
