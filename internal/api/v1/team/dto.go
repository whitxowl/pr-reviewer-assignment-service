package team

import "github.com/whitxowl/pr-reviewer-assignment-service.git/internal/domain"

type CreateTeamRequest struct {
	TeamName string              `json:"team_name" binding:"required"`
	Members  []TeamMemberRequest `json:"members" binding:"required,dive"`
}

type TeamMemberRequest struct {
	UserID   string `json:"user_id" binding:"required"`
	Username string `json:"username" binding:"required"`
	IsActive bool   `json:"is_active"`
}

type CreateTeamResponse struct {
	Team TeamResponse `json:"team"`
}

type TeamResponse struct {
	TeamName string               `json:"team_name"`
	Members  []TeamMemberResponse `json:"members"`
}

type TeamMemberResponse struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	IsActive bool   `json:"is_active"`
}

type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (r *CreateTeamRequest) ToDomain() domain.Team {
	team := domain.Team{
		TeamName: r.TeamName,
		Members:  make([]*domain.User, len(r.Members)),
	}

	for i, member := range r.Members {
		team.Members[i] = &domain.User{
			UserID:   member.UserID,
			Username: member.Username,
			TeamName: r.TeamName,
			IsActive: member.IsActive,
		}
	}

	return team
}

func ToTeamResponse(team *domain.Team) CreateTeamResponse {
	response := CreateTeamResponse{
		Team: TeamResponse{
			TeamName: team.TeamName,
			Members:  make([]TeamMemberResponse, len(team.Members)),
		},
	}

	for i, member := range team.Members {
		response.Team.Members[i] = TeamMemberResponse{
			UserID:   member.UserID,
			Username: member.Username,
			IsActive: member.IsActive,
		}
	}

	return response
}
