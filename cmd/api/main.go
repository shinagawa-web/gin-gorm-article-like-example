package main

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
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

	if err := db.AutoMigrate(&Article{}); err != nil {
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

	log.Println("listening on :8080")
	r.Run(":8080")
}
