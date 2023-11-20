package round

type Round1SliceMetadata struct {
	Type                  string `json:"type"`
	TargetEliminatedCount int    `json:"target_eliminated_count"`
	GroupCount            int    `json:"group_count"`
	EliminatedGroupCount  int    `json:"eliminated_group_count"`

	QuestionNumber uint   `json:"question_number"`
	Content        string `json:"content"`
}

type Round1RemainingTime struct {
	Type          string `json:"type"`
	RemainingTime int    `json:"remaining_time"`
}

type Round1SliceResult struct {
	Type                 string `json:"type"`
	Answer               bool   `json:"answer"`
	LastEliminatedGroups []uint `json:"last_eliminated_groups"`
}
