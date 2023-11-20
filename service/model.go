package service

type WXCode2SessionResponse struct {
	SessionKey string `json:"session_key"`
	Unionid    string `json:"unionid"`
	Errmsg     string `json:"errmsg"`
	Openid     string `json:"openid"`
	Errcode    int32  `json:"errcode"`
}

type LoginRequest struct {
	Code   string `json:"code"`
	Secret string `json:"secret"`
}

type LoginResponse struct {
	Token       string `json:"token"`
	GroupNumber uint   `json:"group_number"`
}

type Round2StateResponse struct {
	Start          bool `json:"start"`
	RemainingTime  int  `json:"remaining_time"`
	PromotionCount int  `json:"promotion_count"`
	Target         int  `json:"target"`
}

type Round2GetQuestionResponse struct {
	Number     uint              `json:"number"`
	IsMultiple bool              `json:"is_multiple"`
	Content    string            `json:"content"`
	Options    map[string]string `json:"options"`
}

type Round2SubmitRequest struct {
	Number uint   `json:"number"`
	Answer string `json:"answer"`
}

type Round2SubmitResponse struct {
	Correct bool `json:"correct"`
	Score   int  `json:"score"`
}

type StartRound2Request struct {
	TargetScore          int `json:"target_score"`
	TargetPromotionCount int `json:"target_promotion_count"`
	RemainingTime        int `json:"remaining_time"`
}

type StartRound1Request struct {
	Candidates            []uint `json:"candidates"`
	TargetEliminatedCount int    `json:"target_eliminated_count"`
	AnswerTime            int    `json:"answer_time"`
	InitialScore          int    `json:"initial_score"`
}
