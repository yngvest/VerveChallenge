package server

import (
	"database/sql"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

type Promotion struct {
	Uuid   string    `json:"id"`
	Price  float64   `json:"price"`
	Expire time.Time `json:"expiration_date"`
}

func (s *Server) GetPromotion(c *gin.Context) {
	id := c.Param("id")
	row := s.db.QueryRow("SELECT uuid, price, expire FROM promotions WHERE uuid=toUUID($1)", id)

	var promo Promotion

	err := row.Scan(&promo.Uuid, &promo.Price, &promo.Expire)
	if err != nil && err != sql.ErrNoRows {
		s.log.Printf("error making query: %s", err)
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	if err == nil {
		c.JSON(http.StatusOK, &promo)
	} else {
		c.Status(http.StatusNotFound)
	}
}
