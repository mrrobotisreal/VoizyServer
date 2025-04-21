package models

type SearchPeopleRequest struct {
	Query string `json:"query"`
}

type SearchPerson struct {
	UserID          int64   `json:"userID"`
	Username        *string `json:"username"`
	FirstName       *string `json:"firstName"`
	LastName        *string `json:"lastName"`
	PreferredName   *string `json:"preferredName"`
	ProfilePicURL   *string `json:"profilePicURL"`
	CityOfResidence *string `json:"cityOfResidence"`
	FriendsInCommon []int64 `json:"friendsInCommon"`
}

type SearchPeopleResponse struct {
	Results []SearchPerson `json:"results"`
	Limit   int64          `json:"limit"`
	Page    int64          `json:"page"`
}
