package dto

type UserProfileUpdateRequest struct {
	NativeLanguage *string `json:"nativeLanguage,omitempty" binding:"omitempty,len=2"`
	TargetLanguage *string `json:"targetLanguage,omitempty" binding:"omitempty,len=2"`
}
