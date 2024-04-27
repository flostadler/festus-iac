package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/flostadler/festus/api/pkg/db"
	"github.com/flostadler/festus/api/pkg/types"
	"github.com/gin-gonic/gin"
)

type OrganizationHandler struct {
	db *db.OrganizationDB
}

func NewOrganizationHandler(db *db.OrganizationDB) *OrganizationHandler {
	return &OrganizationHandler{db: db}
}

func (h *OrganizationHandler) CreateOrganization(c *gin.Context) {
	var org types.Organization
	if err := c.BindJSON(&org); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := validateOrg(org); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := GetUserID(c)

	newOrg, err := h.db.PutItem(userID, &org)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store organization", "details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, *newOrg)
}

func (h *OrganizationHandler) GetOrganization(c *gin.Context) {
	name := c.Param("organizationName")
	userID := GetUserID(c)

	org, err := h.db.GetItem(userID, name, false)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get organization", "details": err.Error()})
		return
	}

	if org == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Organization not found"})
		return
	}

	c.JSON(http.StatusOK, *org)
}

// DeleteHandler deletes an organization
func (h *OrganizationHandler) DeleteOrganization(c *gin.Context) {
	name := c.Param("organizationName")
	userID := GetUserID(c)

	err := h.db.DeleteItem(userID, name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete organization", "details": err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

func (h *OrganizationHandler) UpdateOrganization(c *gin.Context) {}

func validateOrg(org types.Organization) error {
	if strings.Contains(org.OrgName, "#") {
		return fmt.Errorf("organization name contains illegal characters")
	}
	return nil
}
