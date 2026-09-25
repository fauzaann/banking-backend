package router

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"banking/config"
	"banking/controller"
	"banking/handler"
	"banking/middleware"
	"banking/models"
	jwtpkg "banking/pkg/jwt"
	"banking/pkg/validator"
	"banking/repository"
)

// Setup melakukan dependency injection dan mendaftarkan seluruh route.
func Setup(cfg *config.Config, db *gorm.DB) *gin.Engine {
	if cfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

	// --- Layer paling bawah: repository ---
	userRepo := repository.NewUserRepository(db)
	accountRepo := repository.NewAccountRepository(db)
	trxRepo := repository.NewTransactionRepository(db)
	transferRepo := repository.NewTransferRepository(db)
	benefRepo := repository.NewBeneficiaryRepository(db)
	auditRepo := repository.NewAuditRepository(db)

	// --- Utilitas ---
	jwtManager := jwtpkg.NewManager(cfg.JWT.Secret, cfg.JWT.AccessTTL(), cfg.JWT.RefreshTTL())
	blacklist := jwtpkg.NewBlacklist()
	v := validator.New()

	// --- Layer business logic: controller ---
	authCtrl := controller.NewAuthController(userRepo, accountRepo, auditRepo, jwtManager, blacklist)
	userCtrl := controller.NewUserController(userRepo, auditRepo)
	accountCtrl := controller.NewAccountController(accountRepo)
	transferCtrl := controller.NewTransferController(db, accountRepo, transferRepo, trxRepo, auditRepo)
	trxCtrl := controller.NewTransactionController(trxRepo, accountRepo)
	benefCtrl := controller.NewBeneficiaryController(benefRepo, accountRepo, userRepo, auditRepo)
	adminCtrl := controller.NewAdminController(userRepo, accountRepo, trxRepo, auditRepo)

	// --- Layer HTTP: handler ---
	authHandler := handler.NewAuthHandler(authCtrl, v)
	userHandler := handler.NewUserHandler(userCtrl, v)
	accountHandler := handler.NewAccountHandler(accountCtrl)
	transferHandler := handler.NewTransferHandler(transferCtrl, v)
	trxHandler := handler.NewTransactionHandler(trxCtrl)
	benefHandler := handler.NewBeneficiaryHandler(benefCtrl, v)
	adminHandler := handler.NewAdminHandler(adminCtrl)

	r := gin.New()
	r.Use(middleware.Recovery(), middleware.Logger())

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true, "message": "Service berjalan normal"})
	})

	// Dokumentasi API
	r.StaticFile("/swagger/swagger.yaml", "./docs/swagger.yaml")
	r.GET("/swagger/index.html", func(c *gin.Context) {
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(swaggerUIPage))
	})

	authRequired := middleware.Auth(jwtManager, userRepo, blacklist)
	adminOnly := middleware.RequireRole(models.RoleAdmin)

	api := r.Group("/api/v1")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/register", middleware.RateLimit(5, time.Minute), authHandler.Register)
			auth.POST("/login", middleware.RateLimit(10, time.Minute), authHandler.Login)
			auth.POST("/refresh", middleware.RateLimit(20, time.Minute), authHandler.Refresh)
			auth.POST("/logout", authRequired, authHandler.Logout)
		}

		users := api.Group("/users", authRequired)
		{
			users.GET("/profile", userHandler.GetProfile)
			users.PUT("/profile", userHandler.UpdateProfile)
			users.PUT("/password", userHandler.ChangePassword)
		}

		accounts := api.Group("/accounts", authRequired)
		{
			accounts.GET("", accountHandler.List)
			accounts.GET("/:id", accountHandler.Detail)
			accounts.GET("/:id/balance", accountHandler.Balance)
		}

		transfers := api.Group("/transfers", authRequired)
		{
			transfers.POST("", middleware.RateLimit(20, time.Minute), transferHandler.Create)
			transfers.GET("", transferHandler.List)
			transfers.GET("/:id", transferHandler.Detail)
		}

		transactions := api.Group("/transactions", authRequired)
		{
			transactions.GET("", trxHandler.List)
			transactions.GET("/:id", trxHandler.Detail)
		}

		beneficiaries := api.Group("/beneficiaries", authRequired)
		{
			beneficiaries.POST("", benefHandler.Create)
			beneficiaries.GET("", benefHandler.List)
			beneficiaries.GET("/:id", benefHandler.Detail)
			beneficiaries.DELETE("/:id", benefHandler.Delete)
		}

		admin := api.Group("/admin", authRequired, adminOnly)
		{
			admin.GET("/dashboard", adminHandler.Dashboard)

			admin.GET("/users", adminHandler.ListUsers)
			admin.GET("/users/:id", adminHandler.GetUser)
			admin.PUT("/users/:id/block", adminHandler.BlockUser)
			admin.PUT("/users/:id/activate", adminHandler.ActivateUser)

			admin.GET("/accounts", adminHandler.ListAccounts)
			admin.GET("/accounts/:id", adminHandler.GetAccount)
			admin.PUT("/accounts/:id/freeze", adminHandler.FreezeAccount)
			admin.PUT("/accounts/:id/activate", adminHandler.ActivateAccount)

			admin.GET("/transactions", trxHandler.AdminList)
			admin.GET("/audit-logs", adminHandler.ListAuditLogs)
		}
	}

	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false, "message": "Endpoint tidak ditemukan", "error": "NOT_FOUND",
		})
	})

	return r
}

// swaggerUIPage menyajikan Swagger UI dari CDN dan membaca /swagger/swagger.yaml.
const swaggerUIPage = `<!DOCTYPE html>
<html lang="id">
<head>
  <meta charset="utf-8" />
  <title>Digital Banking API - Swagger</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css" />
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
  <script>
    window.onload = function () {
      window.ui = SwaggerUIBundle({ url: "/swagger/swagger.yaml", dom_id: "#swagger-ui" });
    };
  </script>
</body>
</html>`
