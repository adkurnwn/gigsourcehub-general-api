package request_model

type CreateConversationRequest struct {
	CandidateUserID string `json:"candidate_user_id" binding:"required"`
	SubrequestID    string `json:"subrequest_id" binding:"required"`
}

type SendMessageRequest struct {
	Content          string  `json:"content" binding:"required"`
	ReplyToMessageID *string `json:"reply_to_message_id"`
}
