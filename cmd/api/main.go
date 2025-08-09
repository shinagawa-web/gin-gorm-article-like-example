package main

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Article struct {
	ID        int64     `gorm:"primaryKey;autoIncrement"`
	AuthorID  int64     `gorm:"not null;index"`
	Title     string    `gorm:"size:255;not null"`
	Body      string    `gorm:"type:text;not null"`
	LikeCount int64     `gorm:"not null;default:0;index"`
	CreatedAt time.Time `gorm:"not null;index"`
	UpdatedAt time.Time `gorm:"not null;index"`
}

type Like struct {
	UserID    int64 `gorm:"primaryKey"`
	ArticleID int64 `gorm:"primaryKey;index"`
}

func main() {
	dsn := "user:password@tcp(localhost:3306)/article_like?parseTime=true"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal(err)
	}
	defer sqlDB.Close()

	if err := db.AutoMigrate(&Article{}, &Like{}); err != nil {
		log.Fatal(err)
	}

	r := gin.Default()
	r.GET("/healthz", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	r.GET("/articles", func(c *gin.Context) {
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
		offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
		sort := c.DefaultQuery("sort", "new")

		if limit <= 0 || limit > 100 {
			limit = 20
		}
		if offset < 0 {
			offset = 0
		}

		q := db.Model(&Article{})
		switch sort {
		case "popular":
			q = q.Order("like_count DESC").Order("id DESC")
		case "new":
			q = q.Order("created_at DESC").Order("id DESC")
		default:
			q = q.Order("created_at DESC").Order("id DESC")
		}

		var articles []Article
		if err := q.Limit(limit).Offset(offset).Find(&articles).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"items":      articles,
			"nextOffset": offset + len(articles),
		})
	})

	r.GET("/articles/:id", func(c *gin.Context) {
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}

		var a Article
		if err := db.First(&a, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"id":        a.ID,
			"authorId":  a.AuthorID,
			"title":     a.Title,
			"body":      a.Body,
			"likeCount": a.LikeCount,
			"createdAt": a.CreatedAt,
			"updatedAt": a.UpdatedAt,
		})
	})

	r.POST("/articles", func(c *gin.Context) {
		var req struct {
			AuthorID int64  `json:"authorId" binding:"required"`
			Title    string `json:"title" binding:"required"`
			Body     string `json:"body" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}

		a := &Article{
			AuthorID:  req.AuthorID,
			Title:     req.Title,
			Body:      req.Body,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		if err := db.Create(a).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"id": a.ID})
	})

	r.PUT("/articles/:id", func(c *gin.Context) {
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}

		var req struct {
			Title string `json:"title" binding:"required"`
			Body  string `json:"body" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}

		var a Article
		if err := db.First(&a, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		a.Title = req.Title
		a.Body = req.Body
		a.UpdatedAt = time.Now()

		if err := db.Save(&a).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update"})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"id": a.ID, "authorId": a.AuthorID, "title": a.Title, "body": a.Body,
			"likeCount": a.LikeCount, "createdAt": a.CreatedAt, "updatedAt": a.UpdatedAt,
		})
	})

	r.DELETE("/articles/:id", func(c *gin.Context) {
		id, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}

		res := db.Delete(&Article{}, id)
		if res.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete"})
			return
		}
		if res.RowsAffected == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.Status(http.StatusNoContent)
	})

	r.POST("/articles/:id/like", func(c *gin.Context) {
		articleID, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}

		userID, err := strconv.ParseInt(c.DefaultQuery("userId", "1"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid userId"})
			return
		}

		if err := db.Transaction(func(tx *gorm.DB) error {
			res := tx.Clauses(clause.OnConflict{DoNothing: true}).
				Create(&Like{UserID: userID, ArticleID: articleID})
			if res.Error != nil {
				return res.Error
			}

			if res.RowsAffected > 0 {
				if err := tx.Model(&Article{}).
					Where("id = ?", articleID).
					Update("like_count", gorm.Expr("like_count + ?", 1)).Error; err != nil {
					return err
				}
			}
			return nil
		}); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to like"})
			return
		}
		c.Status(http.StatusNoContent)
	})

	r.DELETE("/articles/:id/like", func(c *gin.Context) {
		articleID, err := strconv.ParseInt(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}

		userID, err := strconv.ParseInt(c.DefaultQuery("userId", "1"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid userId"})
			return
		}

		if err := db.Transaction(func(tx *gorm.DB) error {
			res := tx.Where("user_id = ? AND article_id = ?", userID, articleID).
				Delete(&Like{})
			if res.Error != nil {
				return res.Error
			}

			if res.RowsAffected > 0 {
				if err := tx.Model(&Article{}).
					Where("id = ?", articleID).
					Update("like_count", gorm.Expr("GREATEST(like_count - ?, 0)", 1)).Error; err != nil {
					return err
				}
			}
			return nil
		}); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to unlike"})
			return
		}
		c.Status(http.StatusNoContent)
	})

	log.Println("listening on :8080")
	r.Run(":8080")
}
