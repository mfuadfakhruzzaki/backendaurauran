// controllers/team_controller.go
package controllers

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/mfuadfakhruzzaki/backendaurauran/models"
	"gorm.io/gorm"
)

// Initialize validator
var validate = validator.New()

// TeamInput defines the input structure for creating a team
type TeamInput struct {
    Name        string `json:"name" binding:"required"`
    Description string `json:"description"`
}

// TeamUpdateInput defines the input structure for updating a team
type TeamUpdateInput struct {
    Name        string `json:"name" binding:"required"`
    Description string `json:"description"`
}

// CreateTeam handles POST /teams/
func CreateTeam(c *gin.Context) {
    var input TeamInput

    // Bind JSON input to TeamInput struct
    if err := c.ShouldBindJSON(&input); err != nil {
        log.Printf("Error binding JSON: %v", err)
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // Validate input
    if err := validate.Struct(&input); err != nil {
        log.Printf("Validation error: %v", err)
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // Retrieve User object from context
    userInterface, exists := c.Get("user")
    if !exists {
        log.Printf("User not found in context")
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
        return
    }

    user, ok := userInterface.(models.User)
    if !ok {
        log.Printf("User type assertion failed")
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
        return
    }

    // Create a new Team instance
    newTeam := models.Team{
        Name:        input.Name,
        Description: input.Description,
        OwnerID:     user.ID,
        Owner:       user, // Ensure user has Username, Email, Role
    }

    // Save team to the database
    if err := models.DB.Create(&newTeam).Error; err != nil {
        log.Printf("Error creating team: %v", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create team"})
        return
    }

    // Preload Owner to include in the response
    if err := models.DB.Preload("Owner").First(&newTeam, newTeam.ID).Error; err != nil {
        log.Printf("Error preloading team: %v", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve created team"})
        return
    }

    c.JSON(http.StatusCreated, newTeam)
}

// ListTeams handles GET /teams/
func ListTeams(c *gin.Context) {
    var teams []models.Team

    // Retrieve all teams from the database and preload Owner and Members
    if err := models.DB.Preload("Owner").Preload("Members").Find(&teams).Error; err != nil {
        log.Printf("Error listing teams: %v", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, teams)
}

// GetTeam handles GET /teams/:team_id
func GetTeam(c *gin.Context) {
    teamID := c.Param("team_id")
    var team models.Team

    // Find team by ID and preload Owner and Members
    if err := models.DB.Preload("Owner").Preload("Members").First(&team, teamID).Error; err != nil {
        if err == gorm.ErrRecordNotFound {
            c.JSON(http.StatusNotFound, gin.H{"error": "Team not found"})
        } else {
            log.Printf("Error retrieving team: %v", err)
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        }
        return
    }

    c.JSON(http.StatusOK, team)
}

// UpdateTeam handles PUT /teams/:team_id
func UpdateTeam(c *gin.Context) {
    teamID := c.Param("team_id")
    var team models.Team

    // Find team by ID
    if err := models.DB.First(&team, teamID).Error; err != nil {
        if err == gorm.ErrRecordNotFound {
            c.JSON(http.StatusNotFound, gin.H{"error": "Team not found"})
        } else {
            log.Printf("Error finding team: %v", err)
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        }
        return
    }

    var input TeamUpdateInput

    // Bind JSON input to TeamUpdateInput struct
    if err := c.ShouldBindJSON(&input); err != nil {
        log.Printf("Error binding JSON: %v", err)
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // Validate input
    if err := validate.Struct(&input); err != nil {
        log.Printf("Validation error: %v", err)
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // Update team fields
    team.Name = input.Name
    team.Description = input.Description

    // Save changes to the database
    if err := models.DB.Save(&team).Error; err != nil {
        log.Printf("Error updating team: %v", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update team"})
        return
    }

    // Preload Owner to include in the response
    if err := models.DB.Preload("Owner").First(&team, team.ID).Error; err != nil {
        log.Printf("Error preloading team: %v", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve updated team"})
        return
    }

    c.JSON(http.StatusOK, team)
}

// DeleteTeam handles DELETE /teams/:team_id
func DeleteTeam(c *gin.Context) {
    teamID := c.Param("team_id")
    var team models.Team

    // Find team by ID
    if err := models.DB.First(&team, teamID).Error; err != nil {
        if err == gorm.ErrRecordNotFound {
            c.JSON(http.StatusNotFound, gin.H{"error": "Team not found"})
        } else {
            log.Printf("Error finding team: %v", err)
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        }
        return
    }

    // Delete team from the database
    if err := models.DB.Delete(&team).Error; err != nil {
        log.Printf("Error deleting team: %v", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "Team deleted successfully"})
}

// AddTeamMember handles POST /teams/:team_id/members/
func AddTeamMember(c *gin.Context) {
    teamID := c.Param("team_id")
    var team models.Team

    // Find team by ID and preload Owner
    if err := models.DB.Preload("Owner").First(&team, teamID).Error; err != nil {
        if err == gorm.ErrRecordNotFound {
            c.JSON(http.StatusNotFound, gin.H{"error": "Team not found"})
        } else {
            log.Printf("Error finding team: %v", err)
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        }
        return
    }

    // Retrieve User object from context
    userInterface, exists := c.Get("user")
    if !exists {
        log.Printf("User not found in context")
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
        return
    }

    user, ok := userInterface.(models.User)
    if !ok {
        log.Printf("User type assertion failed")
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
        return
    }

    // Authorization: Only team owner or admin can add members
    if team.OwnerID != user.ID && user.Role != "admin" {
        c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
        return
    }

    var input struct {
        UserID uint `json:"user_id" binding:"required"`
    }

    // Bind JSON input to struct input
    if err := c.ShouldBindJSON(&input); err != nil {
        log.Printf("Error binding JSON: %v", err)
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    var member models.User

    // Find user by ID
    if err := models.DB.First(&member, input.UserID).Error; err != nil {
        if err == gorm.ErrRecordNotFound {
            c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
        } else {
            log.Printf("Error finding user: %v", err)
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        }
        return
    }

    // Prevent adding the same user multiple times
    var existingMembers []models.User
    if err := models.DB.Model(&team).Association("Members").Find(&existingMembers); err != nil {
        log.Printf("Error checking existing members: %v", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check existing members"})
        return
    }

    for _, existing := range existingMembers {
        if existing.ID == member.ID {
            c.JSON(http.StatusBadRequest, gin.H{"error": "User is already a member of the team"})
            return
        }
    }

    // Add user to team members
    if err := models.DB.Model(&team).Association("Members").Append(&member); err != nil {
        log.Printf("Error adding member to team: %v", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "User added to team successfully"})
}

// ListTeamMembers handles GET /teams/:team_id/members/
func ListTeamMembers(c *gin.Context) {
    teamID := c.Param("team_id")
    var team models.Team

    // Find team and preload Members
    if err := models.DB.Preload("Members").First(&team, teamID).Error; err != nil {
        if err == gorm.ErrRecordNotFound {
            c.JSON(http.StatusNotFound, gin.H{"error": "Team not found"})
        } else {
            log.Printf("Error retrieving team members: %v", err)
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        }
        return
    }

    c.JSON(http.StatusOK, team.Members)
}

// RemoveTeamMember handles DELETE /teams/:team_id/members/:user_id
func RemoveTeamMember(c *gin.Context) {
    teamID := c.Param("team_id")
    userID := c.Param("user_id")

    var team models.Team
    var user models.User

    // Find team by ID
    if err := models.DB.First(&team, teamID).Error; err != nil {
        if err == gorm.ErrRecordNotFound {
            c.JSON(http.StatusNotFound, gin.H{"error": "Team not found"})
        } else {
            log.Printf("Error finding team: %v", err)
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        }
        return
    }

    // Retrieve User object from context
    authUserInterface, exists := c.Get("user")
    if !exists {
        log.Printf("User not found in context")
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
        return
    }

    authUser, ok := authUserInterface.(models.User)
    if !ok {
        log.Printf("User type assertion failed")
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
        return
    }

    // Authorization: Only team owner or admin can remove members
    if team.OwnerID != authUser.ID && authUser.Role != "admin" {
        c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
        return
    }

    // Find user by ID
    if err := models.DB.First(&user, userID).Error; err != nil {
        if err == gorm.ErrRecordNotFound {
            c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
        } else {
            log.Printf("Error finding user: %v", err)
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        }
        return
    }

    // Remove user from team members
    if err := models.DB.Model(&team).Association("Members").Delete(&user); err != nil {
        log.Printf("Error removing member from team: %v", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "User removed from team successfully"})
}
