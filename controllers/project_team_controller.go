// controllers/project_team_controller.go
package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/mfuadfakhruzzaki/backendaurauran/models"
	"github.com/mfuadfakhruzzaki/backendaurauran/utils"
	"gorm.io/gorm"
)

// AddProjectTeam handles adding a team to a project
func AddProjectTeam(c *gin.Context) {
	// Ambil project_id dari parameter URL
	projectIDParam := c.Param("project_id")
	projectID, err := strconv.ParseUint(projectIDParam, 10, 64)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid project ID")
		return
	}

	// Ambil user_id dari konteks
	currentUserID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusInternalServerError, "User ID not found")
		return
	}
	userID := currentUserID.(uint)

	// Cek apakah pengguna adalah pemilik proyek
	isOwner, err := models.UserIsProjectOwner(userID, uint(projectID))
	if err != nil {
		utils.Logger.Errorf("Failed to check project ownership: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to add team to project")
		return
	}
	if !isOwner {
		utils.ErrorResponse(c, http.StatusForbidden, "Only the project owner can add teams to the project")
		return
	}

	// Bind JSON request untuk mendapatkan team_id
	var req struct {
		TeamID uint `json:"team_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	// Ambil tim dari database
	var team models.Team
	if err := models.DB.First(&team, req.TeamID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.ErrorResponse(c, http.StatusNotFound, "Team not found")
			return
		}
		utils.Logger.Errorf("Failed to retrieve team: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve team")
		return
	}

	// Ambil proyek dari database
	var project models.Project
	if err := models.DB.Preload("Teams").First(&project, uint(projectID)).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.ErrorResponse(c, http.StatusNotFound, "Project not found")
			return
		}
		utils.Logger.Errorf("Failed to retrieve project: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve project")
		return
	}

	// Cek apakah tim sudah terhubung dengan proyek
	for _, t := range project.Teams {
		if t.ID == team.ID {
			utils.ErrorResponse(c, http.StatusBadRequest, "Team is already associated with the project")
			return
		}
	}

	// Asosiasikan tim dengan proyek
	if err := models.DB.Model(&project).Association("Teams").Append(&team); err != nil {
		utils.Logger.Errorf("Failed to add team to project: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to add team to project")
		return
	}

	utils.Logger.Infof("Team ID %d added to Project ID %d by User ID %d", team.ID, project.ID, userID)
	utils.SuccessResponse(c, gin.H{"message": "Team added to project successfully"})
}

// ListProjectTeams handles listing all teams associated with a project
func ListProjectTeams(c *gin.Context) {
	// Ambil project_id dari parameter URL
	projectIDParam := c.Param("project_id")
	projectID, err := strconv.ParseUint(projectIDParam, 10, 64)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid project ID")
		return
	}

	// Ambil user_id dari konteks
	currentUserID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusInternalServerError, "User ID not found")
		return
	}
	userID := currentUserID.(uint)

	// Cek apakah pengguna memiliki akses ke proyek
	hasAccess, err := models.UserHasAccessToProject(userID, uint(projectID))
	if err != nil {
		utils.Logger.Errorf("Failed to check project access: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to list project teams")
		return
	}
	if !hasAccess {
		utils.ErrorResponse(c, http.StatusForbidden, "You do not have access to this project")
		return
	}

	// Ambil proyek beserta tim-tim yang terkait
	var project models.Project
	if err := models.DB.Preload("Teams").First(&project, uint(projectID)).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.ErrorResponse(c, http.StatusNotFound, "Project not found")
			return
		}
		utils.Logger.Errorf("Failed to retrieve project: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve project")
		return
	}

	utils.SuccessResponse(c, project.Teams)
}

// RemoveProjectTeam handles removing a team from a project
func RemoveProjectTeam(c *gin.Context) {
	// Ambil project_id dan team_id dari parameter URL
	projectIDParam := c.Param("project_id")
	projectID, err := strconv.ParseUint(projectIDParam, 10, 64)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid project ID")
		return
	}

	teamIDParam := c.Param("team_id")
	teamID, err := strconv.ParseUint(teamIDParam, 10, 64)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid team ID")
		return
	}

	// Ambil user_id dari konteks
	currentUserID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusInternalServerError, "User ID not found")
		return
	}
	userID := currentUserID.(uint)

	// Cek apakah pengguna adalah pemilik proyek
	isOwner, err := models.UserIsProjectOwner(userID, uint(projectID))
	if err != nil {
		utils.Logger.Errorf("Failed to check project ownership: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to remove team from project")
		return
	}
	if !isOwner {
		utils.ErrorResponse(c, http.StatusForbidden, "Only the project owner can remove teams from the project")
		return
	}

	// Ambil proyek dari database
	var project models.Project
	if err := models.DB.Preload("Teams").First(&project, uint(projectID)).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.ErrorResponse(c, http.StatusNotFound, "Project not found")
			return
		}
		utils.Logger.Errorf("Failed to retrieve project: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve project")
		return
	}

	// Ambil tim dari database
	var team models.Team
	if err := models.DB.First(&team, uint(teamID)).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.ErrorResponse(c, http.StatusNotFound, "Team not found")
			return
		}
		utils.Logger.Errorf("Failed to retrieve team: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve team")
		return
	}

	// Cek apakah tim terkait dengan proyek
	isAssociated := false
	for _, t := range project.Teams {
		if t.ID == team.ID {
			isAssociated = true
			break
		}
	}
	if !isAssociated {
		utils.ErrorResponse(c, http.StatusBadRequest, "Team is not associated with the project")
		return
	}

	// Hapus asosiasi antara tim dan proyek
	if err := models.DB.Model(&project).Association("Teams").Delete(&team); err != nil {
		utils.Logger.Errorf("Failed to remove team from project: %v", err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to remove team from project")
		return
	}

	utils.Logger.Infof("Team ID %d removed from Project ID %d by User ID %d", team.ID, project.ID, userID)
	utils.SuccessResponse(c, gin.H{"message": "Team removed from project successfully"})
}
