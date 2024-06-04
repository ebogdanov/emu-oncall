package user

type Slack struct {
	UserID string `json:"user_id"`
	TeamID string `json:"team_id"`
}

type Item struct {
	ID                    string `json:"id"`
	Email                 string `json:"email"`
	Username              string `json:"username"`
	Role                  string `json:"role"`
	PhoneNumber           string `json:"phone_number,omitempty"`
	IsPhoneNumberVerified bool   `json:"is_phone_number_verified"`
	Slack                 *Slack `json:"slack,omitempty"`
}

type List struct {
	Count    int64   `json:"count"`
	Next     *uint64 `json:"next"`
	Previous *uint64 `json:"previous"`
	Result   []Item  `json:"results"`
}
