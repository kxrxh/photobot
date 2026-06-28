package dto

type ClassificationsResponse struct {
	MainGroups    map[string]string `json:"main_groups"`
	MainSubgroups map[string]string `json:"main_subgroups"`
	Subgroups     map[string]string `json:"subgroups"`
	Hierarchy     any               `json:"hierarchy"`
}
