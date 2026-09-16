package routes

import (
	"net/http"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

const noopUserAPIKeyPrefix = "noop-user:"

type noopProvisionAPIKeyRequest struct {
	Name string `json:"name"`
}

// RegisterNoopInternalRoutes registers the private BFF-to-relay provisioning
// endpoint. The relay is deployed on a private network; the shared Noop key
// authenticates the request and the resulting key is returned only to Noop.
func RegisterNoopInternalRoutes(
	r *gin.Engine,
	apiKeyAuth middleware.APIKeyAuthMiddleware,
	apiKeyService *service.APIKeyService,
) {
	r.POST(
		"/internal/noop/provision-api-key",
		gin.HandlerFunc(apiKeyAuth),
		func(c *gin.Context) {
			subject, ok := middleware.GetAuthSubjectFromContext(c)
			if !ok {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
				return
			}

			var req noopProvisionAPIKeyRequest
			if err := c.ShouldBindJSON(&req); err != nil || !validNoopUserAPIKeyName(req.Name) {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid Noop user API key name"})
				return
			}

			groupID := inheritedAPIKeyGroupID(c)
			key, err := findNoopUserAPIKey(c, apiKeyService, subject.UserID, req.Name)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to find Noop user API key"})
				return
			}
			if key == nil {
				key, err = apiKeyService.Create(c.Request.Context(), subject.UserID, service.CreateAPIKeyRequest{
					Name:    req.Name,
					GroupID: groupID,
				})
				if err != nil {
					// A concurrent request can create the same named key between
					// search and create. Return that key instead of minting another.
					key, err = findNoopUserAPIKey(c, apiKeyService, subject.UserID, req.Name)
					if err != nil || key == nil {
						c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to provision Noop user API key"})
						return
					}
				}
			}

			c.JSON(http.StatusOK, gin.H{
				"id":   key.ID,
				"key":  key.Key,
				"name": key.Name,
			})
		},
	)
}

func validNoopUserAPIKeyName(name string) bool {
	return strings.HasPrefix(name, noopUserAPIKeyPrefix) &&
		len(name) <= 100 &&
		strings.TrimSpace(name) == name &&
		!strings.ContainsAny(name, "\r\n")
}

func inheritedAPIKeyGroupID(c *gin.Context) *int64 {
	apiKey, ok := middleware.GetAPIKeyFromContext(c)
	if !ok || apiKey == nil || apiKey.GroupID == nil {
		return nil
	}

	groupID := *apiKey.GroupID
	return &groupID
}

func findNoopUserAPIKey(
	c *gin.Context,
	apiKeyService *service.APIKeyService,
	userID int64,
	name string,
) (*service.APIKey, error) {
	keys, err := apiKeyService.SearchAPIKeys(c.Request.Context(), userID, name, 100)
	if err != nil {
		return nil, err
	}
	for i := range keys {
		if keys[i].Name == name && keys[i].IsActive() {
			return &keys[i], nil
		}
	}
	return nil, nil
}
