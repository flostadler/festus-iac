package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/flostadler/festus/api/pkg/db"
	"github.com/flostadler/festus/api/pkg/types"
	"github.com/gin-gonic/gin"
)

type AccountsHandler struct {
	orgDb *db.OrganizationDB
	accountDb *db.AccountDB
}

func NewAccountsHandler(orgDb *db.OrganizationDB, accountDb *db.AccountDB) *AccountsHandler {
	return &AccountsHandler{orgDb: orgDb, accountDb: accountDb}
}

func (h *AccountsHandler) CreateAccount(c *gin.Context) {
	orgName := c.Param("organizationName")
	userID := GetUserID(c)

	var acc types.Account
	if err := c.BindJSON(&acc); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := validateAccount(acc); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	org, err := h.orgDb.GetItem(userID, orgName, false)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if org == nil {
		c.JSON(http.StatusFound, gin.H{"error": "Organization does not exist"})
		return
	}

	newAcc, err := h.accountDb.PutItem(userID, orgName, &acc)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, newAcc)
}

func (h *AccountsHandler) GetAccount(c *gin.Context) {
	orgName := c.Param("organizationName")
	accountName := c.Param("accountName")
	userID := GetUserID(c)

	account, err := h.accountDb.GetItem(userID, orgName, accountName, false)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if account == nil {
		c.JSON(http.StatusFound, gin.H{"error": "Account does not exist"})
		return
	}

	c.JSON(http.StatusOK, account)
}

func (h *AccountsHandler) DeleteAccount(c *gin.Context) {
	orgName := c.Param("organizationName")
	accountName := c.Param("accountName")
	userID := GetUserID(c)

	err := h.accountDb.DeleteItem(userID, orgName, accountName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

func validateAccount(account types.Account) error {
	if strings.Contains(account.AccountName, "#") {
		return fmt.Errorf("account name contains illegal characters")
	}

	return nil
}
