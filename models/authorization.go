// models/authorization.go
package models

import (
	"gorm.io/gorm"
)

// UserIsProjectOwner checks if a user is the owner of a specific project
func UserIsProjectOwner(userID uint, projectID uint) (bool, error) {
	var count int64
	err := DB.Model(&Project{}).
		Where("id = ? AND owner_id = ?", projectID, userID).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// UserHasAccessToProject checks if a user has access to a specific project
// Either as the owner or as a member of any team associated with the project
func UserHasAccessToProject(userID uint, projectID uint) (bool, error) {
	// Check if the user is the owner of the project
	isOwner, err := UserIsProjectOwner(userID, projectID)
	if err != nil {
		return false, err
	}
	if isOwner {
		return true, nil
	}

	// Check if the user is a member of any team associated with the project
	var count int64
	err = DB.Table("team_members").
		Joins("JOIN project_teams ON team_members.team_id = project_teams.team_id").
		Where("project_teams.project_id = ? AND team_members.user_id = ?", projectID, userID).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// UserIsMemberOfProjectTeams checks if a user is a member of any team associated with a project
func UserIsMemberOfProjectTeams(userID uint, projectID uint) (bool, error) {
	var count int64
	err := DB.Table("team_members").
		Joins("JOIN project_teams ON team_members.team_id = project_teams.team_id").
		Where("project_teams.project_id = ? AND team_members.user_id = ?", projectID, userID).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// UserHasAccessToTask checks if a user has access to a specific task
func UserHasAccessToTask(userID uint, taskID uint) (bool, error) {
	var task Task
	err := DB.Preload("Project").First(&task, taskID).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil // Task not found
		}
		return false, err
	}
	// Check if the user has access to the project the task belongs to
	return UserHasAccessToProject(userID, task.ProjectID)
}

// UserIsTaskAssignee checks if a user is the assignee of a specific task
func UserIsTaskAssignee(userID uint, taskID uint) (bool, error) {
	var count int64
	err := DB.Model(&Task{}).
		Where("id = ? AND assigned_to_id = ?", taskID, userID).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// UserHasAccessToTeam checks if a user is a member of a specific team
func UserHasAccessToTeam(userID uint, teamID uint) (bool, error) {
	var count int64
	err := DB.Table("team_members").
		Where("team_id = ? AND user_id = ?", teamID, userID).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// UserIsTeamOwner checks if a user is the owner of a specific team
func UserIsTeamOwner(userID uint, teamID uint) (bool, error) {
	var count int64
	err := DB.Model(&Team{}).
		Where("id = ? AND owner_id = ?", teamID, userID).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// UserIsAdmin checks if a user has an admin role
func UserIsAdmin(userID uint) (bool, error) {
	var user User
	err := DB.First(&user, userID).Error
	if err != nil {
		return false, err
	}
	return user.Role == RoleAdmin, nil
}
