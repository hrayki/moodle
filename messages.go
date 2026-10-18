package moodle

import "context"

type Member struct {
	ID       Int    `json:"id"`
	FullName string `json:"fullname"`
}

type Message struct {
	ID      Int    `json:"id"`
	From    Int    `json:"useridfrom"`
	Text    string `json:"text"`
	Created Time   `json:"timecreated"`
}

type Conversation struct {
	ID       Int       `json:"id"`
	Name     string    `json:"name"`
	Type     Int       `json:"type"`
	Members  []Member  `json:"members"`
	Messages []Message `json:"messages"`
	Unread   *Int      `json:"unreadcount"`
	IsRead   Bool      `json:"isread"`
	Muted    Bool      `json:"ismuted"`
}

func (c *Client) Conversations(ctx context.Context, userID Int, limit int) ([]Conversation, error) {
	var out struct {
		Conversations []Conversation `json:"conversations"`
	}
	params := map[string]any{"userid": userID, "limitfrom": 0, "limitnum": limit}
	if err := c.Call(ctx, "core_message_get_conversations", params, &out); err != nil {
		return nil, err
	}
	return out.Conversations, nil
}

func (c *Client) ConversationMessages(ctx context.Context, userID, conversationID Int, limit int) (*Conversation, error) {
	var out Conversation
	params := map[string]any{"currentuserid": userID, "convid": conversationID, "limitfrom": 0, "limitnum": limit, "newest": true}
	if err := c.Call(ctx, "core_message_get_conversation_messages", params, &out); err != nil {
		return nil, err
	}
	out.ID = conversationID
	return &out, nil
}

func (c *Client) SendToConversation(ctx context.Context, conversationID Int, text string) ([]Message, error) {
	var out []Message
	params := map[string]any{"conversationid": conversationID, "messages": []map[string]any{{"text": text, "textformat": 1}}}
	if err := c.Call(ctx, "core_message_send_messages_to_conversation", params, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) DeleteMessage(ctx context.Context, userID, messageID Int) error {
	return c.Call(ctx, "core_message_delete_message", map[string]any{"messageid": messageID, "userid": userID, "read": true}, nil)
}

type Notification struct {
	ID          Int    `json:"id"`
	From        Int    `json:"useridfrom"`
	Subject     string `json:"subject"`
	Short       string `json:"shortenedsubject"`
	Text        string `json:"text"`
	Small       string `json:"smallmessage"`
	Component   string `json:"component"`
	EventType   string `json:"eventtype"`
	ContextURL  string `json:"contexturl"`
	ContextName string `json:"contexturlname"`
	Created     Time   `json:"timecreated"`
	Read        Bool   `json:"read"`
}

func (c *Client) Notifications(ctx context.Context, userID Int, limit int) ([]Notification, error) {
	var out struct {
		Notifications []Notification `json:"notifications"`
		Unread        Int            `json:"unreadcount"`
	}
	params := map[string]any{"useridto": userID, "newestfirst": 1, "limit": limit, "offset": 0}
	if err := c.Call(ctx, "message_popup_get_popup_notifications", params, &out); err != nil {
		return nil, err
	}
	return out.Notifications, nil
}
